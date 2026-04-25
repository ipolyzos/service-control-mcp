# Contributing

Thank you for your interest in contributing to `service-control-mcp`.

## Prerequisites

- Go 1.26 or later
- `golangci-lint` v1.64.8 ([install](https://golangci-lint.run/usage/install/))
- `make`

## Development Workflow

This project uses **trunk-based development** with a `develop` integration
branch and short-lived feature branches.

### Branch strategy

```
main          ← stable releases only (tagged)
  └── develop ← integration branch
        └── feature/<topic>   ← your work
        └── fix/<topic>
        └── chore/<topic>
```

1. Branch from `develop`:
   ```sh
   git checkout -b feature/<topic> develop
   ```
2. Work in small, focused commits. Sign every commit (`-s`).
3. Open a PR targeting `develop`.
4. CI must be green before merge.
5. Releases are promoted from `develop` to `main` via a `release/<version>` branch.

### Commit messages

Titles must start with an imperative verb from the approved list:
`Add`, `Fix`, `Update`, `Remove`, `Refactor`, `Implement`, `Improve`,
`Document`, `Bump`, `Change`, `Revert`, `Simplify`, `Upgrade`, `Ensure`,
`Support`.

```
Add service name validation to security package

Validates service names against a strict allowlist pattern to
prevent command injection before any manager invocation.

Signed-off-by: Your Name <email@example.com>
```

All commits require `Signed-off-by` (`git commit -s`).

## Building and Testing

```sh
# Build
make build

# Run all tests (with race detector)
make test

# Run a single test package
go test -count=1 -race ./internal/security/...

# Run a single test function
go test -count=1 -race -run TestValidateServiceName ./internal/security/...

# Lint
make lint

# Format check
make fmt-check

# Apply formatting
make fmt

# Full CI check (what the pipeline runs)
make ci
```

## Test Requirements

This project follows strict TDD. **Tests must be written before
production code** for every change. No production code may be committed
without accompanying tests. Bug fixes must begin with a reproducing test.

## Code Organization

```
cmd/service-control-mcp/   entry point
internal/
  audit/                   structured audit logging
  platform/                platform detection and manager interface
    systemd/               systemd adapter
    launchd/               launchd adapter
  plugin/                  tool interface and plugin registry
  runtime/                 MCP server, stdio transport, dispatch
  security/                input validation, service name allowlist
plugins/
  <tool_name>/             one directory per MCP tool plugin
docs/                      architecture and design notes
```

## Adding a Tool Plugin

1. Write `plugins/<tool_name>/plugin_test.go` first (TDD — tests before implementation).
2. Create `plugins/<tool_name>/plugin.go` implementing `plugin.Tool`.
3. Register the plugin in `cmd/service-control-mcp/main.go` (the `tools` slice in `run()`).
4. Document the tool contract in `docs/tools/<tool_name>.md`.
5. Update `README.md` tool table.
6. Add a CHANGELOG entry under `[Unreleased]`.

## Security Constraints

All tools that invoke platform commands must:

- Use the security package to validate service names.
- Pass only allowlisted arguments to platform adapters.
- Set an explicit execution timeout via the platform adapter.
- Record write operations through the audit logger.

No tool may escalate privileges or execute arbitrary shell commands.

## Pull Request Checklist

- [ ] Tests written before implementation
- [ ] All tests pass (`make test`)
- [ ] Lint passes (`make lint`)
- [ ] Format applied (`make fmt`)
- [ ] Commit messages signed and follow the convention
- [ ] Documentation updated for user-visible changes
- [ ] CHANGELOG.md updated under `[Unreleased]`
