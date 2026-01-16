# Quick Start Guide

This guide will get you up and running with Updatectrl in minutes.

## Prerequisites

- Go installed
- Git installed
- PM2 (for PM2 projects)
- Docker (for Docker and Image projects)

## Installation

### Build from Source (Recommended)

```bash
git clone https://github.com/yourusername/updatectrl.git
cd updatectrl
go build -o updatectrl main.go
sudo mv updatectrl /usr/local/bin/
```

### Using Installers

- **Linux**: `./install.sh` (requires sudo)
- **Windows**: Run `install.bat` as administrator

### Using Docker

Pull and run the Docker image:

```bash
docker pull ghcr.io/parcoil/updatectrl:latest
docker run -d --name updatectrl ghcr.io/parcoil/updatectrl:latest
```

Or build the multi-platform image locally:

```bash
# Linux/macOS
./build-docker.sh

# Windows
build-docker.bat
```

## First Setup

1. Initialize the configuration:

```bash
sudo updatectrl init
```

2. Edit the config file at `/etc/updatectrl/updatectrl.yaml`

3. Add your first project:

For static projects (e.g., static site generators):

```yaml
interval: 900  # 15 minutes in seconds
projects:
  - name: mysite
    path: /path/to/site
    repo: https://github.com/user/site.git
    type: static
    buildCommand: npm run build
```

For PM2 projects (Node.js applications):

```yaml
interval: 900  # 15 minutes in seconds
projects:
  - name: myapp
    path: /path/to/app
    repo: https://github.com/user/app.git
    type: pm2
    buildCommand: npm install && npm run build
```

For Docker projects:

```yaml
interval: 900  # 15 minutes in seconds
projects:
  - name: myproject
    path: /path/to/project
    repo: https://github.com/user/project.git
    type: docker
    buildCommand: docker compose up -d --build
```

For image-based projects (pre-built Docker images):

```yaml
interval: 900  # 15 minutes in seconds
projects:
  - name: my-app
    type: image
    image: ghcr.io/user/my-app:latest
    port: "80:80"
    env:
      NODE_ENV: production
```

4. The daemon will start automatically and check for updates every 15 minutes.

## Advanced: Running Updatectrl in Docker

For advanced users who prefer containerized deployment, updatectrl can run inside Docker with automatic container discovery.

### Building Multi-Platform Images

Use the provided build scripts to create images for Linux AMD64 and ARM64:

```bash
# Linux/macOS
./build-docker.sh

# Windows
build-docker.bat
```

### Running in Docker

```yaml
version: '3.8'
services:
  updatectrl:
    image: ghcr.io/parcoil/updatectrl:latest
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - DOCKER_HOST=unix:///var/run/docker.sock
      - UPDATECTL_INTERVAL=600  # 10 minutes in seconds
    restart: unless-stopped
```

When running in Docker, updatectrl automatically discovers and manages running containers. Run with:

```bash
docker compose up -d
```

## Direct Docker Run

For direct Docker execution without Compose:

```bash
# Linux
docker run -d --name updatectrl -v /var/run/docker.sock:/var/run/docker.sock ghcr.io/parcoil/updatectrl:latest

# Windows (Method 1 - Named Pipe)
docker run -d --name updatectrl -v //./pipe/docker_engine://./pipe/docker_engine -e DOCKER_HOST=npipe:////./pipe/docker_engine ghcr.io/parcoil/updatectrl:latest

# Windows (Method 2 - Alternative with privileged mode)
docker run -d --name updatectrl --privileged -v /var/run/docker.sock:/var/run/docker.sock -e DOCKER_HOST=unix:///var/run/docker.sock ghcr.io/parcoil/updatectrl:latest
```