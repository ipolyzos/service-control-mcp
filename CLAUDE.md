# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`service-control-mcp` is a stdio MCP server (Go 1.26) that exposes
systemd (Linux) and launchd (macOS) service management as MCP tools,
with strict input validation, command allowlisting, and structured
audit logging.

## Commands

```bash
make build          # build binary to bin/service-control-mcp
make test           # go test -count=1 -race ./... + coverage profile
make lint           # golangci-lint run ./...
make fmt            # go fmt ./...
make fmt-check      # verify formatting without modifying files
make vet            # go vet ./...
make ci             # fmt-check + vet + lint + test + coverage gate (80%)
make clean          # remove bin/ and coverage.txt

# Run a single package
go test -count=1 -race ./internal/security/...

# Run a single test function
go test -count=1 -race -run TestValidateServiceName ./internal/security/...
```

## Branch and Commit Rules

- Work only on `feature/`, `fix/`, or `chore/` branches cut from `develop`.
- PRs target `develop`; releases promote to `main` via `release/<version>`.
- Never commit directly to `main` or `develop`.
- Commit titles must begin with an approved imperative verb:
  `Add`, `Fix`, `Update`, `Remove`, `Refactor`, `Implement`, `Improve`,
  `Document`, `Bump`, `Change`, `Revert`, `Simplify`, `Upgrade`, `Ensure`,
  `Support`.
- Every commit requires `Signed-off-by` (`git commit -s`).

## Architecture Summary

```
cmd/service-control-mcp/   main entry point — wires everything
internal/
  audit/                   structured JSON audit log (stderr)
  platform/                Manager interface + Detect()
    systemd/               systemctl adapter (Linux)
    launchd/               launchctl adapter (macOS)
  plugin/                  Tool interface + Registry
  runtime/                 MCP stdio server + JSON-RPC dispatch
  security/                ValidateServiceName, input sanitization
plugins/
  <tool_name>/             one package per MCP tool
docs/
  architecture.md          design rationale and package contracts
  tools/                   per-tool input/output schema docs
```

## TDD Requirement

Tests must be written **before** production code. No production code
may be committed without tests. Bug fixes must include a reproducing
test added before the fix.

## Security Invariants

- No privilege escalation; no `sudo`/`su` calls anywhere.
- Platform adapters use `exec.CommandContext` with fixed binary paths
  and pre-validated argument lists — never `sh -c`.
- `ValidateServiceName` must be called before every platform invocation.
- Write operations must be recorded via `internal/audit` with an `audit_id`.
- `stop_service` and `restart_service` must reject requests where `confirm != true`.
- All exec calls use a 30-second `context.WithTimeout` ceiling.

## Adding a Tool

1. Write `plugins/<tool_name>/plugin_test.go` first (TDD — tests before implementation).
2. Create `plugins/<tool_name>/plugin.go` implementing `plugin.Tool`.
3. Register the plugin in `cmd/service-control-mcp/main.go` (the `tools` slice in `run()`).
4. Add the tool contract to `docs/tools/<tool_name>.md`.
5. Update the tool table in `README.md`.
6. Add a CHANGELOG entry under `[Unreleased]`.
