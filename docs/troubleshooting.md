# Troubleshooting

Common issues and their solutions.

## Daemon Not Starting

**Symptoms:** `systemctl status updatectrl` shows failed

**Solutions:**

- Check systemd service file: `cat /etc/systemd/system/updatectrl.service`
- Verify user has permissions for project paths
- Check logs: `journalctl -u updatectrl`
- Try running `updatectrl init` again

## Git Pull Failures

**Symptoms:** "Git pull failed" in logs

**Solutions:**

- Ensure SSH keys are set up for private repos
- Check repository permissions
- Verify the path exists and is a Git repository

## Build Command Failures

**Symptoms:** Docker/PM2 commands fail

**Solutions:**

- Test commands manually in the project directory
- Check for missing dependencies (Docker, PM2)
- Verify environment variables are available

## Docker Socket Issues (Running in Docker)

**Symptoms:** "Failed to list containers: exit status 1" when running updatectrl in Docker

**Solutions:**

**Linux:**
- Ensure Docker socket is mounted: `-v /var/run/docker.sock:/var/run/docker.sock`
- Check socket permissions: `ls -la /var/run/docker.sock`
- Try running with privileged mode: `--privileged`

**Windows:**
- **Method 1 (Named Pipe):** Use `-v //./pipe/docker_engine://./pipe/docker_engine -e DOCKER_HOST=npipe:////./pipe/docker_engine`
- **Method 2 (Unix socket emulation):** Use `--privileged -v /var/run/docker.sock:/var/run/docker.sock -e DOCKER_HOST=unix:///var/run/docker.sock`
- Ensure Docker Desktop is running
- Try restarting Docker Desktop

**General:**
- Test Docker access: `docker ps` should work from within the container
- Check container logs: `docker logs updatectrl`
- Verify Docker daemon is accessible

## Permission Issues

**Symptoms:** Access denied errors

**Solutions:**

- Run updatectrl as appropriate user (not root if possible)
- Ensure project directories are writable by the service user
- Check file ownership: `ls -la /path/to/project`
