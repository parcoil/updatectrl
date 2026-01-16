---
# https://vitepress.dev/reference/default-theme-home-page
layout: home

hero:
  name: "Updatectrl"
  text: "Project Auto-Updater"
  tagline: Automatically update and restart your projects - Static sites, PM2 apps, Docker containers, and more
  image:
    src: /logo.svg
    alt: Updatectrl Logo
  actions:
    - theme: brand
      text: Get Started
      link: /quickstart
    - theme: alt
      text: View on GitHub
      link: https://github.com/parcoil/updatectrl

features:
  - title: Multi-Project Support
    icon: <span class="material-symbols-rounded">category</span>
    details: Supports Static sites, PM2 Node.js apps, Docker containers, and pre-built images - all in one tool.
  - title: Git & Registry Updates
    icon: <span class="material-symbols-rounded">code</span>
    details: Automatically pulls Git changes or latest container images and rebuilds/restarts your projects.
  - title: Flexible Deployment
    icon: <span class="material-symbols-rounded">settings</span>
    details: Run natively on your system or inside Docker - choose what works best for your setup.
  - title: Cross-Platform
    icon: <span class="material-symbols-rounded">devices</span>
    details: Works on Linux (systemd) and Windows (Task Scheduler) with simple installation and configuration.
---

## What is Updatectrl?

Updatectrl is a lightweight CLI tool that automatically keeps your projects up-to-date. Whether you're running static websites, Node.js applications with PM2, Docker containers, or pre-built container images, Updatectrl ensures your deployments stay current with minimal configuration.

Simply configure your projects once, and Updatectrl will:

- Check for updates at regular intervals
- Pull the latest changes from Git repositories or container registries
- Execute build commands (when needed)
- Restart services automatically

## How It Works

<div class="how-it-works">

### 1. Configure Your Projects

Define your projects in a simple YAML configuration file:

```yaml
interval: 600 # Check every 10 minutes
projects:
  - name: my-website
    path: /srv/website
    repo: https://github.com/user/website.git
    type: static
    buildCommand: npm run build

  - name: my-api
    path: /srv/api
    repo: https://github.com/user/api.git
    type: pm2
    buildCommand: npm install && npm run build
```

### 2. Initialize & Run

```bash
# Build from source
git clone https://github.com/parcoil/updatectrl.git
cd updatectrl
go build -o updatectrl main.go
sudo mv updatectrl /usr/local/bin/

# Initialize configuration and daemon
updatectrl init
```

### 3. Automatic Updates

Updatectrl runs as a background service, checking for updates and keeping your projects current.

</div>

## Use Cases

### Static Site Generators

Perfect for Hugo, Jekyll, or Next.js static sites. Updatectrl pulls the latest content and rebuilds your site automatically.

### Node.js Applications

Manage PM2 applications with automatic dependency updates and process restarts.

### Docker Deployments

Handle both git-based Docker projects and pre-built container images from registries like Docker Hub or GitHub Container Registry.

### Mixed Environments

Run different types of projects side-by-side - static sites, APIs, and containerized services all managed by a single tool.

## Quick Examples

### Static Website

```yaml
projects:
  - name: blog
    path: /srv/blog
    repo: https://github.com/user/blog.git
    type: static
    buildCommand: hugo --minify
```

### PM2 API Server

```yaml
projects:
  - name: api
    path: /srv/api
    repo: https://github.com/user/api.git
    type: pm2
    buildCommand: npm install && npm run build
```

### Docker Application

```yaml
projects:
  - name: webapp
    path: /srv/webapp
    repo: https://github.com/user/webapp.git
    type: docker
    buildCommand: docker compose up -d --build
```

### Container Image

```yaml
projects:
  - name: dashboard
    type: image
    image: ghcr.io/user/dashboard:latest
    port: "3000:80"
    env:
      NODE_ENV: production
```

## Why Choose Updatectrl?

- **Simple Configuration**: YAML-based config that's easy to understand and modify
- **Resource Efficient**: Lightweight Go binary with minimal system impact
- **Flexible**: Supports multiple deployment strategies in one tool
- **Reliable**: Built-in error handling and logging for troubleshooting
- **Cross-Platform**: Works on Linux and Windows with native service integration

## Get Started

Ready to automate your project updates? [Follow the Quick Start Guide](/quickstart) to get Updatectrl running in minutes.
