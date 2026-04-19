# Architecture

## Overview

`service-control-mcp` is a stdio MCP server. The process reads JSON-RPC
messages from stdin, dispatches each request to the matching tool plugin,
and writes responses to stdout. No network listener is opened. The process
runs as the invoking user and inherits that user's service manager permissions.

```
MCP Client (application / IDE)
         │  stdin / stdout (JSON-RPC)
         ▼
┌─────────────────────────────────────────┐
│              MCP Runtime                │
│   (JSON-RPC decode → tool dispatch)     │
└────────────────┬────────────────────────┘
                 │
         ┌───────▼────────┐
         │ Plugin Registry│
         └───────┬────────┘
         ┌───────┴────────────────┐
         │                        │
   Read plugins             Write plugins
   (no side effects)        (validate → execute → audit)
         │                        │
         └──────────┬─────────────┘
                    │
            Platform Detector
                    │
          ┌─────────┴──────────┐
          │                    │
   systemd adapter       launchd adapter
   (Linux)               (macOS)
          │                    │
     systemctl            launchctl
```

## Request lifecycle

1. The MCP client sends a `tools/call` JSON-RPC request over stdin.
2. `internal/runtime` decodes the request and looks up the tool by name in the
   plugin registry.
3. The plugin validates all arguments (via `internal/security`).
4. For write tools: the plugin invokes the platform adapter, then writes a
   JSON audit entry to stderr regardless of outcome.
5. The plugin returns a normalized result struct, which `internal/runtime`
   encodes as a JSON string in the MCP `CallToolResult`.
6. The response is written to stdout.

Tool-level errors (invalid input, service not found, manager failure) are
returned as `isError: true` MCP results — they do not terminate the server.
Only unrecoverable startup errors exit the process.

## Package reference

### `cmd/service-control-mcp`

Entry point. Detects the platform, constructs the plugin registry, wires the
audit logger, and starts the MCP runtime. The `version` variable is injected
at build time via `-ldflags "-X main.version=<tag>"`; local builds default
to `"dev"`.

### `internal/runtime`

Owns the MCP server lifecycle and the JSON-RPC dispatch loop. Accepts a
pre-built plugin registry; has no knowledge of individual tools or platform
specifics. `ExecuteTool` is also exported for direct use in integration tests.

### `internal/plugin`

Defines the `Tool` interface every plugin must implement:

```go
type Tool interface {
    Name()        string
    Description() string
    InputSchema() json.RawMessage
    Execute(ctx context.Context, args map[string]any) (any, error)
}
```

`Registry` maps tool names to `Tool` implementations and is safe for
concurrent reads after all tools have been registered at startup.

### `internal/platform`

Defines the `Manager` interface that each platform adapter implements, along
with the normalized data types (`ServiceInfo`, `ServiceStatus`, `ServiceDetail`)
shared across all adapters.

`Detect()` returns the `Manager` for the current OS by consulting a registry
populated by adapter `init()` functions. `DetectFor(goos string)` is exported
for testing with synthetic OS names.

```go
type Manager interface {
    Name() string
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

### `internal/platform/systemd`

Wraps `systemctl` with an allowlisted subcommand set and a 30-second
execution timeout. Parses `systemctl show` output into normalized structs.
Registers itself for `GOOS=linux` via `init()`.

The adapter uses an `Executor` interface around `exec.CommandContext` so
unit tests can inject a fake executor — no real `systemctl` is required.

### `internal/platform/launchd`

Wraps `launchctl` with an allowlisted subcommand set and a 30-second
execution timeout. Operates within the invoking user's bootstrap context.
Registers itself for `GOOS=darwin` via `init()`.

Uses the same `Executor` abstraction as the systemd adapter.

> **Note:** `get_service_logs` is not supported on launchd. The adapter
> returns a typed error; the tool surfaces this to the caller cleanly.

### `internal/security`

Provides two validators called by every tool before any platform invocation:

- `ValidateServiceName(name string) error` — enforces length (1–128),
  character set (`[a-zA-Z0-9._@:-]`), and no path traversal (`..`, `/`).
- `ValidateLogLines(n int) error` — enforces range 1–500.
- `ValidateConfirmFlag(confirm bool) error` — rejects `false` or absent
  confirm on destructive operations.

### `internal/audit`

Provides a structured JSON audit logger. Entries are written to a caller-
supplied `io.Writer` (stderr in production) so they never interfere with
the stdio MCP channel. Each entry includes a cryptographically random
`audit_id` generated via `crypto/rand`.

### `plugins/<tool_name>`

One package per MCP tool. Each plugin is responsible for:

1. Declaring its name, description, and JSON Schema (`InputSchema()`).
2. Validating inputs using `internal/security`.
3. Delegating to the platform `Manager` through a **locally-defined minimal
   interface** — plugins do not import `platform.Manager` directly. This
   keeps each plugin independently testable with a simple stub.
4. Returning a normalized output struct.
5. Writing an audit entry for write operations (success and failure alike).

## Adding a platform adapter

1. Create `internal/platform/<os>/` with a file that implements `platform.Manager`.
2. In an `init()` function, call `platform.Register("<goos>", factory)`.
3. Add a blank import of the new package in `cmd/service-control-mcp/main.go`.
4. Add the adapter's `Executor` interface and a fake implementation for tests.

## Adding a tool plugin

See [CONTRIBUTING.md](../CONTRIBUTING.md#adding-a-tool-plugin) for the full
step-by-step checklist.

## Security design

See [security.md](security.md) for the threat model and security invariants.
