# Security Policy

## Supported Versions

Only the latest tagged release on `main` receives security fixes.

| Version | Supported |
|---------|-----------|
| latest  | ✓         |
| older   | ✗         |

## Reporting a Vulnerability

**Do not open a public GitHub issue for security vulnerabilities.**

Please report vulnerabilities by emailing **git@ipolyzos.com** with the
subject line `[security-control-mcp] Security Vulnerability`.

Include:

- A description of the vulnerability and its potential impact.
- Steps to reproduce or a proof-of-concept.
- Affected version(s).
- Any suggested mitigations you are aware of.

You will receive an acknowledgement within 72 hours. We aim to triage,
develop a fix, and publish a patched release within 30 days for critical
issues and within 90 days for non-critical issues.

## Security Design

`service-control-mcp` is designed with the following security invariants:

- **No privilege escalation.** The server operates entirely under the
  permissions of the invoking user. It never calls `sudo`, `su`, or any
  setuid mechanism.
- **Command allowlist only.** Platform adapters may only invoke
  pre-approved commands (`systemctl`, `launchctl`) with pre-approved
  subcommands. No shell interpolation or user-controlled arguments are
  passed to the operating system.
- **Strict service name validation.** Service identifiers are validated
  against a restricted character set before any platform invocation.
- **Bounded timeouts.** All external command invocations have a
  hard-coded maximum execution time; no operation can block indefinitely.
- **Structured audit logging.** Every write operation is recorded with a
  unique `audit_id`, timestamp, action, service name, user context, and
  outcome.
- **Explicit confirmation for destructive operations.** `stop_service`
  requires `confirm: true` in the tool input; the server rejects the
  call otherwise.
