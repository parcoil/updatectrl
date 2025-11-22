# Quick Start Guide

This guide will get you up and running with Updatectl in minutes.

## Prerequisites

- Go installed (for building from source)
- Git installed
- Docker or PM2 depending on your projects
- Docker Compose (for containerized deployment)

## Installation

### From Source

```bash
go build -o updatectl
sudo mv updatectl /usr/local/bin/
```

### Using Docker

Pull the official Docker image:

```bash
docker pull ghcr.io/parcoil/updatectl:latest
```

## First Setup

1. Initialize the configuration:

```bash
sudo updatectl init
```

2. Edit the config file at `/etc/updatectl/updatectl.yaml`

3. Add your first project:

For git-based projects:

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

## Recommended: Running with Docker Compose

> [!TIP]
> Running updatectl inside Docker is the recommended approach. It provides isolation, automatic container discovery, and easy management.

When running in Docker, updatectl automatically discovers and manages all running containers with images from Docker Hub or GHCR. No manual configuration needed!

Create a `docker-compose.yml`:

```yaml
version: '3.8'
services:
  updatectl:
    image: ghcr.io/parcoil/updatectl:latest
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - DOCKER_HOST=unix:///var/run/docker.sock
      - UPDATECTL_INTERVAL=600  # 10 minutes in seconds
    restart: unless-stopped

  # Your other services...
  my-app:
    image: docker.io/my-app:latest
    ports:
      - "80:80"
    # ...
```

Environment variables:
- `UPDATECTL_INTERVAL`: Check interval in seconds (default: 600)

### Benefits of Docker Deployment

- **Automatic Discovery**: Finds all running containers automatically
- **No Configuration**: Works out-of-the-box with existing containers
- **Isolation**: Runs in its own container without affecting host system
- **Easy Updates**: Update updatectl itself by rebuilding the image

### How It Works

Updatectl will automatically monitor all containers with `docker.io/`, `ghcr.io/`, or registry images and restart them when new versions are available. It inspects container ports and environment variables to preserve your configuration.

Then run:

```bash
docker compose up -d
```

Updatectl runs inside a container and controls Docker on the host via the mounted socket.

## Direct Docker Run (Without Compose)

If you prefer not to use Docker Compose, you can run updatectl directly with `docker run`.

### Linux

```bash
docker run -d \
  --name updatectl \
  -e UPDATECTL_INTERVAL=30 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  ghcr.io/parcoil/updatectl:latest
```

### Windows (PowerShell)

```powershell
docker run -d `
  --name updatectl `
  -e UPDATECTL_INTERVAL=30 `
  -v //./pipe/docker_engine://./pipe/docker_engine `
  -e DOCKER_HOST=npipe:////./pipe/docker_engine `
  ghcr.io/parcoil/updatectl:latest
```

### Windows (Command Prompt)

```cmd
docker run -d ^
  --name updatectl ^
  -e UPDATECTL_INTERVAL=30 ^
  -v //./pipe/docker_engine://./pipe/docker_engine ^
  -e DOCKER_HOST=npipe:////./pipe/docker_engine ^
  ghcr.io/parcoil/updatectl:latest
```