// Package security provides input validation for all MCP tool arguments.
// Every platform adapter invocation must pass through this package first.
package security

import (
	"fmt"
	"regexp"
)

const (
	maxServiceNameLen = 128
	maxLogLines       = 500
)

// serviceNamePattern matches valid service name characters.
// Allows alphanumeric, hyphen, underscore, dot, at-sign, and colon.
// This covers common systemd unit names (e.g. "dbus-broker.service",
// "getty@tty1.service") and launchd labels (e.g. "com.apple.logd").
var serviceNamePattern = regexp.MustCompile(`^[a-zA-Z0-9._@:-]+$`)

// ValidationError is returned when an input fails a validation rule.
type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: field %q: %s", e.Field, e.Reason)
}

// ValidateServiceName checks that name is a safe, well-formed service
// identifier before it is passed to any platform adapter.
//
// Rules:
//   - 1 to 128 characters
//   - Only [a-zA-Z0-9._@:-] — no whitespace, shell metacharacters, or path separators
//   - No path traversal sequences (must not contain / or ..)
func ValidateServiceName(name string) error {
	if len(name) == 0 {
		return &ValidationError{Field: "service_name", Reason: "must not be empty"}
	}
	if len(name) > maxServiceNameLen {
		return &ValidationError{
			Field:  "service_name",
			Reason: fmt.Sprintf("must not exceed %d characters, got %d", maxServiceNameLen, len(name)),
		}
	}
	if !serviceNamePattern.MatchString(name) {
		return &ValidationError{
			Field:  "service_name",
			Reason: "contains invalid characters; only [a-zA-Z0-9._@:-] are permitted",
		}
	}
	return nil
}

// ValidateConfirmFlag enforces that destructive operations receive an
// explicit confirm: true before proceeding.
func ValidateConfirmFlag(confirm bool) error {
	if !confirm {
		return &ValidationError{
			Field:  "confirm",
			Reason: "must be true to execute this destructive operation",
		}
	}
	return nil
}

// ValidateLogLines checks that lines is a positive integer within the
// maximum bounded range.
func ValidateLogLines(lines int) error {
	if lines <= 0 {
		return &ValidationError{
			Field:  "lines",
			Reason: "must be a positive integer",
		}
	}
	if lines > maxLogLines {
		return &ValidationError{
			Field:  "lines",
			Reason: fmt.Sprintf("must not exceed %d", maxLogLines),
		}
	}
	return nil
}
