# Tool: get_service_status

## Description

Returns the normalized runtime status of a single named service.

## Input

```json
{
  "service_name": "nginx.service"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `service_name` | string | yes | Name of the service to query |

`service_name` must match `[a-zA-Z0-9._@:-]+` and be 1–128 characters.

## Output

```json
{
  "name": "nginx.service",
  "manager": "systemd",
  "state": "running",
  "sub_state": "running",
  "load_state": "loaded",
  "description": "A high performance web server",
  "active_since": "2024-04-01T10:00:00Z",
  "pid": 1234
}
```

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Service name as provided |
| `manager` | string | Platform adapter (`systemd` or `launchd`) |
| `state` | string | Normalized state |
| `sub_state` | string | Manager-specific sub-state (may be empty) |
| `load_state` | string | Unit load state (systemd only; may be empty) |
| `description` | string | Human-readable service description (may be empty) |
| `active_since` | string | ISO-8601 timestamp when the service became active (may be empty) |
| `pid` | int | Main process PID when running; 0 otherwise |

## Errors

| Condition | Error message |
|-----------|---------------|
| Invalid service name | `validation error: field "service_name": …` |
| Service not found | Manager-specific error, e.g. `Unit nginx.service not found` |

## Security

`service_name` is validated before any platform call.

## Audit

This is a read-only tool. No audit entry is written.
