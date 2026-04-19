# Tool: validate_service_name

## Description

Validates a service identifier against the allowed character set and length
constraints. Unlike all other tools, an invalid name is **not** a Go-level
error — it is always returned as a structured result with `valid: false` and a
human-readable `error` string. This lets the caller inspect and report the
problem gracefully without error-handling logic.

**When to use:** Use this tool to pre-validate a name before constructing a
write operation, or to give callers a clean way to explain why a
service name is rejected. Note that all write tools run the same validation
internally — passing this check is advisory, not a bypass.

**Platform support:** This tool is platform-independent. It does not invoke
any system command.

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
  "service_name": "my service!",
  "valid": false,
  "error": "validation error: field \"service_name\": contains invalid characters; only [a-zA-Z0-9._@:-] are permitted"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `service_name` | string | Name as provided by the caller |
| `valid` | boolean | `true` if the name passes all validation rules |
| `error` | string | Reason for failure; empty string when `valid` is `true` |

## Validation rules

| Rule | Constraint |
|------|-----------|
| Non-empty | At least 1 character |
| Maximum length | 128 characters |
| Allowed characters | `[a-zA-Z0-9._@:-]` only |
| No path traversal | Must not contain `/` or `..` |

## Errors

A Go-level error (returned as `isError: true` in the MCP response) is returned
only when the `service_name` argument is missing entirely or is not a string.
An invalid name is never a Go-level error.

| Condition | Error type |
|-----------|-----------|
| `service_name` missing | Go-level error |
| `service_name` not a string | Go-level error |
| Name fails validation rules | Structured result with `valid: false` |

## Security

This tool never invokes any platform command. It is a pure in-process
validation check using the same logic applied by all other tools.

## Audit

This is a read-only tool. No audit entry is written.
