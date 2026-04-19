# Security Model

## Threat model

`service-control-mcp` is a local stdio server invoked by an MCP client
(typically an MCP client or IDE extension) running as the same OS user.
It is **not** a network service and is **not** designed to be exposed over a
socket or HTTP.

**In scope:**
- A malicious or buggy MCP client supplying crafted input (injection attempts,
  oversized payloads, unexpected argument types)
- Prompt injection: a service name or argument value derived from untrusted
  content (e.g. a log file read by an MCP client) that attempts to
  escape validation

**Out of scope:**
- Network attackers (no listener is opened)
- Privilege escalation from a lower-privileged caller — the server runs as the
  invoking user; no setuid binary, no `sudo`, no capability grants are used

## Security invariants

### No privilege escalation

The server never calls `sudo`, `su`, `pkexec`, or any SUID helper binary.
Every operation is bounded by the permissions the invoking user already holds
on the host service manager. If the user cannot stop a service manually,
neither can the MCP server.

### No shell execution

Platform adapters call `exec.CommandContext` with a fixed binary path
(`/usr/bin/systemctl`, `/bin/launchctl`) and a pre-constructed, allowlisted
argument slice. `/bin/sh -c` is never used. There is no string interpolation
into shell commands at any layer of the stack.

### Input validation before every platform call

Every tool validates the `service_name` argument through
`security.ValidateServiceName` before any platform adapter is invoked:

| Rule | Constraint |
|------|-----------|
| Length | 1–128 characters |
| Character set | `[a-zA-Z0-9._@:-]` only |
| No path traversal | Must not contain `/` or `..` |

Tools that accept a `lines` argument pass it through `security.ValidateLogLines`,
which enforces the range 1–500.

### Allowlisted subcommands

Each platform adapter maintains an explicit allowlist of subcommands it will
pass to the underlying binary. Any request for an operation outside the
allowlist is rejected with an error before any subprocess is spawned. The
adapter never constructs a command string from caller-supplied data.

### Bounded execution timeouts

Every subprocess spawned by a platform adapter uses a `context.WithTimeout`
with a 30-second ceiling. A timed-out operation returns a typed error; no
goroutine or process is leaked.

### Fail closed

The server returns a typed error — and takes no platform action — when:

- The platform is not supported (no adapter registered)
- The service name fails validation
- A required argument is missing or the wrong type
- The `confirm` flag is absent or `false` for destructive operations
- A subprocess times out

No operation is attempted on ambiguous or unvalidated input.

### Confirmation gate for destructive operations

`stop_service` and `restart_service` require the caller to pass
`confirm: true` explicitly. The check is evaluated before any platform call:

```
confirm absent  →  rejected immediately, no platform call
confirm = false →  rejected immediately, no platform call
confirm = true  →  proceed to service name validation, then platform call
```

`start_service` and `reload_service` do not require confirmation because they
are non-destructive or recoverable operations.

### Structured audit trail

Every write operation (`start_service`, `stop_service`, `restart_service`,
`reload_service`) records a JSON audit entry to stderr **regardless of
outcome**. A failed or rejected operation is audited the same way as a
successful one, so the audit trail is complete even under adversarial
conditions.

**Audit entry format:**

```json
{
  "audit_id":     "9c2ef3a1b0d74e8c2f56a3891d4b7c20",
  "timestamp":    "2024-04-01T10:05:00.123456789Z",
  "action":       "stop_service",
  "service_name": "nginx.service",
  "manager":      "systemd",
  "outcome":      "success",
  "error":        ""
}
```

| Field | Description |
|-------|-------------|
| `audit_id` | 16-byte cryptographically random hex string (via `crypto/rand`) |
| `timestamp` | RFC3339Nano UTC timestamp |
| `action` | Tool name that was invoked |
| `service_name` | Validated service name |
| `manager` | Platform adapter used (`systemd` or `launchd`) |
| `outcome` | `"success"` or `"failure"` |
| `error` | Error message on failure; empty string on success |

Audit entries are written to stderr so they do not interfere with the
stdout-based MCP transport channel. Operators can redirect stderr to a log
aggregator independently of the MCP response stream.

**Consuming audit logs:**

```sh
# Run server and capture audit trail
service-control-mcp 2>audit.log

# Stream and pretty-print audit entries in real time
service-control-mcp 2> >(jq .)

# Filter for failures only
cat audit.log | jq 'select(.outcome == "failure")'
```

## Reporting vulnerabilities

See [SECURITY.md](../SECURITY.md) for the responsible disclosure process.
