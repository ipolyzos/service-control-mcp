# service-control-mcp

A stdio [Model Context Protocol](https://modelcontextprotocol.io/) server that
exposes system service management as structured, audited MCP tools — safely
bridging MCP clients to systemd (Linux) and launchd (macOS).

## Overview

`service-control-mcp` lets an MCP client (IDE extension, automation, or other integration) inspect and control system services without opening a network
port or escalating privileges. Every write operation is validated, allowlisted,
and recorded in a structured audit log before any system command is executed.

**Supported platforms**

| Platform | Service manager |
|----------|----------------|
| Linux    | systemd        |
| macOS    | launchd        |

## Tools

### Read operations

| Tool | Description |
|------|-------------|
| [`list_services`](docs/tools/list_services.md) | Enumerate services visible to the current user |
| [`get_service_status`](docs/tools/get_service_status.md) | Return normalized status for a named service |
| [`describe_service`](docs/tools/describe_service.md) | Return metadata and manager-specific detail for a service |
| [`get_service_logs`](docs/tools/get_service_logs.md) | Fetch recent bounded log lines for a service |
| [`validate_service_name`](docs/tools/validate_service_name.md) | Validate a service identifier before use in write operations |

### Write operations

| Tool | Description |
|------|-------------|
| [`start_service`](docs/tools/start_service.md) | Start a stopped service |
| [`stop_service`](docs/tools/stop_service.md) | Stop a running service (requires `confirm: true`) |
| [`restart_service`](docs/tools/restart_service.md) | Restart a service (requires `confirm: true`) |
| [`reload_service`](docs/tools/reload_service.md) | Send a configuration reload signal to a service |

All write operations emit a structured JSON audit entry to stderr. See
[docs/security.md](docs/security.md) for the full security model.

## Installation

**From source (recommended for development):**

```sh
git clone https://github.com/ipolyzos/service-control-mcp.git
cd service-control-mcp
make build
# Binary: bin/service-control-mcp
```

**Using `go install`:**

```sh
go install github.com/ipolyzos/service-control-mcp/cmd/service-control-mcp@latest
```

## MCP Client Configuration

Add the server to your MCP client configuration. The server communicates over
stdio and requires no additional arguments or environment variables.

**Example `mcpServers` entry (location depends on your MCP client):**

```json
{
  "mcpServers": {
    "service-control": {
      "command": "service-control-mcp"
    }
  }
}
```

**Generic MCP client:**

```json
{
  "mcpServers": {
    "service-control": {
      "command": "/path/to/service-control-mcp"
    }
  }
}
```

The process runs as the invoking user and inherits their service manager
permissions — no additional configuration is needed for read operations.
Write operations (start, stop, restart, reload) require the user to have
the appropriate permissions on the host service manager.

## Documentation

| Document | Contents |
|----------|----------|
| [docs/README.md](docs/README.md) | Documentation index and tool reference matrix |
| [docs/architecture.md](docs/architecture.md) | Internal design, package structure, request lifecycle |
| [docs/security.md](docs/security.md) | Threat model, security invariants, audit log format |
| [docs/branching.md](docs/branching.md) | Branching strategy and release workflow |
| [docs/tools/](docs/tools/) | Per-tool input/output schemas and examples |

## Development

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full development workflow,
testing requirements, and contribution guidelines.

```sh
make build      # build binary → bin/service-control-mcp
make test       # run tests with race detector + coverage
make lint       # run golangci-lint
make ci         # full CI gate (format, vet, lint, test, coverage)
```

## License

Apache License, Version 2.0. See [LICENSE](LICENSE).
