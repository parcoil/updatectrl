# CLI Reference

Complete command reference for Updatectrl.

## updatectrl

Root command.

```bash
updatectrl [command]
```

### Available Commands

- `init` - Initialize configuration and daemon
- `watch` - Run update daemon manually
- `build` - Run build command for a specific project
- `list` - List configured projects
- `logs` - View updatectrl daemon logs
- `version` - Show version information

## init

Initialize Updatectrl configuration and set up the daemon.

```bash
updatectrl init
```

Creates config file and systemd service (Linux) or Task Scheduler job (Windows).

## watch

Run the update daemon. Checks all projects for updates at configured intervals.

```bash
updatectrl watch
```

Use for manual testing or when daemon is not running.

## build

Run the build command for a specific project.

```bash
updatectrl build [project-name]
```

Executes the configured `buildCommand` for the specified project without pulling changes.

## list

List all configured projects.

```bash
updatectrl list
```

Displays the name, type, and relevant details for each project in the configuration.

## logs

View logs from the updatectrl daemon service.

```bash
updatectrl logs [flags]
```

### Flags

- `-f, --follow` - Follow log output (live tail)
- `-n, --lines int` - Number of log lines to show (default 50)

On Linux, uses `journalctl` to view systemd service logs. On Windows, provides instructions for viewing Task Scheduler logs.

## version

Display version information.

```bash
updatectrl version
```

## Global Flags

- `--help` - Show help
- `--version` - Show version