// Package audit provides structured JSON audit logging for all write
// operations. Entries are written to a supplied io.Writer (typically stderr)
// so they do not interfere with the stdio MCP transport channel.
package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// Outcome describes whether a write operation succeeded or failed.
type Outcome string

// Outcome values for structured audit entries.
const (
	OutcomeSuccess Outcome = "success"
	OutcomeFailure Outcome = "failure"
)

// Entry is one structured audit record.
type Entry struct {
	AuditID     string    `json:"audit_id"`
	Timestamp   time.Time `json:"timestamp"`
	Action      string    `json:"action"`
	ServiceName string    `json:"service_name"`
	Manager     string    `json:"manager"`
	Outcome     Outcome   `json:"outcome"`
	Error       string    `json:"error,omitempty"`
}

// Logger writes audit entries as newline-delimited JSON to an io.Writer.
type Logger struct {
	w   io.Writer
	enc *json.Encoder
}

// NewLogger creates a Logger that writes to w.
func NewLogger(w io.Writer) *Logger {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return &Logger{w: w, enc: enc}
}

// Write encodes e as a single JSON object followed by a newline.
func (l *Logger) Write(e Entry) error {
	if err := l.enc.Encode(e); err != nil {
		return fmt.Errorf("audit: encode entry: %w", err)
	}
	return nil
}
