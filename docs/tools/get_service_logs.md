# Tool: get_service_logs

## Description

Returns recent log lines for a named service. The number of lines is
bounded to prevent unbounded output. Not all platform adapters support
log retrieval; unsupported adapters return an explicit error.

**Platform support**

| Platform | Supported | Backend |
|----------|-----------|---------|
| Linux (systemd) | Yes | `journalctl -u <name> -n <lines>` |
| macOS (launchd) | No | Returns `unsupported` error |

## Input

```json
{
  "service_name": "nginx.service",
  "lines": 50
}
```

| Field | Type | Required | Default | Constraints |
|-------|------|----------|---------|-------------|
| `service_name` | string | yes | — | `[a-zA-Z0-9._@:-]+`, 1–128 chars |
| `lines` | integer | no | `100` | 1–500 |

## Output

```json
{
  "service_name": "nginx.service",
  "manager": "systemd",
  "line_count": 2,
  "lines": [
    "Apr 01 10:00:00 host nginx[1234]: started",
    "Apr 01 10:00:01 host nginx[1234]: listening on :80"
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `service_name` | string | Service name as provided |
| `manager` | string | Platform adapter |
| `line_count` | int | Actual number of lines returned |
| `lines` | array of string | Log lines, newest last |

## Errors

| Condition | Error message |
|-----------|---------------|
| Invalid service name | `validation error: field "service_name": …` |
| `lines` out of range | `validation error: field "lines": …` |
| Unsupported platform | `launchd: GetLogs not supported` |

## Security

Both `service_name` and `lines` are validated before any platform call.

## Audit

This is a read-only tool. No audit entry is written.
