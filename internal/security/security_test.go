package security_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/ipolyzos/service-control-mcp/internal/security"
)

func TestValidateServiceName_Valid(t *testing.T) {
	t.Parallel()

	cases := []string{
		"nginx",
		"sshd",
		"my-service",
		"my_service",
		"service.socket",
		"dbus-broker.service",
		"com.apple.logd",
		"service@1",
		"a",
		strings.Repeat("a", 128),
	}
	for _, name := range cases {
		if err := security.ValidateServiceName(name); err != nil {
			t.Errorf("ValidateServiceName(%q): unexpected error: %v", name, err)
		}
	}
}

func TestValidateServiceName_Invalid(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		desc string
	}{
		{"", "empty name"},
		{strings.Repeat("a", 129), "name too long (129 chars)"},
		{"../etc/passwd", "path traversal with ../"},
		{"/etc/nginx", "absolute path with /"},
		{"service name", "contains space"},
		{"service;name", "contains semicolon"},
		{"service|name", "contains pipe"},
		{"service`cmd`", "contains backtick"},
		{"service$(cmd)", "contains $()"},
		{"service&name", "contains ampersand"},
		{"service\nname", "contains newline"},
		{"service\x00name", "contains null byte"},
	}
	for _, tc := range cases {
		err := security.ValidateServiceName(tc.name)
		if err == nil {
			t.Errorf("ValidateServiceName(%q) [%s]: expected error, got nil", tc.name, tc.desc)
			continue
		}
		var valErr *security.ValidationError
		if !errors.As(err, &valErr) {
			t.Errorf("ValidateServiceName(%q): expected *ValidationError, got %T", tc.name, err)
		}
	}
}

func TestValidationError_ErrorString(t *testing.T) {
	t.Parallel()

	e := &security.ValidationError{Field: "service_name", Reason: "too long"}
	if e.Error() == "" {
		t.Error("ValidationError.Error() must not be empty")
	}
	if !strings.Contains(e.Error(), "service_name") {
		t.Errorf("ValidationError.Error() should mention the field name, got: %s", e.Error())
	}
}

func TestValidateServiceName_BoundaryLengths(t *testing.T) {
	t.Parallel()

	// exactly 128 chars is valid
	exactly128 := strings.Repeat("x", 128)
	if err := security.ValidateServiceName(exactly128); err != nil {
		t.Errorf("name of 128 chars: unexpected error: %v", err)
	}

	// 129 chars is invalid
	exactly129 := strings.Repeat("x", 129)
	if err := security.ValidateServiceName(exactly129); err == nil {
		t.Error("name of 129 chars: expected error, got nil")
	}
}

func TestValidateConfirmFlag(t *testing.T) {
	t.Parallel()

	if err := security.ValidateConfirmFlag(true); err != nil {
		t.Errorf("ValidateConfirmFlag(true): unexpected error: %v", err)
	}
	if err := security.ValidateConfirmFlag(false); err == nil {
		t.Error("ValidateConfirmFlag(false): expected error for unconfirmed destructive op, got nil")
	}
}

func TestValidateLogLines(t *testing.T) {
	t.Parallel()

	cases := []struct {
		lines   int
		wantErr bool
	}{
		{1, false},
		{100, false},
		{500, false},
		{0, true},
		{-1, true},
		{501, true},
	}
	for _, tc := range cases {
		err := security.ValidateLogLines(tc.lines)
		if tc.wantErr && err == nil {
			t.Errorf("ValidateLogLines(%d): expected error, got nil", tc.lines)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("ValidateLogLines(%d): unexpected error: %v", tc.lines, err)
		}
	}
}
