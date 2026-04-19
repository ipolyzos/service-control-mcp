# Changelog

All notable changes to this project are documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-04-19

### Added

- Initial project scaffold: Go module, Apache-2.0 license, README,
  CONTRIBUTING, SECURITY, and CHANGELOG.
- GitHub Actions CI: build, test (with `-race`), lint (golangci-lint),
  format check, module-tidy verification, and an 80 % coverage gate.
- GitHub Actions release workflow: cross-compiles `linux/amd64`,
  `linux/arm64`, `darwin/amd64`, `darwin/arm64`; produces
  `sha256sums.txt`; creates a GitHub Release on `v*.*.*` tags.
- `internal/plugin`: `Tool` interface and thread-safe `Registry`.
- `internal/platform`: `Manager` interface, `ServiceInfo`, `ServiceStatus`,
  `ServiceDetail` types, and `Detect`/`DetectFor` adapter selection with
  a mutex-protected registry.
- `internal/platform/systemd`: wraps `systemctl` and `journalctl` with an
  allowlisted subcommand set, 30-second timeouts, and `Executor` interface
  for unit-testing without real system binaries.
- `internal/platform/launchd`: wraps `launchctl` with the same safety
  constraints; `GetLogs` returns `ErrNotSupported`.
- `internal/security`: `ValidateServiceName`, `ValidateConfirmFlag`, and
  `ValidateLogLines` with typed `ValidationError`.
- `internal/audit`: structured JSON audit logger writing newline-delimited
  entries to an `io.Writer`.
- `internal/runtime`: bridges the plugin registry to the `mcp-go` stdio
  transport; errors surface as `isError: true` so MCP clients can
  self-correct without a protocol-level failure.
- `cmd/service-control-mcp`: entry point that detects the platform adapter,
  wires the audit logger, and registers all nine tools.
- Read tools: `list_services`, `get_service_status`, `describe_service`,
  `validate_service_name`, `get_service_logs`.
- Write tools: `start_service`, `stop_service` (`confirm: true` required),
  `restart_service` (`confirm: true` required), `reload_service`.

[Unreleased]: https://github.com/ipolyzos/service-control-mcp/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/ipolyzos/service-control-mcp/releases/tag/v0.1.0
