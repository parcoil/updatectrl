package main

import (
	"bufio"
	"fmt"
	"strings"

	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var version = "0.1.0"

type Project struct {
	Name         string `yaml:"name"`
	Path         string `yaml:"path"`
	Repo         string `yaml:"repo"`
	Type         string `yaml:"type"`
	BuildCommand string `yaml:"buildCommand"`
}

type Config struct {
	IntervalMinutes int       `yaml:"intervalMinutes"`
	Projects        []Project `yaml:"projects"`
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured projects",
	Run: func(cmd *cobra.Command, args []string) {
		config := loadConfig()
		if len(config.Projects) == 0 {
			fmt.Println("No projects configured.")
			return
		}
		fmt.Println("Configured projects:")
		for _, p := range config.Projects {
			fmt.Printf("- %s (%s): %s\n", p.Name, p.Type, p.Path)
		}
	},
}

func main() {
	rootCmd := &cobra.Command{
		Use:     "updatectl",
		Version: version,
	}
	rootCmd.AddCommand(initCmd, watchCmd, buildCmd, listCmd,logsCmd)
	rootCmd.Execute()
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize updatectl configuration and daemon",
	Run: func(cmd *cobra.Command, args []string) {
		if runtime.GOOS != "windows" && os.Geteuid() != 0 {
			fmt.Println("Error: This command requires root privileges on Linux.")
			fmt.Println("Please run: sudo updatectl init")
			os.Exit(1)
		}

		var configDir, configPath string
		if runtime.GOOS == "windows" {
			configDir = filepath.Join(os.Getenv("USERPROFILE"), "updatectl")
		} else {
			configDir = "/etc/updatectl"
		}
		configPath = filepath.Join(configDir, "updatectl.yaml")

		if err := os.MkdirAll(configDir, 0755); err != nil {
			fmt.Printf("Failed to create config directory: %v\n", err)
			os.Exit(1)
		}

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			defaultConfig := []byte(`intervalMinutes: 10
projects:
  - name: example
    path: /srv/example
    repo: https://github.com/user/example.git
    type: docker
    buildCommand: docker compose up -d --build
`)
			if err := os.WriteFile(configPath, defaultConfig, 0644); err != nil {
				fmt.Printf("Failed to write config file: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Created config at", configPath)
		} else {
			fmt.Println("Config already exists at", configPath)
		}

		if runtime.GOOS == "windows" {
			taskName := "updatectl"
			configDir := filepath.Join(os.Getenv("USERPROFILE"), "updatectl")
			batScript := fmt.Sprintf(`@echo off
start "" /b "%s" watch
`, filepath.Join(configDir, "updatectl.exe"))
			batScriptPath := filepath.Join(configDir, "run_updatectl.bat")
			err := os.WriteFile(batScriptPath, []byte(batScript), 0644)
			if err != nil {
				fmt.Println("Failed to write batch wrapper script:", err)
				return
			}
			taskRun := batScriptPath
			createCmd := exec.Command(
				"schtasks",
				"/Create",
				"/TN", taskName,
				"/TR", taskRun,
				"/SC", "ONSTART",
				"/RL", "HIGHEST",
				"/F",
			)
			output, err := createCmd.CombinedOutput()
			if err != nil {
				fmt.Printf("Failed to create scheduled task: %v\nOutput: %s\n", err, output)
				return
			}
			fmt.Println("Created Windows Task Scheduler job for updatectl.")
			runCmd := exec.Command("schtasks", "/Run", "/TN", taskName)
			runOutput, runErr := runCmd.CombinedOutput()
			if runErr != nil {
				fmt.Printf("Failed to run scheduled task immediately: %v\nOutput: %s\n", runErr, runOutput)
			} else {
				fmt.Println("Scheduled task started immediately.")
			}
		} else {
			fmt.Print("Enter the user for the systemd service (default: root): ")
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			user := strings.TrimSpace(scanner.Text())
			if user == "" {
				user = "root"
			}
			servicePath := "/etc/systemd/system/updatectl.service"
			service := fmt.Sprintf(`[Unit]
Description=Updatectl Daemon - Auto-update your projects
After=network.target

[Service]
ExecStart=/usr/local/bin/updatectl watch
WorkingDirectory=/etc/updatectl
Restart=always
User=%s

[Install]
WantedBy=multi-user.target
`, user)
			if err := os.WriteFile(servicePath, []byte(service), 0644); err != nil {
				fmt.Printf("Failed to write systemd service file: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Created systemd service file at", servicePath)

			reloadCmd := exec.Command("systemctl", "daemon-reload")
			if err := reloadCmd.Run(); err != nil {
				fmt.Printf("Failed to reload systemd daemon: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Reloaded systemd daemon")

			enableCmd := exec.Command("systemctl", "enable", "--now", "updatectl")
			if output, err := enableCmd.CombinedOutput(); err != nil {
				fmt.Printf("Failed to enable and start service: %v\nOutput: %s\n", err, output)
				os.Exit(1)
			}
			fmt.Println("Systemd service installed and started.")
			fmt.Println("\nCheck status with: sudo systemctl status updatectl")
			fmt.Println("View logs with: sudo journalctl -u updatectl -f")
		}
	},
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View updatectl daemon logs",
	Long:  "View logs from the updatectl daemon service",
	Run: func(cmd *cobra.Command, args []string) {
		follow, _ := cmd.Flags().GetBool("follow")
		lines, _ := cmd.Flags().GetInt("lines")

		if runtime.GOOS == "windows" {
			fmt.Println("Viewing Windows Task Scheduler logs...")
			fmt.Println("You can view task history in Task Scheduler GUI or use:")
			fmt.Println("  Get-WinEvent -LogName Microsoft-Windows-TaskScheduler/Operational | Where-Object {$_.Message -like '*updatectl*'}")
			return
		}

		// Linux - use journalctl
		journalArgs := []string{"-u", "updatectl"}
		
		if follow {
			journalArgs = append(journalArgs, "-f")
		}
		
		if lines > 0 {
			journalArgs = append(journalArgs, "-n", fmt.Sprintf("%d", lines))
		}

		journalCmd := exec.Command("journalctl", journalArgs...)
		journalCmd.Stdout = os.Stdout
		journalCmd.Stderr = os.Stderr
		journalCmd.Stdin = os.Stdin

		if err := journalCmd.Run(); err != nil {
			fmt.Printf("Failed to view logs: %v\n", err)
			fmt.Println("Try: sudo journalctl -u updatectl -f")
			os.Exit(1)
		}
	},
}

func init() {
	logsCmd.Flags().BoolP("follow", "f", false, "Follow log output (live tail)")
	logsCmd.Flags().IntP("lines", "n", 50, "Number of log lines to show")
}


var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Run updatectl daemon to auto-update projects",
	Run: func(cmd *cobra.Command, args []string) {
		config := loadConfig()
		fmt.Printf("Running updatectl every %d minutes...\n", config.IntervalMinutes)

		for {
			for _, p := range config.Projects {
				fmt.Println("Checking", p.Name)
				updateProject(p)
			}
			time.Sleep(time.Duration(config.IntervalMinutes) * time.Minute)
		}
	},
}

var buildCmd = &cobra.Command{
	Use:   "build [project-name]",
	Short: "Run build command for a specific project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		config := loadConfig()

		for _, p := range config.Projects {
			if p.Name == projectName {
				if p.BuildCommand == "" {
					fmt.Printf("No build command configured for project %s\n", projectName)
					return
				}

				fmt.Printf("Building project %s...\n", projectName)
				err := runBuildCommand(p.BuildCommand, p.Path)
				if err != nil {
					fmt.Printf("Build failed for %s: %v\n", projectName, err)
				} else {
					fmt.Printf("Build completed for %s\n", projectName)
				}
				return
			}
		}
		fmt.Printf("Project %s not found in configuration\n", projectName)
	},
}

func loadConfig() Config {
	var configPath string
	if runtime.GOOS == "windows" {
		configPath = filepath.Join(os.Getenv("USERPROFILE"), "updatectl", "updatectl.yaml")
	} else {
		configPath = "/etc/updatectl/updatectl.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Println("Failed to read config:", err)
		os.Exit(1)
	}

	var c Config
	yaml.Unmarshal(data, &c)
	return c
}

func runBuildCommand(command, dir string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("bash", "-c", command)
	}
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func updateProject(p Project) {
	if _, err := os.Stat(p.Path); os.IsNotExist(err) {
		fmt.Println("✘ Path not found:", p.Path)
		return
	}

	fmt.Println("→ Pulling latest changes for", p.Name)
	gitPull := exec.Command("git", "-C", p.Path, "pull")
	output, err := gitPull.CombinedOutput()
	if err != nil {
		fmt.Println("✘ Git pull failed:", err)
		return
	}
	fmt.Print(string(output))

	if strings.Contains(string(output), "Already up to date.") {
		fmt.Println("● No new commits for", p.Name)
		return
	}

	if p.BuildCommand != "" {
		fmt.Println("→ Running build command for", p.Name)
		runBuildCommand(p.BuildCommand, p.Path)
	}

	switch p.Type {
	case "pm2":
		fmt.Println("→ Restarting PM2 process:", p.Name)
		cmd := exec.Command("pm2", "restart", p.Name)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	case "docker":
		// Build command already run above
	case "static":
		// No additional action needed
	default:
		fmt.Println("Unknown type:", p.Type)
	}
}
