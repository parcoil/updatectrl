package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

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
			if p.Type == "image" && p.Image != "" {
				portInfo := ""
				if p.Port != "" {
					portInfo = fmt.Sprintf(", port=%s", p.Port)
				}
				fmt.Printf("- %s (%s): image=%s%s\n", p.Name, p.Type, p.Image, portInfo)
			} else {
				fmt.Printf("- %s (%s): %s\n", p.Name, p.Type, p.Path)
			}
		}
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize updatectrl configuration and daemon",
	Run: func(cmd *cobra.Command, args []string) {
		if runtime.GOOS != "windows" && os.Geteuid() != 0 {
			fmt.Println("Error: This command requires root privileges on Linux.")
			fmt.Println("Please run: sudo updatectrl init")
			os.Exit(1)
		}

		var configDir, configPath string
		if runtime.GOOS == "windows" {
			configDir = filepath.Join(os.Getenv("USERPROFILE"), "updatectrl")
		} else {
			configDir = "/etc/updatectrl"
		}
		configPath = filepath.Join(configDir, "updatectrl.yaml")

		if err := os.MkdirAll(configDir, 0755); err != nil {
			fmt.Printf("Failed to create config directory: %v\n", err)
			os.Exit(1)
		}

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			defaultConfig := []byte(`interval: 600
intervalMinutes: 10
projects:
  # Git-based project with Docker build
  - name: example-git
    path: /srv/example
    repo: https://github.com/user/example.git
    type: docker
    buildCommand: docker compose up -d --build

  # Vite app from GitHub Container Registry
  - name: vite-app
    type: image
    image: ghcr.io/user/vite-app:main
    port: "80:80"
    env:
      NODE_ENV: production
      API_URL: https://api.example.com

  # React app from Docker Hub
  - name: react-dashboard
    type: image
    image: user/react-dashboard:latest
    port: "3000:80"
    containerName: my-dashboard
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
			taskName := "updatectrl"
			configDir := filepath.Join(os.Getenv("USERPROFILE"), "updatectrl")
			batScript := fmt.Sprintf(`@echo off
start "" /b "%s" watch
`, filepath.Join(configDir, "updatectrl.exe"))
			batScriptPath := filepath.Join(configDir, "run_updatectrl.bat")
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
			fmt.Println("Created Windows Task Scheduler job for updatectrl.")
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
			servicePath := "/etc/systemd/system/updatectrl.service"
			service := fmt.Sprintf(`[Unit]
Description=Updatectrl Daemon - Auto-update your projects
After=network.target

[Service]
ExecStart=/usr/local/bin/updatectrl watch
WorkingDirectory=/etc/updatectrl
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

			enableCmd := exec.Command("systemctl", "enable", "--now", "updatectrl")
			if output, err := enableCmd.CombinedOutput(); err != nil {
				fmt.Printf("Failed to enable and start service: %v\nOutput: %s\n", err, output)
				os.Exit(1)
			}
			fmt.Println("Systemd service installed and started.")
			fmt.Println("\nCheck status with: sudo systemctl status updatectrl")
			fmt.Println("View logs with: sudo journalctl -u updatectrl -f")
		}
	},
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View updatectrl daemon logs",
	Long:  "View logs from the updatectrl daemon service",
	Run: func(cmd *cobra.Command, args []string) {
		follow, _ := cmd.Flags().GetBool("follow")
		lines, _ := cmd.Flags().GetInt("lines")

		if runtime.GOOS == "windows" {
			fmt.Println("Viewing Windows Task Scheduler logs...")
			fmt.Println("You can view task history in Task Scheduler GUI or use:")
			fmt.Println("  Get-WinEvent -LogName Microsoft-Windows-TaskScheduler/Operational | Where-Object {$_.Message -like '*updatectrl*'}")
			return
		}

		// Linux - use journalctl
		journalArgs := []string{"-u", "updatectrl"}

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
			fmt.Println("Try: sudo journalctl -u updatectrl -f")
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
	Short: "Run updatectrl daemon to auto-update projects",
	Run: func(cmd *cobra.Command, args []string) {
		config := loadConfig()
		var intervalSeconds int
		if config.Interval > 0 {
			intervalSeconds = config.Interval
		} else {
			intervalSeconds = config.IntervalMinutes * 60
		}
		fmt.Printf("Running updatectrl every %d seconds...\n", intervalSeconds)

		if isRunningInDocker() {
			fmt.Println("→ Running in Docker mode - auto-discovering containers")
		}

		for {
			// Reload config each iteration when in Docker mode to pick up new containers
			if isRunningInDocker() {
				config = loadConfig()
			}

			if len(config.Projects) == 0 {
				fmt.Println("⚠ No projects found to monitor")
			}

			for _, p := range config.Projects {
				fmt.Println("\n→ Checking", p.Name)
				updateProject(p)
			}

			fmt.Printf("\n→ Sleeping for %d seconds...\n", intervalSeconds)
			time.Sleep(time.Duration(intervalSeconds) * time.Second)
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
