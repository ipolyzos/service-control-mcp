# Tool: reload_service

## Description

Sends a configuration reload signal to a running service (equivalent to
`SIGHUP` on most daemons). This does not stop or restart the process,
so it is not considered a destructive operation and does not require a
confirmation flag. The operation is recorded in the audit log regardless
of outcome.

## Input

```json
{
  "service_name": "nginx.service"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `service_name` | string | yes | Name of the service to reload |

`service_name` must match `[a-zA-Z0-9._@:-]+` and be 1–128 characters.

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
| Invalid service name | `reload_service: validation error: field "service_name": …` |
| Service does not support reload | Manager-specific error |
| Manager error | `reload_service: <adapter error>` |

## Security

- `service_name` is validated before the platform call.
- No privilege escalation; operates under the invoking user's permissions.
- The underlying platform command is `systemctl reload <name>` or
  `launchctl kill HUP <domain>/<service>` — never `sh -c`.

## Audit

An audit entry is written to stderr on every execution:

```json
{
  "audit_id": "f7a1…",
  "timestamp": "2024-04-01T10:15:00Z",
  "action": "reload_service",
  "service_name": "nginx.service",
  "manager": "systemd",
  "outcome": "success",
  "error": ""
}
```
