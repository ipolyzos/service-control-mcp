package list_services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ipolyzos/service-control-mcp/internal/platform"
	plugin "github.com/ipolyzos/service-control-mcp/plugins/list_services"
)

type fakeManager struct {
	err      error
	services []platform.ServiceInfo
}

func (f *fakeManager) Name() string { return "fake" }
func (f *fakeManager) ListServices(_ context.Context) ([]platform.ServiceInfo, error) {
	return f.services, f.err
}
func (f *fakeManager) GetStatus(_ context.Context, _ string) (platform.ServiceStatus, error) {
	return platform.ServiceStatus{}, nil
}
func (f *fakeManager) Describe(_ context.Context, _ string) (platform.ServiceDetail, error) {
	return platform.ServiceDetail{}, nil
}
func (f *fakeManager) GetLogs(_ context.Context, _ string, _ int) ([]string, error) {
	return nil, nil
}
func (f *fakeManager) Start(_ context.Context, _ string) error   { return nil }
func (f *fakeManager) Stop(_ context.Context, _ string) error    { return nil }
func (f *fakeManager) Restart(_ context.Context, _ string) error { return nil }
func (f *fakeManager) Reload(_ context.Context, _ string) error  { return nil }

func TestPlugin_Metadata(t *testing.T) {
	t.Parallel()

	p := plugin.New(&fakeManager{})
	if p.Name() != "list_services" {
		t.Errorf("Name: got %q, want %q", p.Name(), "list_services")
	}
	if p.Description() == "" {
		t.Error("Description must not be empty")
	}
	if len(p.InputSchema()) == 0 {
		t.Error("InputSchema must not be empty")
	}
}

func TestPlugin_Execute_Success(t *testing.T) {
	t.Parallel()

	mgr := &fakeManager{
		services: []platform.ServiceInfo{
			{Name: "nginx.service", Manager: "fake", State: platform.StateRunning},
			{Name: "sshd.service", Manager: "fake", State: platform.StateStopped},
		},
	}
	p := plugin.New(mgr)

	result, err := p.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("Execute: unexpected error: %v", err)
	}
	out, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("Execute: result type %T, want map[string]any", result)
	}
	count, ok := out["count"].(int)
	if !ok || count != 2 {
		t.Errorf("count: got %v, want 2", out["count"])
	}
}

func TestPlugin_Execute_ManagerError(t *testing.T) {
	t.Parallel()

	mgr := &fakeManager{err: errors.New("dbus unavailable")}
	p := plugin.New(mgr)

	_, err := p.Execute(context.Background(), map[string]any{})
	if err == nil {
		t.Fatal("Execute: expected error from manager, got nil")
	}
}

func TestPlugin_Execute_EmptyList(t *testing.T) {
	t.Parallel()

	p := plugin.New(&fakeManager{services: []platform.ServiceInfo{}})

	result, err := p.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("Execute: unexpected error: %v", err)
	}
	out := result.(map[string]any)
	if out["count"].(int) != 0 {
		t.Errorf("count: got %v, want 0", out["count"])
	}
}
