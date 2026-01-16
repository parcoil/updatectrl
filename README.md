# updatectrl

A CLI tool for automating project updates. It periodically pulls the latest changes from Git repositories or checks for new Docker images and rebuilds/restarts projects based on their type (Static, PM2, Docker, or Image).

> [!TIP]
> Build from source for the most reliable installation - it's quick and ensures compatibility with your system.

> [!WARNING]
> This project is very barebones and a work in progress and not ready for production use.

## Installation

### Build from Source (Recommended)

1. Ensure Go is installed on your system
2. Clone the repository: `git clone https://github.com/yourusername/updatectrl.git`
3. Navigate to the directory: `cd updatectrl`
4. Build the binary: `go build -o updatectrl main.go`
5. Move to PATH: `sudo mv updatectrl /usr/local/bin/` (Linux/Mac) or move to a directory in your PATH (Windows)

### Using Installers

For automated installation with scripts:

- **Linux**: Run `./install.sh` (requires sudo)
- **Windows**: Run `install.bat` as administrator

### Build and Upload

Build locally and upload the binary to your server:

```bash
go build -o updatectrl main.go
# Then upload updatectrl to your server and place in PATH
```

## Docker Image

### Building Multi-Platform Images

Build images for Linux AMD64 and ARM64 architectures:

**Linux/macOS:**
```bash
./build-docker.sh
```

**Windows:**
```cmd
build-docker.bat
```

This builds and pushes the image to `ghcr.io/parcoil/updatectrl:latest` with multi-platform support.

### Automated Builds

The repository includes a GitHub Actions workflow (`.github/workflows/docker-build.yml`) that automatically builds and pushes multi-platform images to GitHub Container Registry on every push to main/master and on tagged releases.

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
  # Git-based Static project (e.g., static site generator)
  - name: mysite
    path: /srv/mysite
    repo: https://github.com/user/mysite.git
    type: static
    buildCommand: npm run build
  # Git-based PM2 project
  - name: webserver
    path: /srv/webserver
    repo: https://github.com/user/webserver.git
    type: pm2
    buildCommand: npm install && npm run build
  # Git-based Docker project
  - name: myapp
    path: /srv/myapp
    repo: https://github.com/user/myapp.git
    type: docker
    buildCommand: docker compose up -d --build
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
  - `type`: "pm2", "docker", "static", or "image" (use "docker" for projects with Dockerfiles/Compose files, "image" for pre-built registry images).
  - `buildCommand`: Command to run after pulling (for git-based types).
  - `image`: Docker image to pull (required for image type).
  - `port`: Port mapping (optional for image type, e.g., "80:80").
  - `env`: Environment variables (optional for image type).
  - `containerName`: Custom container name (optional for image type, defaults to project name).

## Supported Project Types

- **Static**: Runs the build command after git pull (for static sites and projects).
- **PM2**: Restarts the PM2 process with the project name after git pull and optional build command.
- **Docker**: Runs the specified build command after git pull (for projects with Dockerfiles or Docker Compose - use `docker compose up -d --build`).
- **Image**: Pulls the latest Docker image and restarts the container if the image has been updated (for pre-built registry images).

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
