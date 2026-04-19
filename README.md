# service-control-mcp

A stdio MCP server that wraps system service managers — systemd (Linux) and
launchd (macOS) — in a safe, structured, read-write interface.

## Overview

`service-control-mcp` exposes service management operations as
[Model Context Protocol](https://modelcontextprotocol.io/) tools. It
provides read and write access to platform service managers through a
plugin architecture, with strict input validation, command allowlisting,
bounded execution timeouts, and structured audit logging on every write
operation.

**Supported platforms**

| Platform | Service manager |
|----------|----------------|
| Linux    | systemd        |
| macOS    | launchd        |

## Tools

### Read

| Tool | Description |
|------|-------------|
| `list_services` | Enumerate manageable services visible to the current user |
| `get_service_status` | Return normalized status for a named service |
| `describe_service` | Return metadata and manager-specific details for a service |
| `get_service_logs` | Fetch recent bounded logs for a service |
| `validate_service_name` | Validate and normalize a service identifier |

### Write

| Tool | Description |
|------|-------------|
| `start_service` | Start a stopped service |
| `stop_service` | Stop a running service (requires `confirm: true`) |
| `restart_service` | Restart a service (requires `confirm: true`) |
| `reload_service` | Send a configuration reload signal to a service |

All write operations emit a structured audit entry (see [Security](#security)).

## Security

- No privilege escalation — operates under the invoking user's permissions.
- Command allowlist only — no arbitrary shell execution.
- Strict service name validation (alphanumeric, hyphens, dots, underscores; 1–128 characters).
- All external commands run with bounded timeouts.
- Every write action is recorded with a unique `audit_id`, timestamp, service
  name, action, and outcome.
- Destructive operations (`stop_service`, `restart_service`) require explicit `confirm: true`.

## Installation

```sh
go install github.com/ipolyzos/service-control-mcp/cmd/service-control-mcp@latest
```

Or build from source:

```sh
git clone https://github.com/ipolyzos/service-control-mcp.git
cd service-control-mcp
make build
```

## MCP Client Configuration

Add the following to your MCP client configuration:

```json
{
  "mcpServers": {
    "service-control": {
      "command": "service-control-mcp"
    }
  }
}
```

The server communicates over stdio and requires no additional arguments.

## Development

See [CONTRIBUTING.md](CONTRIBUTING.md) for the development workflow,
branch strategy, and contribution guidelines.

See [docs/architecture.md](docs/architecture.md) for the design rationale
and internal structure.

## License

Apache License, Version 2.0. See [LICENSE](LICENSE).
