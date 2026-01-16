package main

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func isRunningInDocker() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	data, err := os.ReadFile("/proc/1/cgroup")
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "docker") || strings.Contains(string(data), "containerd")
}

func discoverProjectsFromContainers() []Project {
	cmd := exec.Command("docker", "ps", "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("✘ Failed to list containers:", err)
		return nil
	}

	var projects []Project
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	fmt.Printf("→ Discovering containers from %d running containers\n", len(lines))

	for _, line := range lines {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}

		if strings.Contains(name, "updatectrl") {
			fmt.Printf("  ⊘ Skipping updatectrl container: %s\n", name)
			continue
		}

		composeCmd := exec.Command("docker", "inspect", "--format", "{{index .Config.Labels \"com.docker.compose.project\"}}", name)
		composeOutput, _ := composeCmd.Output()
		if strings.TrimSpace(string(composeOutput)) != "" && strings.TrimSpace(string(composeOutput)) != "<no value>" {
			fmt.Printf("  ⊘ Skipping Docker Compose container: %s (project: %s)\n", name, strings.TrimSpace(string(composeOutput)))
			continue
		}

		// Get the actual image name from inspect (handles image IDs)
		imageCmd := exec.Command("docker", "inspect", "--format", "{{.Config.Image}}", name)
		imageOutput, err := imageCmd.Output()
		if err != nil {
			fmt.Printf("  ⊘ Failed to inspect container: %s\n", name)
			continue
		}
		image := strings.TrimSpace(string(imageOutput))

		// Filter for docker.io or ghcr.io images, but also allow images without prefix
		// Docker Hub images often don't have docker.io/ prefix
		hasValidPrefix := strings.HasPrefix(image, "docker.io/") ||
			strings.HasPrefix(image, "ghcr.io/")

		// Also check if it looks like a registry image (contains / or :)
		looksLikeRegistryImage := strings.Contains(image, "/") || strings.Contains(image, ":")

		if !hasValidPrefix && !looksLikeRegistryImage {
			fmt.Printf("  ⊘ Skipping local image: %s (%s)\n", name, image)
			continue
		}

		ports := getContainerPublishedPorts(name)
		env := getContainerEnv(name)

		project := Project{
			Name:  name,
			Type:  "image",
			Image: image,
			Port:  ports,
			Env:   env,
		}
		projects = append(projects, project)
		fmt.Printf("  ✓ Discovered: %s (image: %s, ports: %s, env vars: %d)\n", name, image, ports, len(env))
	}

	fmt.Printf("→ Total containers to monitor: %d\n", len(projects))
	return projects
}

func getContainerPublishedPorts(containerName string) string {
	// Get the port bindings in a more reliable format
	cmd := exec.Command("docker", "port", containerName)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	portMap := make(map[string]bool) // Use map to deduplicate
	var portMappings []string

	for _, line := range lines {
		// Format is like: "80/tcp -> 0.0.0.0:8081" or "80/tcp -> [::]:8081"
		if strings.Contains(line, "->") {
			parts := strings.Split(line, "->")
			if len(parts) != 2 {
				continue
			}

			// Get container port (left side, e.g., "80/tcp")
			containerPort := strings.TrimSpace(parts[0])
			containerPort = strings.TrimSuffix(containerPort, "/tcp")
			containerPort = strings.TrimSuffix(containerPort, "/udp")

			// Get host binding (right side, e.g., "0.0.0.0:8081" or "[::]:8081")
			hostBinding := strings.TrimSpace(parts[1])

			hostParts := strings.Split(hostBinding, ":")
			if len(hostParts) >= 2 {
				hostPort := hostParts[len(hostParts)-1]
				portMapping := fmt.Sprintf("%s:%s", hostPort, containerPort)

				// Only add if we haven't seen this mapping before
				if !portMap[portMapping] {
					portMap[portMapping] = true
					portMappings = append(portMappings, portMapping)
				}
			}
		}
	}

	return strings.Join(portMappings, " ")
}

func getContainerEnv(containerName string) map[string]string {
	cmd := exec.Command("docker", "inspect", "--format", `{{range .Config.Env}}{{println .}}{{end}}`, containerName)
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	env := make(map[string]string)

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if kv := strings.SplitN(line, "=", 2); len(kv) == 2 {
			// Skip PATH and other system env vars that might cause issues
			key := kv[0]
			if key != "PATH" && key != "HOSTNAME" {
				env[key] = kv[1]
			}
		}
	}

	return env
}

var version = "0.1.0"

type Project struct {
	Name          string            `yaml:"name"`
	Path          string            `yaml:"path"`
	Repo          string            `yaml:"repo"`
	Type          string            `yaml:"type"`
	BuildCommand  string            `yaml:"buildCommand"`
	Image         string            `yaml:"image"`         // Docker image to pull (e.g., "ghcr.io/user/vite-app:main")
	Port          string            `yaml:"port"`          // Port mapping (e.g., "80:80" or "3000:80")
	Env           map[string]string `yaml:"env"`           // Environment variables
	ContainerName string            `yaml:"containerName"` // Optional custom container name
}

