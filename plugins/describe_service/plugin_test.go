package describe_service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ipolyzos/service-control-mcp/internal/platform"
	plugin "github.com/ipolyzos/service-control-mcp/plugins/describe_service"
)

type fakeManager struct {
	err    error
	detail platform.ServiceDetail
}

func (f *fakeManager) Name() string { return "fake" }
func (f *fakeManager) ListServices(_ context.Context) ([]platform.ServiceInfo, error) {
	return nil, nil
}
func (f *fakeManager) GetStatus(_ context.Context, _ string) (platform.ServiceStatus, error) {
	return platform.ServiceStatus{}, nil
}
func (f *fakeManager) Describe(_ context.Context, _ string) (platform.ServiceDetail, error) {
	return f.detail, f.err
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
	if p.Name() != "describe_service" {
		t.Errorf("Name: got %q, want %q", p.Name(), "describe_service")
	}
	if p.Description() == "" {
		t.Error("Description must not be empty")
	}
}

func TestPlugin_Execute_Success(t *testing.T) {
	t.Parallel()

	mgr := &fakeManager{
		detail: platform.ServiceDetail{
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
		"service_name": "service|name",
	})
	if err == nil {
		t.Fatal("Execute: expected validation error, got nil")
	}
}

func TestPlugin_Execute_ManagerError(t *testing.T) {
	t.Parallel()

	mgr := &fakeManager{err: errors.New("not found")}
	p := plugin.New(mgr)

	_, err := p.Execute(context.Background(), map[string]any{
		"service_name": "nginx.service",
	})
	if err == nil {
		t.Fatal("Execute: expected error from manager, got nil")
	}
}
