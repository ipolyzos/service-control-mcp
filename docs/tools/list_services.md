# Tool: list_services

## Description

Enumerates all services visible to the current user through the active
platform adapter. The result set reflects what the invoking user's permissions
allow — on systemd this typically includes all system units; on launchd it
reflects the user's bootstrap domain.

**When to use:** Use this tool to discover what services are available before
calling `get_service_status` or `describe_service`. Because it accepts no
arguments, it is safe to call without any prior knowledge of the host.

**Platform support**

| Platform | Backend | Notes |
|----------|---------|-------|
| Linux (systemd) | `systemctl list-units` | Returns all loaded units |
| macOS (launchd) | `launchctl list` | Returns services in the user bootstrap domain |

## Input

This tool accepts no arguments.

```json
{}
```

## Output

```json
{
  "manager": "systemd",
  "count": 3,
  "services": [
    {
      "name": "nginx.service",
      "manager": "systemd",
      "state": "running",
      "description": "A high performance web server"
    },
    {
      "name": "sshd.service",
      "manager": "systemd",
      "state": "running",
      "description": "OpenSSH server daemon"
    },
    {
      "name": "postgresql.service",
      "manager": "systemd",
      "state": "stopped",
      "description": "PostgreSQL database server"
    }
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `manager` | string | Active platform adapter (`systemd` or `launchd`) |
| `count` | integer | Number of services in the `services` array |
| `services` | array | List of `ServiceInfo` objects |
| `services[].name` | string | Service unit name |
| `services[].manager` | string | Adapter that reported this service |
| `services[].state` | string | Normalized state (see [States](#states)) |
| `services[].description` | string | Human-readable description; may be empty |

## States

| Value | Meaning |
|-------|---------|
| `running` | Service is active and running |
| `stopped` | Service is inactive or not loaded |
| `failed` | Service has failed and exited |
| `starting` | Service is in the process of starting |
| `stopping` | Service is in the process of stopping |
| `unknown` | State could not be determined from adapter output |

## Errors

This tool does not accept service names as input, so there are no
validation errors. A platform error is returned if the underlying
command fails entirely (e.g. the service manager is not running).

## Security

No arguments are validated (none are accepted). The result set is bounded by
the invoking user's permissions — the server never requests elevated
enumeration.

## Audit

This is a read-only tool. No audit entry is written.