type Config struct {
	// Deprecated: Use Interval instead.
	IntervalMinutes int       `yaml:"intervalMinutes"`
	Interval        int       `yaml:"interval"`
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

func main() {
	rootCmd := &cobra.Command{
		Use:     "updatectrl",
		Version: version,
	}
	rootCmd.AddCommand(initCmd, watchCmd, buildCmd, listCmd, logsCmd)
	rootCmd.Execute()
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

func loadConfig() Config {
	if isRunningInDocker() {
		return loadConfigFromEnv()
	}

	var configPath string
	if runtime.GOOS == "windows" {
		configPath = filepath.Join(os.Getenv("USERPROFILE"), "updatectrl", "updatectrl.yaml")
	} else {
		configPath = "/etc/updatectrl/updatectrl.yaml"
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

func loadConfigFromEnv() Config {
	config := Config{}

	if intervalStr := os.Getenv("UPDATECTL_INTERVAL"); intervalStr != "" {
		if interval, err := strconv.Atoi(intervalStr); err == nil {
			config.Interval = interval
		}
	} else {
		config.Interval = 600 // default 10 minutes
	}

	// Auto-discover projects from running containers
	config.Projects = discoverProjectsFromContainers()

	return config
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

func getImageDigest(image string) (string, error) {
	cmd := exec.Command("docker", "inspect", "--format={{index .RepoDigests 0}}", image)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func pullDockerImage(image string) error {
	fmt.Println("→ Pulling Docker image:", image)
	cmd := exec.Command("docker", "pull", image)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func restartDockerContainer(p Project) error {
	containerName := p.ContainerName
	if containerName == "" {
		containerName = p.Name
	}

	// Stop and remove old container if it exists
	fmt.Println("→ Stopping old container:", containerName)
	stopCmd := exec.Command("docker", "stop", containerName)
	stopCmd.Run() // Ignore error if container doesn't exist

	rmCmd := exec.Command("docker", "rm", containerName)
	rmCmd.Run() // Ignore error if container doesn't exist

	// Build docker run command
	args := []string{"run", "-d", "--name", containerName}

	// Add port mappings if specified (can be space-separated for multiple ports)
	if p.Port != "" {
		fmt.Printf("→ Configuring ports: %s\n", p.Port)
		portMappings := strings.Fields(p.Port)
		for _, portMapping := range portMappings {
			args = append(args, "-p", portMapping)
			fmt.Printf("  - Port mapping: %s\n", portMapping)
		}
	}

	// Add environment variables
	if len(p.Env) > 0 {
		fmt.Printf("→ Configuring %d environment variables\n", len(p.Env))
	}
	for key, value := range p.Env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	// Add restart policy
	args = append(args, "--restart", "unless-stopped")

	// Add image
	args = append(args, p.Image)

	fmt.Printf("→ Starting new container: docker run %s\n", strings.Join(args, " "))
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
func getRemoteImageDigest(image string) (string, error) {
	cmd := exec.Command("docker", "manifest", "inspect", image)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, `"digest":`) {
			// Extract digest value: "digest": "sha256:xxxxx"
			parts := strings.Split(line, `"`)
			if len(parts) >= 4 {
				return parts[3], nil
			}
		}
	}
	return "", fmt.Errorf("could not parse digest from manifest")
}
func updateProject(p Project) {
	if p.Type == "image" {
		if p.Image == "" {
			fmt.Println("✘ No image specified for project:", p.Name)
			return
		}

		containerName := p.ContainerName
		if containerName == "" {
			containerName = p.Name
		}

		checkCmd := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", containerName)
		output, err := checkCmd.Output()
		containerRunning := err == nil && strings.TrimSpace(string(output)) == "true"

		// Get current local image digest
		currentDigest, err := getImageDigest(p.Image)
		if err != nil || currentDigest == "" {
			fmt.Println("→ Local image not found or no digest available")
			currentDigest = ""
		} else {
			fmt.Println("→ Current local digest:", currentDigest)
		}

		// Get remote registry digest
		remoteDigest, err := getRemoteImageDigest(p.Image)
		if err != nil {
			fmt.Println("→ Could not check remote digest:", err)
			// If we can't check remote, pull anyway to be safe
			remoteDigest = ""
		} else {
			fmt.Println("→ Remote registry digest:", remoteDigest)
		}

		// Determine if image needs update
		imageNeedsUpdate := false
		if currentDigest == "" {
			// No local image exists
			imageNeedsUpdate = true
		} else if remoteDigest == "" {
			// Couldn't check remote, assume update needed
			imageNeedsUpdate = true
		} else {
			// Extract just the sha256 hash from currentDigest if it's a full repo digest
			// currentDigest might be like "ghcr.io/user/app@sha256:abc123"
			// remoteDigest is like "sha256:abc123"
			currentHash := currentDigest
			if strings.Contains(currentDigest, "@") {
				parts := strings.Split(currentDigest, "@")
				if len(parts) == 2 {
					currentHash = parts[1]
				}
			}

			// Compare hashes
			imageNeedsUpdate = currentHash != remoteDigest
		}

		if !imageNeedsUpdate && containerRunning {
			fmt.Println("● Image already up to date and container running:", p.Name)
			return
		}

		if imageNeedsUpdate {
			fmt.Println("→ Pulling latest image:", p.Image)
			if err := pullDockerImage(p.Image); err != nil {
				fmt.Println("✘ Failed to pull image:", err)
				return
			}
			fmt.Println("✓ New image version detected:", p.Name)
		} else if !containerRunning {
			fmt.Println("→ Container not running, starting it:", p.Name)
		}

		if err := restartDockerContainer(p); err != nil {
			fmt.Println("✘ Failed to restart container:", err)
			return
		}
		fmt.Println("✓ Container started successfully")

		return
	}

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
