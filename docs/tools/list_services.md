# Tool: list_services

## Description

Enumerates all services visible to the current user through the active
platform adapter.

## Input

```json
{}
```

No arguments are required or accepted.

## Output

```json
{
  "manager": "systemd",
  "count": 2,
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
      "state": "stopped",
      "description": "OpenSSH server daemon"
    }
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `manager` | string | Active platform adapter name (`systemd` or `launchd`) |
| `count` | int | Number of services returned |
| `services` | array | List of `ServiceInfo` objects |
| `services[].name` | string | Service unit name |
| `services[].manager` | string | Adapter that reported this service |
| `services[].state` | string | Normalized state (see [States](#states)) |
| `services[].description` | string | Human-readable service description (may be empty) |

## States

| Value | Meaning |
|-------|---------|
| `running` | Service is active and running |
| `stopped` | Service is inactive |
| `failed` | Service has failed |
| `starting` | Service is in the process of starting |
| `stopping` | Service is in the process of stopping |
| `unknown` | State could not be determined |

## Security

- No service names are accepted as input; no validation is required.
- The result set is limited to services visible under the invoking user's
  permissions.

## Audit

This is a read-only tool. No audit entry is written.
