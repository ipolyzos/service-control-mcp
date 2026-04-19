# Tool: start_service

## Description

Starts a stopped service. The underlying platform command is invoked directly
with the validated service name — no shell is used. The operation is recorded
in the audit log regardless of whether it succeeds or fails.

**When to use:** Use this tool to bring a stopped or failed service back up.
If the service is already running, the underlying platform command may return
an error or be a no-op depending on the adapter.

**Platform support**

| Platform | Backend command | Notes |
|----------|----------------|-------|
| Linux (systemd) | `systemctl start <name>` | Starts the unit; may trigger dependent units |
| macOS (launchd) | `launchctl bootstrap <domain> <plist>` | Bootstraps the service into the user domain |

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
| `outcome` | string | Always `"success"` on a successful response |

## Errors

Errors are returned as `isError: true` in the MCP response. The audit entry
is written with `outcome: "failure"` for all error cases.

| Condition | Error message |
|-----------|---------------|
| `service_name` missing or not a string | `start_service: validation error: field "service_name": …` |
| Invalid characters or length | `start_service: validation error: field "service_name": …` |
| Platform manager error | `start_service: <adapter error>` |

## Security

- `service_name` is validated through `security.ValidateServiceName` before any
  platform call. No platform call is made on invalid input.
- No privilege escalation — the command runs under the invoking user's permissions.
- The underlying command is never executed via a shell.

## Audit

An audit entry is written to stderr on every execution, including failures:

```json
{
  "audit_id":     "3a7f9e4c1b2d6f8a0e5c7b9d3f1a4e2c",
  "timestamp":    "2024-04-01T10:00:00.000000000Z",
  "action":       "start_service",
  "service_name": "nginx.service",
  "manager":      "systemd",
  "outcome":      "success",
  "error":        ""
}
```
