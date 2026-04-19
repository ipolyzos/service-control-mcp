# Tool: start_service

## Description

Starts a stopped service. The operation is recorded in the audit log
regardless of outcome.

## Input

```json
{
  "service_name": "nginx.service"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `service_name` | string | yes | Name of the service to start |

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

A Go-level error (surfaced as `isError: true` in the MCP response) is
returned on validation failure or manager error. The audit entry is
written with `outcome: "failure"` in both cases.

| Condition | Error message |
|-----------|---------------|
| Invalid service name | `start_service: validation error: field "service_name": …` |
| Manager error | `start_service: <adapter error>` |

## Security

- `service_name` is validated before the platform call.
- No privilege escalation; operates under the invoking user's permissions.
- The underlying platform command is `systemctl start <name>` or
  `launchctl bootstrap <domain> <plist>` — never `sh -c`.

## Audit

An audit entry is written to stderr on every execution:

```json
{
  "audit_id": "3a7f…",
  "timestamp": "2024-04-01T10:00:00Z",
  "action": "start_service",
  "service_name": "nginx.service",
  "manager": "systemd",
  "outcome": "success",
  "error": ""
}
```
