package audit_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ipolyzos/service-control-mcp/internal/audit"
)

func TestEntry_RequiredFields(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	e := audit.Entry{
		AuditID:     "test-id",
		Timestamp:   now,
		Action:      "stop_service",
		ServiceName: "nginx",
		Manager:     "systemd",
		Outcome:     audit.OutcomeSuccess,
	}
	if e.AuditID == "" {
		t.Error("AuditID must not be empty")
	}
	if e.Action == "" {
		t.Error("Action must not be empty")
	}
	if e.ServiceName == "" {
		t.Error("ServiceName must not be empty")
	}
	if e.Manager == "" {
		t.Error("Manager must not be empty")
	}
	if e.Outcome == "" {
		t.Error("Outcome must not be empty")
	}
	if e.Timestamp.IsZero() {
		t.Error("Timestamp must not be zero")
	}
}

func TestLogger_WriteEntry_JSON(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := audit.NewLogger(&buf)

	e := audit.Entry{
		AuditID:     "abc-123",
		Timestamp:   time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC),
		Action:      "restart_service",
		ServiceName: "caddy",
		Manager:     "systemd",
		Outcome:     audit.OutcomeSuccess,
	}

	if err := logger.Write(e); err != nil {
		t.Fatalf("Write: unexpected error: %v", err)
	}

	var decoded audit.Entry
	if err := json.NewDecoder(&buf).Decode(&decoded); err != nil {
		t.Fatalf("JSON decode: %v — output was: %s", err, buf.String())
	}
	if decoded.AuditID != e.AuditID {
		t.Errorf("AuditID: got %q, want %q", decoded.AuditID, e.AuditID)
	}
	if decoded.Action != e.Action {
		t.Errorf("Action: got %q, want %q", decoded.Action, e.Action)
	}
	if decoded.Outcome != e.Outcome {
		t.Errorf("Outcome: got %q, want %q", decoded.Outcome, e.Outcome)
	}
}

func TestLogger_WriteEntry_IncludesNewline(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := audit.NewLogger(&buf)

	e := audit.Entry{
		AuditID:     "x",
		Timestamp:   time.Now().UTC(),
		Action:      "start_service",
		ServiceName: "sshd",
		Manager:     "systemd",
		Outcome:     audit.OutcomeSuccess,
	}
	if err := logger.Write(e); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if !strings.HasSuffix(buf.String(), "\n") {
		t.Error("expected log entry to end with newline")
	}
}

func TestLogger_WriteEntry_ErrorPropagated(t *testing.T) {
	t.Parallel()

	// failWriter always returns an error on Write.
	w := &failWriter{}
	logger := audit.NewLogger(w)

	e := audit.Entry{
		AuditID:     "err-test",
		Timestamp:   time.Now().UTC(),
		Action:      "stop_service",
		ServiceName: "redis",
		Manager:     "systemd",
		Outcome:     audit.OutcomeFailure,
		Error:       "permission denied",
	}
	if err := logger.Write(e); err == nil {
		t.Fatal("Write: expected error from failing writer, got nil")
	}
}

// failWriter implements io.Writer and always returns an error.
type failWriter struct{}

func (fw *failWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestOutcomeValues(t *testing.T) {
	t.Parallel()

	if audit.OutcomeSuccess != "success" {
		t.Errorf("OutcomeSuccess: got %q, want %q", audit.OutcomeSuccess, "success")
	}
	if audit.OutcomeFailure != "failure" {
		t.Errorf("OutcomeFailure: got %q, want %q", audit.OutcomeFailure, "failure")
	}
}
