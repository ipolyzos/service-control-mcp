# Tool: validate_service_name

## Description

Validates a service identifier against the allowed character set and
length constraints before it is used in a write operation. This tool
never returns a Go error for an invalid name — it always returns a
structured result with `valid: false` and a reason string, so the
caller can inspect and report the problem without an error handler.

## Input

```json
{
  "service_name": "nginx.service"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `service_name` | string | yes | Identifier to validate |

## Output — valid name

```json
{
  "service_name": "nginx.service",
  "valid": true,
  "error": ""
}
```

## Output — invalid name

```json
{
  "service_name": "bad name!",
  "valid": false,
  "error": "validation error: field \"service_name\": contains invalid characters; only [a-zA-Z0-9._@:-] are permitted"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `service_name` | string | Name as provided |
| `valid` | bool | `true` if the name passes all validation rules |
| `error` | string | Reason for failure; empty string when `valid` is `true` |

## Validation rules

| Rule | Constraint |
|------|-----------|
| Length | 1–128 characters |
| Characters | `[a-zA-Z0-9._@:-]` only |
| No path traversal | Must not contain `/` or `..` |

## Errors

A Go-level error is returned only if the `service_name` argument is
missing or is not a string. An invalid name is not a Go error.

## Security

Use this tool to pre-validate names before passing them to write tools.
All write tools run the same validation internally, so this is an
advisory check, not a bypass mechanism.

## Audit

This is a read-only tool. No audit entry is written.
