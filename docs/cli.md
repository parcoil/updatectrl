# CLI Reference

Complete command reference for Updatectl.

## updatectl

Root command.

```bash
updatectl [command]
```

### Available Commands

- `init` - Initialize configuration and daemon
- `watch` - Run update daemon manually
- `build` - Run build command for a specific project
- `logs` - View updatectl daemon logs
- `version` - Show version information

## init

Initialize Updatectl configuration and set up the daemon.

```bash
updatectl init
```

Creates config file and systemd service (Linux) or Task Scheduler job (Windows).

## watch

Run the update daemon. Checks all projects for updates at configured intervals.

```bash
updatectl watch
```

Use for manual testing or when daemon is not running.

## build

Run the build command for a specific project.

```bash
updatectl build [project-name]
```

Executes the configured `buildCommand` for the specified project without pulling changes.

## logs

View logs from the updatectl daemon service.

```bash
updatectl logs [flags]
```

### Flags

- `-f, --follow` - Follow log output (live tail)
- `-n, --lines int` - Number of log lines to show (default 50)

On Linux, uses `journalctl` to view systemd service logs. On Windows, provides instructions for viewing Task Scheduler logs.

## version

Display version information.

```bash
updatectl version
```

## Global Flags

- `--help` - Show help
- `--version` - Show version