# Tool: describe_service

## Description

Returns rich metadata for a named service — unit file path, start command,
adapter-specific key-value properties, and normalized state. This is the most
detailed read operation available.

**When to use:** Use this tool when you need to understand _how_ a service is
configured, not just whether it is running. It is more expensive than
`get_service_status` because it parses additional adapter output; prefer
`get_service_status` when only the runtime state is needed.

**Platform support**

| Platform | Backend | Notes |
|----------|---------|-------|
| Linux (systemd) | `systemctl show` | Populates `unit_file_path`, `exec_start`, and `extra` |
| macOS (launchd) | `launchctl list` + plist read | `extra` contains plist-level keys |

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
  "description": "A high performance web server and a reverse proxy server",
  "unit_file_path": "/usr/lib/systemd/system/nginx.service",
  "exec_start": "/usr/sbin/nginx -g 'daemon on; master_process on;'",
  "extra": {
    "Type": "forking",
    "Restart": "on-failure",
    "WantedBy": "multi-user.target"
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Service name as provided |
| `manager` | string | Platform adapter (`systemd` or `launchd`) |
| `state` | string | Normalized state — see [States](#states) |
| `description` | string | Human-readable service description; may be empty |
| `unit_file_path` | string | Absolute path to the unit or plist file; may be empty |
| `exec_start` | string | Command used to start the service; may be empty |
| `extra` | object | Adapter-specific key-value properties; may be empty |

## States

| Value | Meaning |
|-------|---------|
| `running` | Service is active and running |
| `stopped` | Service is inactive |
| `failed` | Service has failed |
| `starting` | Service is in the process of starting |
| `stopping` | Service is in the process of stopping |
| `unknown` | State could not be determined |

## Errors

| Condition | Error message |
|-----------|---------------|
| `service_name` missing or not a string | `validation error: field "service_name": …` |
| Invalid characters or length | `validation error: field "service_name": …` |
| Service not found | Manager-specific error message |

## Security

`service_name` is validated through `security.ValidateServiceName` before any
platform call. No platform call is made on invalid input.

## Audit

This is a read-only tool. No audit entry is written.
