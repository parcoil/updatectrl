package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var version = "0.1.0"

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
