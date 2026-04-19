package stop_service_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ipolyzos/service-control-mcp/internal/audit"
	plugin "github.com/ipolyzos/service-control-mcp/plugins/stop_service"
)

type fakeManager struct {
	err  error
	name string
}

func (f *fakeManager) Name() string { return f.name }
func (f *fakeManager) Stop(_ context.Context, _ string) error {
	return f.err
}

func newPlugin(err error) (*plugin.Plugin, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	log := audit.NewLogger(buf)
	mgr := &fakeManager{name: "fake", err: err}
	return plugin.New(mgr, log), buf
}

func TestPlugin_Metadata(t *testing.T) {
	t.Parallel()
	p, _ := newPlugin(nil)
	if p.Name() != "stop_service" {
		t.Errorf("Name: got %q, want %q", p.Name(), "stop_service")
	}
	if p.Description() == "" {
		t.Error("Description must not be empty")
	}
}

func TestPlugin_Execute_Success(t *testing.T) {
	t.Parallel()
	p, buf := newPlugin(nil)

	result, err := p.Execute(context.Background(), map[string]any{
		"service_name": "nginx.service",
		"confirm":      true,
	})
	if err != nil {
		t.Fatalf("Execute: unexpected error: %v", err)
	}
	out, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("result type %T, want map[string]any", result)
	}
	if out["outcome"] != "success" {
		t.Errorf("outcome: got %v, want success", out["outcome"])
	}
	if !strings.Contains(buf.String(), `"outcome":"success"`) {
		t.Errorf("audit log missing success entry: %s", buf.String())
	}
}

func TestPlugin_Execute_MissingConfirm(t *testing.T) {
	t.Parallel()
	p, _ := newPlugin(nil)

	_, err := p.Execute(context.Background(), map[string]any{
		"service_name": "nginx.service",
	})
	if err == nil {
		t.Fatal("Execute: expected error when confirm absent")
	}
}

func TestPlugin_Execute_ConfirmFalse(t *testing.T) {
	t.Parallel()
	p, _ := newPlugin(nil)

	_, err := p.Execute(context.Background(), map[string]any{
		"service_name": "nginx.service",
		"confirm":      false,
	})
	if err == nil {
		t.Fatal("Execute: expected error when confirm=false")
	}
}

func TestPlugin_Execute_MissingServiceName(t *testing.T) {
	t.Parallel()
	p, _ := newPlugin(nil)

	_, err := p.Execute(context.Background(), map[string]any{
		"confirm": true,
	})
	if err == nil {
		t.Fatal("Execute: expected error for missing service_name")
	}
}

func TestPlugin_Execute_InvalidServiceName(t *testing.T) {
	t.Parallel()
	p, _ := newPlugin(nil)

	_, err := p.Execute(context.Background(), map[string]any{
		"service_name": "bad name!",
		"confirm":      true,
	})
	if err == nil {
		t.Fatal("Execute: expected validation error for invalid name")
	}
}

func TestPlugin_Execute_ManagerError(t *testing.T) {
	t.Parallel()
	p, buf := newPlugin(errors.New("unit not found"))

	_, err := p.Execute(context.Background(), map[string]any{
		"service_name": "nginx.service",
		"confirm":      true,
	})
	if err == nil {
		t.Fatal("Execute: expected error from manager")
	}
	if !strings.Contains(buf.String(), `"outcome":"failure"`) {
		t.Errorf("audit log missing failure entry: %s", buf.String())
	}
}
