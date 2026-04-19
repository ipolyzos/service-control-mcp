# Architecture

## Overview

`service-control-mcp` is a stdio MCP server. The process reads
JSON-RPC messages from stdin, dispatches to the matching tool plugin,
and writes responses to stdout. No network listener is opened.

```
stdin  ──► MCP Runtime (JSON-RPC decode)
                │
                ▼
           Plugin Registry
                │
          ┌─────┴──────┐
          │             │
     Read plugins   Write plugins
          │             │
          └─────┬───────┘
                │
           Platform Detector
                │
         ┌──────┴──────┐
         │              │
    systemd adapter  launchd adapter
         │              │
      systemctl      launchctl
```

## Package Structure

### `cmd/service-control-mcp`

Entry point. Builds the plugin registry, selects the platform adapter,
wires the audit logger, and starts the MCP runtime on stdio.

### `internal/runtime`

Owns the MCP server lifecycle and JSON-RPC dispatch loop. Accepts a
pre-built plugin registry and delegates all tool logic to plugins.
Has no knowledge of individual tools or platform specifics.

### `internal/plugin`

Defines the `Tool` interface that every plugin must satisfy:

```go
type Tool interface {
    Name() string
    Description() string
    InputSchema() json.RawMessage
    Execute(ctx context.Context, args map[string]any) (any, error)
}
```

The `Registry` maps tool names to `Tool` implementations. Plugins
self-register at startup.

### `internal/platform`

Defines the `Manager` interface that each platform adapter implements:

```go
type Manager interface {
    ListServices(ctx context.Context) ([]ServiceInfo, error)
    GetStatus(ctx context.Context, name string) (ServiceStatus, error)
    Describe(ctx context.Context, name string) (ServiceDetail, error)
    GetLogs(ctx context.Context, name string, lines int) ([]string, error)
    Start(ctx context.Context, name string) error
    Stop(ctx context.Context, name string) error
    Restart(ctx context.Context, name string) error
    Reload(ctx context.Context, name string) error
}
```

`Detect()` returns the `Manager` for the current OS, or an error on
unsupported platforms.

### `internal/platform/systemd`

Wraps `systemctl` with an allowlisted subcommand set and 30-second
execution timeout. Parses `systemctl show` output into normalized
`ServiceDetail` structs. Never escalates privileges.

### `internal/platform/launchd`

Wraps `launchctl` with an allowlisted subcommand set and 30-second
execution timeout. Operates under the invoking user's bootstrap context.

### `internal/security`

Provides `ValidateServiceName(name string) error`, which enforces:
- Length 1–128 characters
- Characters restricted to `[a-zA-Z0-9._@:-]`
- No path traversal sequences (`..`, `/`)

Used by every tool before any platform call.

### `internal/audit`

Provides a structured JSON audit logger. Every write operation records:

```json
{
  "audit_id": "<uuid-v4>",
  "timestamp": "<RFC3339>",
  "action": "stop_service",
  "service_name": "nginx",
  "manager": "systemd",
  "outcome": "success",
  "error": ""
}
```

Audit entries are written to stderr by default so they do not
interfere with the stdio MCP channel.

### `plugins/<tool_name>`

Each tool lives in its own package. A plugin is responsible for:

1. Declaring its name, description, and JSON Schema.
2. Validating inputs using `internal/security`.
3. Delegating to the platform `Manager`.
4. Formatting a normalized output struct.
5. Logging write operations through `internal/audit`.

## Security Design Decisions

**No shell execution.** Platform adapters call `exec.CommandContext`
directly with a fixed binary path and a pre-validated argument list.
`/bin/sh -c` is never used.

**Timeout-bounded execution.** Every `exec.CommandContext` call uses a
`context.WithTimeout` with a 30-second ceiling.

**Fail closed.** Unknown platforms, unsupported managers, and invalid
service names all return typed errors. No operation is attempted on
ambiguous input.

**Audit trail.** Write operations are recorded regardless of outcome.
A failed `stop_service` is audited the same as a successful one.
