package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
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
