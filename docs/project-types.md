# Project Types

Updatectrl supports different types of projects with varying update strategies.

## Static

For static websites and projects that need Git pulls and optional build commands.

**Process:**

1. Pull latest Git changes
2. Execute the `buildCommand` (if configured)

**Example:**

```yaml
type: static
buildCommand: npm run build
```

**Use cases:** Static site generators (Hugo, Jekyll, Next.js export), documentation sites, simple web projects

**Requirements:** Git must be installed

## PM2

For Node.js applications managed by PM2 process manager.

**Process:**

1. Pull latest Git changes
2. Execute the `buildCommand` (if configured)
3. Restart the PM2 process

**Example:**

```yaml
type: pm2
buildCommand: npm install && npm run build
```

**Requirements:** PM2 must be installed and the app started with `pm2 start`

**Example PM2 command:** `pm2 start index.js --name my-app` (name must match the name in updatectrl config)

**Use cases:** Node.js web servers, APIs, real-time applications

## Docker

For containerized applications using Docker or Docker Compose. This is for projects with Dockerfiles or docker-compose.yml files that need to be built/rebuilt.

**Process:**

1. Pull latest Git changes
2. Execute the `buildCommand` (typically `docker compose up -d --build` or `docker build` + `docker run`)

**Example (Docker Compose):**

```yaml
type: docker
buildCommand: docker compose up -d --build
```

**Example (Direct Docker):**

```yaml
type: docker
buildCommand: docker build -t myapp . && docker stop myapp || true && docker rm myapp || true && docker run -d --name myapp -p 8080:8080 myapp
```

**Use cases:** Applications with Dockerfiles, Docker Compose stacks, containerized services

**Requirements:** Docker must be installed and running

**Note:** Use this for projects that need building. For pre-built images from registries, use the `image` type instead.

## Image

For applications deployed as pre-built Docker images from container registries. This monitors specific images for updates and restarts containers when new versions are available.

**Process:**

1. Check if the remote image has a newer digest than the local image
2. Pull the latest version if needed
3. Restart the container with the updated image
4. Preserve port mappings, environment variables, and container names

**Example:**

```yaml
type: image
image: ghcr.io/user/my-app:latest
port: "3000:80"
env:
  NODE_ENV: production
containerName: my-custom-app
```

**Requirements:** Docker must be installed and running.

**Use cases:** Pre-built applications, microservices, third-party containers, applications distributed as Docker images

**Note:** When running updatectrl in Docker mode, it automatically discovers and monitors running containers. Docker Compose managed containers are skipped to avoid conflicts with Compose orchestration.
