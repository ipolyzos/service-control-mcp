# Tool: describe_service

## Description

Returns metadata and manager-specific details for a named service.
Richer than `get_service_status` — includes unit file path, exec
command, and an `extra` map of adapter-specific properties.

## Input

```json
{
  "service_name": "nginx.service"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `service_name` | string | yes | Name of the service to describe |

`service_name` must match `[a-zA-Z0-9._@:-]+` and be 1–128 characters.

## Output

```json
{
  "name": "nginx.service",
  "manager": "systemd",
  "state": "running",
  "description": "A high performance web server",
  "unit_file_path": "/usr/lib/systemd/system/nginx.service",
  "exec_start": "/usr/sbin/nginx -g 'daemon on; master_process on;'",
  "extra": {
    "Type": "forking",
    "Restart": "on-failure"
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Service name as provided |
| `manager` | string | Platform adapter (`systemd` or `launchd`) |
| `state` | string | Normalized state |
| `description` | string | Human-readable service description (may be empty) |
| `unit_file_path` | string | Path to the unit or plist file (may be empty) |
| `exec_start` | string | Start command (may be empty) |
| `extra` | object | Adapter-specific key-value properties |

## Errors

| Condition | Error message |
|-----------|---------------|
| Invalid service name | `validation error: field "service_name": …` |
| Service not found | Manager-specific error |

## Security

`service_name` is validated before any platform call.

## Audit

This is a read-only tool. No audit entry is written.
