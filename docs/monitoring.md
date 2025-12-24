# Monitoring

How to monitor Updatectrl's activity and performance.

## Logs

### Systemd Logs

```bash
journalctl -u updatectrl -f
```

### Manual Runs

When running `updatectrl watch` manually, output goes to stdout.

## Metrics

### Update Frequency

Monitor how often projects are updated:

```bash
journalctl -u updatectrl | grep "Checking" | wc -l
```

### Success Rate

Check for failures:

```bash
journalctl -u updatectrl | grep "failed\|Failed" | tail
```

## Health Checks

### Service Status

```bash
systemctl status updatectrl
```

### Configuration Validation

```bash
python3 -c "import yaml; yaml.safe_load(open('/etc/updatectrl/updatectrl.yaml'))"
```

## Performance

Monitor resource usage:

```bash
ps aux | grep updatectrl
```

Adjust `interval` based on system load.
