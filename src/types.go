package main

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
