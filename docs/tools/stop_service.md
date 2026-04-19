# Tool: stop_service

## Description

Stops a running service. Because stopping a service is a disruptive
operation, the caller must pass `confirm: true` explicitly. The
operation is recorded in the audit log regardless of outcome.

## Input

```json
{
  "service_name": "nginx.service",
  "confirm": true
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `service_name` | string | yes | Name of the service to stop |
| `confirm` | bool | yes | Must be `true` to proceed |

`service_name` must match `[a-zA-Z0-9._@:-]+` and be 1–128 characters.
If `confirm` is `false` or absent, the request is rejected before any
platform call is made.

## Output

```json
{
  "service_name": "nginx.service",
  "manager": "systemd",
  "outcome": "success"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `service_name` | string | Service name as provided |
| `manager` | string | Platform adapter used |
| `outcome` | string | `"success"` |

## Errors

| Condition | Error message |
|-----------|---------------|
| Invalid service name | `stop_service: validation error: field "service_name": …` |
| `confirm` absent or `false` | `stop_service: validation error: field "confirm": must be true to execute this destructive operation` |
| Manager error | `stop_service: <adapter error>` |

## Security

- `service_name` is validated before the platform call.
- `confirm: true` is enforced as a guard against accidental invocation.
- No privilege escalation; operates under the invoking user's permissions.
- The underlying platform command is `systemctl stop <name>` or
  `launchctl bootout <domain> <service>` — never `sh -c`.

## Audit

An audit entry is written to stderr on every execution (including
failures caused by missing `confirm`):

```json
{
  "audit_id": "9c2e…",
  "timestamp": "2024-04-01T10:05:00Z",
  "action": "stop_service",
  "service_name": "nginx.service",
  "manager": "systemd",
  "outcome": "success",
  "error": ""
}
```
