# updatectrl

A CLI tool for automating project updates. It periodically pulls the latest changes from Git repositories or checks for new Docker images and rebuilds/restarts projects based on their type (PM2, Docker, or Image).

> [!TIP]
> For the best experience, run updatectrl inside Docker. It can automatically discover and manage all your running containers.

> [!WARNING]
> This project is very barebones and a work in progress and not ready for production use.

## Installation

### Using Installers

Installer scripts are provided in the repository for easy installation.

- **Linux**: Run `./install.sh` (requires sudo for installation)
- **Windows**: Run `install.bat` as administrator

### Manual Installation

1. Clone or download the repository.
2. Build the executable: `go build -o updatectrl main.go`
3. Move `updatectrl` to a directory in your PATH (e.g., `/usr/local/bin/` on Linux or `C:\Program Files\updatectrl\` on Windows).

or build it and upload to your server

## Usage

### Initialize Configuration

Run `updatectrl init` to create the default configuration file and set up the daemon.

- On Linux: Creates a systemd service.
- On Windows: Creates a Task Scheduler job.

### Start Watching

The daemon runs automatically after init. To run manually: `updatectrl watch`

### List Projects

List configured projects: `updatectrl list`

### View Logs

View daemon logs: `updatectrl logs`

- Use `updatectrl logs -f` to follow logs in real-time
- Use `updatectrl logs -n 100` to show the last 100 lines

## Configuration

Configuration is stored in:

- Linux: `/etc/updatectrl/updatectrl.yaml`
- Windows: `%USERPROFILE%\updatectrl\updatectrl.yaml`

Example config:

```yaml
interval: 600  # Check every 10 minutes (in seconds)
projects:
  # Git-based Docker project
  - name: myapp
    path: /srv/myapp
    repo: https://github.com/user/myapp.git
    type: docker
    buildCommand: docker compose up -d --build
  # Git-based PM2 project
  - name: webserver
    path: /srv/webserver
    repo: https://github.com/user/webserver.git
    type: pm2
    buildCommand: "" # Not used for PM2
  # Image-based project
  - name: dashboard
    type: image
    image: ghcr.io/user/dashboard:latest
    port: "3000:80"
    env:
      NODE_ENV: production
      API_URL: https://api.example.com
    containerName: my-dashboard
```

- `interval`: How often to check for updates (in seconds). Recommended over `intervalMinutes`.
- `intervalMinutes`: Deprecated. Use `interval` instead.
- `projects`: List of projects to monitor.
  - `name`: Project name.
  - `path`: Local path to the project (required for git-based types).
  - `repo`: Git repository URL (required for git-based types).
  - `type`: "pm2", "docker", "static", or "image".
  - `buildCommand`: Command to run after pulling (for git-based types).
  - `image`: Docker image to pull (required for image type).
  - `port`: Port mapping (optional for image type, e.g., "80:80").
  - `env`: Environment variables (optional for image type).
  - `containerName`: Custom container name (optional for image type, defaults to project name).

## Supported Project Types

- **PM2**: Restarts the PM2 process with the project name.
- **Docker**: Runs the specified build command (e.g., Docker Compose rebuild or direct Docker commands).
- **Static**: Runs the build command after git pull (for static sites).
- **Image**: Pulls the latest Docker image and restarts the container if the image has been updated.

### Docker Without Compose

For projects using Docker without Compose, set the `buildCommand` to build and run the container directly. Example:

```yaml
projects:
  - name: myapp
    path: /srv/myapp
    repo: https://github.com/user/myapp.git
    type: docker
    buildCommand: docker build -t myapp . && docker stop myapp || true && docker rm myapp || true && docker run -d --name myapp -p 8080:8080 myapp
```

### Image-based Projects

For applications distributed as pre-built Docker images (e.g., from GitHub Container Registry or Docker Hub). The tool will automatically pull new versions and restart containers.

```yaml
projects:
  - name: webapp
    type: image
    image: ghcr.io/company/webapp:v1.2.3
    port: "80:3000"
    env:
      DATABASE_URL: postgres://localhost/mydb
    containerName: production-webapp
```

## Requirements

- Go
- Git
- PM2 (for PM2 projects)
- Docker (for Docker and Image projects)
