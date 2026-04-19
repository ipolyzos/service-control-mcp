package plugin_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ipolyzos/service-control-mcp/internal/plugin"
)

// fakePlugin is a minimal Tool implementation used only in tests.
type fakePlugin struct {
	name   string
	schema string
}

func (f *fakePlugin) Name() string                 { return f.name }
func (f *fakePlugin) Description() string          { return "fake tool for testing" }
func (f *fakePlugin) InputSchema() json.RawMessage { return json.RawMessage(f.schema) }
func (f *fakePlugin) Execute(_ context.Context, _ map[string]any) (any, error) {
	return map[string]any{"ok": true}, nil
}

func TestRegistry_RegisterAndLookup(t *testing.T) {
	t.Parallel()

	reg := plugin.NewRegistry()
	p := &fakePlugin{name: "my_tool", schema: `{"type":"object"}`}

	if err := reg.Register(p); err != nil {
		t.Fatalf("Register: unexpected error: %v", err)
	}

	got, err := reg.Lookup("my_tool")
	if err != nil {
		t.Fatalf("Lookup: unexpected error: %v", err)
	}
	if got.Name() != "my_tool" {
		t.Errorf("Lookup returned wrong tool: got %q, want %q", got.Name(), "my_tool")
	}
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	t.Parallel()

	reg := plugin.NewRegistry()
	p := &fakePlugin{name: "dup_tool", schema: `{}`}

	if err := reg.Register(p); err != nil {
		t.Fatalf("first Register: unexpected error: %v", err)
	}
	if err := reg.Register(p); err == nil {
		t.Fatal("second Register: expected error for duplicate name, got nil")
	}
}

func TestRegistry_LookupUnknown(t *testing.T) {
	t.Parallel()

	reg := plugin.NewRegistry()
	_, err := reg.Lookup("nonexistent")
	if err == nil {
		t.Fatal("Lookup: expected error for unknown tool, got nil")
	}
	if !errors.Is(err, plugin.ErrToolNotFound) {
		t.Errorf("Lookup: expected ErrToolNotFound, got %v", err)
	}
}

func TestRegistry_List(t *testing.T) {
	t.Parallel()

	reg := plugin.NewRegistry()
	tools := []string{"alpha", "beta", "gamma"}
	for _, name := range tools {
		if err := reg.Register(&fakePlugin{name: name, schema: `{}`}); err != nil {
			t.Fatalf("Register %q: %v", name, err)
		}
	}

	listed := reg.List()
	if len(listed) != len(tools) {
		t.Fatalf("List: got %d tools, want %d", len(listed), len(tools))
	}
}

func TestRegistry_RegisterNilPanics(t *testing.T) {
	t.Parallel()

	reg := plugin.NewRegistry()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic registering nil Tool, got none")
		}
	}()
	_ = reg.Register(nil) //nolint:errcheck
}
