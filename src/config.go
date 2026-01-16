package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"gopkg.in/yaml.v3"
)

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
