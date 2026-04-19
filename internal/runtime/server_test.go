package runtime_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ipolyzos/service-control-mcp/internal/plugin"
	"github.com/ipolyzos/service-control-mcp/internal/runtime"
)

// fakeTool is a minimal plugin.Tool for testing the runtime layer.
type fakeTool struct {
	result any
	err    error
	name   string
}

func (f *fakeTool) Name() string        { return f.name }
func (f *fakeTool) Description() string { return "fake tool for testing" }
func (f *fakeTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{}}`)
}
func (f *fakeTool) Execute(_ context.Context, _ map[string]any) (any, error) {
	return f.result, f.err
}

func TestNew_RegistersToolsFromRegistry(t *testing.T) {
	t.Parallel()

	reg := plugin.NewRegistry()
	tools := []*fakeTool{
		{name: "tool_a"},
		{name: "tool_b"},
	}
	for _, ft := range tools {
		if err := reg.Register(ft); err != nil {
			t.Fatalf("Register: %v", err)
		}
	}

	// New must not panic and must return a non-nil Server.
	srv := runtime.New(reg, "test")
	if srv == nil {
		t.Fatal("New: returned nil server")
	}
}

func TestNew_EmptyRegistryIsValid(t *testing.T) {
	t.Parallel()

	reg := plugin.NewRegistry()
	srv := runtime.New(reg, "test")
	if srv == nil {
		t.Fatal("New: nil server for empty registry")
	}
}

func TestNew_DuplicateToolNamePanics(t *testing.T) {
	t.Parallel()

	reg := plugin.NewRegistry()
	if err := reg.Register(&fakeTool{name: "dup"}); err != nil {
		t.Fatalf("first Register: %v", err)
	}
	// Second registration into the same registry should fail at the
	// plugin.Registry level, not the runtime level.
	err := reg.Register(&fakeTool{name: "dup"})
	if err == nil {
		t.Fatal("expected error registering duplicate tool name, got nil")
	}
}

func TestServer_ToolExecuteSuccess(t *testing.T) {
	t.Parallel()

	ft := &fakeTool{
		name:   "ok_tool",
		result: map[string]any{"status": "ok"},
	}
	reg := plugin.NewRegistry()
	if err := reg.Register(ft); err != nil {
		t.Fatalf("Register: %v", err)
	}

	srv := runtime.New(reg, "test")
	result, err := srv.ExecuteTool(context.Background(), "ok_tool", map[string]any{})
	if err != nil {
		t.Fatalf("ExecuteTool: unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("ExecuteTool: result must not be nil")
	}
	if result.IsError {
		t.Errorf("ExecuteTool: IsError=true, want false")
	}
}

func TestServer_ToolExecuteToolError(t *testing.T) {
	t.Parallel()

	ft := &fakeTool{
		name: "err_tool",
		err:  errors.New("something went wrong"),
	}
	reg := plugin.NewRegistry()
	if err := reg.Register(ft); err != nil {
		t.Fatalf("Register: %v", err)
	}

	srv := runtime.New(reg, "test")
	result, err := srv.ExecuteTool(context.Background(), "err_tool", map[string]any{})
	if err != nil {
		t.Fatalf("ExecuteTool: unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("ExecuteTool: result must not be nil on tool error")
	}
	if !result.IsError {
		t.Error("ExecuteTool: IsError=false, want true on tool error")
	}
}

func TestServer_ToolExecuteUnknownTool(t *testing.T) {
	t.Parallel()

	reg := plugin.NewRegistry()
	srv := runtime.New(reg, "test")

	_, err := srv.ExecuteTool(context.Background(), "no_such_tool", map[string]any{})
	if err == nil {
		t.Fatal("ExecuteTool: expected error for unknown tool name, got nil")
	}
}
