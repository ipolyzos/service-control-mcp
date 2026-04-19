package get_service_status_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ipolyzos/service-control-mcp/internal/platform"
	plugin "github.com/ipolyzos/service-control-mcp/plugins/get_service_status"
)

type fakeManager struct {
	err    error
	status platform.ServiceStatus
}

func (f *fakeManager) Name() string { return "fake" }
func (f *fakeManager) ListServices(_ context.Context) ([]platform.ServiceInfo, error) {
	return nil, nil
}
func (f *fakeManager) GetStatus(_ context.Context, _ string) (platform.ServiceStatus, error) {
	return f.status, f.err
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
	if p.Name() != "get_service_status" {
		t.Errorf("Name: got %q, want %q", p.Name(), "get_service_status")
	}
	if p.Description() == "" {
		t.Error("Description must not be empty")
	}
}

func TestPlugin_Execute_Success(t *testing.T) {
	t.Parallel()

	mgr := &fakeManager{
		status: platform.ServiceStatus{
			Name:    "nginx.service",
			Manager: "fake",
			State:   platform.StateRunning,
		},
	}
	p := plugin.New(mgr)

	result, err := p.Execute(context.Background(), map[string]any{
		"service_name": "nginx.service",
	})
	if err != nil {
		t.Fatalf("Execute: unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("Execute: result must not be nil")
	}
}

func TestPlugin_Execute_MissingServiceName(t *testing.T) {
	t.Parallel()

	p := plugin.New(&fakeManager{})

	_, err := p.Execute(context.Background(), map[string]any{})
	if err == nil {
		t.Fatal("Execute: expected error for missing service_name, got nil")
	}
}

func TestPlugin_Execute_InvalidServiceName(t *testing.T) {
	t.Parallel()

	p := plugin.New(&fakeManager{})

	_, err := p.Execute(context.Background(), map[string]any{
		"service_name": "../etc/passwd",
	})
	if err == nil {
		t.Fatal("Execute: expected error for invalid service name, got nil")
	}
}

func TestPlugin_Execute_ManagerError(t *testing.T) {
	t.Parallel()

	mgr := &fakeManager{err: errors.New("unit not found")}
	p := plugin.New(mgr)

	_, err := p.Execute(context.Background(), map[string]any{
		"service_name": "nginx.service",
	})
	if err == nil {
		t.Fatal("Execute: expected error from manager, got nil")
	}
}
