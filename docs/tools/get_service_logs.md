# Tool: get_service_logs

## Description

Fetches recent log lines for a named service. The number of lines returned is
bounded to prevent unbounded output. Log lines are returned newest-last (chronological
order), matching the output of `journalctl -n`.

**When to use:** Use this tool to inspect recent service output after a start,
stop, or unexpected state change. It is read-only and does not affect service state.

**Platform support**

| Platform | Supported | Backend | Notes |
|----------|-----------|---------|-------|
| Linux (systemd) | Yes | `journalctl -u <name> -n <lines>` | Returns kernel and service log entries |
| macOS (launchd) | No | — | Returns an explicit unsupported error |

## Input

```json
{
  "service_name": "nginx.service",
  "lines": 50
}
```

| Field | Type | Required | Default | Constraints |
|-------|------|----------|---------|-------------|
| `service_name` | string | yes | — | `[a-zA-Z0-9._@:-]+`, 1–128 characters |
| `lines` | integer | no | `100` | 1–500 inclusive |

If `lines` is omitted, the last 100 log entries are returned. Requesting more
than 500 lines returns a validation error.

## Output

```json
{
  "service_name": "nginx.service",
  "manager": "systemd",
  "line_count": 3,
  "lines": [
    "Apr 01 10:00:00 host nginx[1234]: Starting A high performance web server...",
    "Apr 01 10:00:00 host nginx[1234]: Started A high performance web server.",
    "Apr 01 10:00:01 host nginx[1234]: listening on 0.0.0.0:80"
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `service_name` | string | Service name as provided |
| `manager` | string | Platform adapter that retrieved the logs |
| `line_count` | integer | Number of lines in the `lines` array |
| `lines` | array of string | Log lines in chronological order (newest last) |

## Errors

| Condition | Error message |
|-----------|---------------|
| `service_name` missing or not a string | `validation error: field "service_name": …` |
| Invalid characters or length | `validation error: field "service_name": …` |
| `lines` out of range (< 1 or > 500) | `validation error: field "lines": must be between 1 and 500` |
| Platform does not support logs | `launchd: GetLogs not supported` |
| Service not found | Manager-specific error message |

## Security

Both `service_name` and `lines` are validated before any platform call.
Log content is returned verbatim — the server does not filter or redact log output.

## Audit

This is a read-only tool. No audit entry is written.
