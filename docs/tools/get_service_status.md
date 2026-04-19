# Tool: get_service_status

## Description

Returns the normalized runtime status of a single named service. Provides
the essential runtime state — whether the service is running, its PID, and
when it became active — without the full metadata returned by `describe_service`.

**When to use:** Use this tool to check whether a service is running before
attempting a write operation, or to poll for state changes after starting or
stopping a service. For richer metadata (unit file path, exec command, adapter
extras), use `describe_service` instead.

**Platform support**

| Platform | Backend | Notes |
|----------|---------|-------|
| Linux (systemd) | `systemctl show` | Returns `sub_state` and `load_state` |
| macOS (launchd) | `launchctl list` | `sub_state` and `load_state` may be empty |

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
| `state` | string | Normalized state — see [States](#states) |
| `sub_state` | string | Manager-specific sub-state; may be empty on launchd |
| `load_state` | string | Unit load state (systemd only; e.g. `loaded`, `not-found`) |
| `description` | string | Human-readable service description; may be empty |
| `active_since` | string | ISO-8601 timestamp when the service became active; empty if not running |
| `pid` | integer | Main process PID when running; `0` otherwise |

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
| Service not found (systemd) | `Unit nginx.service could not be found` |
| Service not found (launchd) | Manager-specific error message |

## Security

`service_name` is validated through `security.ValidateServiceName` before any
platform call. No platform call is made on invalid input.

## Audit

This is a read-only tool. No audit entry is written.
