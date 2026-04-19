# Security Model

## Threat model

`service-control-mcp` is a local stdio server invoked by an MCP client
(typically an MCP client or IDE extension) running as the same user.
It is **not** a network service and is **not** designed to be exposed
over a socket or HTTP.

Threat actors in scope:
- A malicious or buggy MCP client supplying unexpected input
- Prompt injection through service names or arguments

Threat actors out of scope:
- Network attackers (no listener is opened)
- Privilege escalation from a lower-privileged caller (the server runs as
  the invoking user; no setuid, no sudo)

## Invariants

### No privilege escalation

The server never calls `sudo`, `su`, `pkexec`, or any SUID binary.
Operations are bounded by what the invoking user is already allowed to do
on the host service manager.

### No shell execution

Platform adapters call `exec.CommandContext` with a fixed binary path
(`/usr/bin/systemctl`, `/bin/launchctl`) and a pre-constructed argument
list. `/bin/sh -c` is never used. There is no string interpolation into
shell commands.

### Input validation before every platform call

Every tool validates the `service_name` argument through
`security.ValidateServiceName` before any platform adapter is invoked.
The validator enforces:

- Length 1–128 characters
- Character set: `[a-zA-Z0-9._@:-]` plus the `.service` and similar suffixes
- No path traversal sequences (`..`, `/`)
- No shell metacharacters

Tools that accept a `lines` argument enforce the range 1–500 via
`security.ValidateLogLines`.

### Allowlisted arguments

Each platform adapter defines an explicit allowlist of subcommands it
will pass to the underlying binary. Any request for an operation outside
the allowlist is rejected with an error before any subprocess is spawned.

### Bounded execution timeouts

Every subprocess spawned by a platform adapter uses a
`context.WithTimeout` with a 30-second ceiling. A timed-out operation
returns an error; no goroutine or process is leaked.

### Fail closed

The server returns a typed error for:
- Unsupported platforms
- Unsupported managers
- Invalid service names
- Ambiguous or missing arguments
- Timeout conditions

No operation is attempted when input cannot be validated.

### Confirmation gate for destructive operations

`stop_service` and `restart_service` require the caller to pass
`confirm: true` explicitly. The absence of this flag or a value of
`false` causes the tool to return an error without invoking any platform
command.

### Structured audit trail

Every write operation (`start_service`, `stop_service`, `restart_service`,
`reload_service`) records a JSON audit entry to stderr **regardless of
outcome**. A failed operation is audited the same way as a successful one.

Audit entry fields:
```json
{
  "audit_id":    "<16-byte random hex>",
  "timestamp":   "<RFC3339Nano>",
  "action":      "<tool_name>",
  "service_name":"<validated name>",
  "manager":     "<systemd|launchd>",
  "outcome":     "<success|failure>",
  "error":       "<empty or error message>"
}
```

The audit ID is generated using `crypto/rand`; it is not a sequential
counter and cannot be predicted by a caller.

## Reporting vulnerabilities

See [SECURITY.md](../SECURITY.md) for the responsible disclosure process.
