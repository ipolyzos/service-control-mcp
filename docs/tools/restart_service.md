# Tool: restart_service

## Description

Stops and restarts a service. Because a restart causes a brief service
interruption, the caller must pass `confirm: true` explicitly. The
operation is recorded in the audit log regardless of outcome.

**When to use:** Use when you need a full stop/start cycle (not just a reload signal).

**Platform support**

| Platform | Backend command | Notes |
|----------|-----------------|-------|
| Linux (systemd) | `systemctl restart <name>` | Hard restart |
| macOS (launchd) | `launchctl kickstart -k <domain>/<service>` | Kickstart with kill |

## Input

```json
{
  "service_name": "nginx.service",
  "confirm": true
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `service_name` | string | yes | Name of the service to restart |
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
| Invalid service name | `restart_service: validation error: field "service_name": …` |
| `confirm` absent or `false` | `restart_service: validation error: field "confirm": must be true to execute this destructive operation` |
| Manager error | `restart_service: <adapter error>` |

## Security

- `service_name` is validated before the platform call.
- `confirm: true` is enforced as a guard against accidental invocation.
- No privilege escalation; operates under the invoking user's permissions.
- The underlying platform command is `systemctl restart <name>` or
  `launchctl kickstart -k <domain>/<service>` — never `sh -c`.

## Audit

An audit entry is written to stderr on every execution:

```json
{
  "audit_id": "b1f4…",
  "timestamp": "2024-04-01T10:10:00Z",
  "action": "restart_service",
  "service_name": "nginx.service",
  "manager": "systemd",
  "outcome": "success",
  "error": ""
}
```
