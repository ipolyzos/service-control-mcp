package validate_service_name_test

import (
	"context"
	"testing"

	plugin "github.com/ipolyzos/service-control-mcp/plugins/validate_service_name"
)

func TestPlugin_Metadata(t *testing.T) {
	t.Parallel()

	p := plugin.New()
	if p.Name() != "validate_service_name" {
		t.Errorf("Name: got %q, want %q", p.Name(), "validate_service_name")
	}
	if p.Description() == "" {
		t.Error("Description must not be empty")
	}
}

func TestPlugin_Execute_ValidName(t *testing.T) {
	t.Parallel()

	p := plugin.New()
	result, err := p.Execute(context.Background(), map[string]any{
		"service_name": "nginx.service",
	})
	if err != nil {
		t.Fatalf("Execute: unexpected error: %v", err)
	}
	out, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("result type %T, want map[string]any", result)
	}
	if out["valid"] != true {
		t.Errorf("valid: got %v, want true", out["valid"])
	}
}

func TestPlugin_Execute_InvalidName(t *testing.T) {
	t.Parallel()

	p := plugin.New()
	result, err := p.Execute(context.Background(), map[string]any{
		"service_name": "../etc/passwd",
	})
	if err != nil {
		t.Fatalf("Execute: unexpected error: %v", err)
	}
	out := result.(map[string]any)
	if out["valid"] != false {
		t.Errorf("valid: got %v, want false", out["valid"])
	}
	if out["error"] == "" {
		t.Error("error field should be non-empty for invalid name")
	}
}

func TestPlugin_Execute_MissingServiceName(t *testing.T) {
	t.Parallel()

	p := plugin.New()
	_, err := p.Execute(context.Background(), map[string]any{})
	if err == nil {
		t.Fatal("Execute: expected error for missing service_name, got nil")
	}
}

func TestPlugin_Execute_EmptyName(t *testing.T) {
	t.Parallel()

	p := plugin.New()
	result, err := p.Execute(context.Background(), map[string]any{
		"service_name": "",
	})
	if err != nil {
		t.Fatalf("Execute: unexpected error: %v", err)
	}
	out := result.(map[string]any)
	if out["valid"] != false {
		t.Errorf("valid: got %v, want false for empty name", out["valid"])
	}
}
