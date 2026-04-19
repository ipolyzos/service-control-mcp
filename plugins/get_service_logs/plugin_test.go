package get_service_logs_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ipolyzos/service-control-mcp/internal/platform"
	plugin "github.com/ipolyzos/service-control-mcp/plugins/get_service_logs"
)

type fakeManager struct {
	err  error
	logs []string
}

func (f *fakeManager) Name() string { return "fake" }
func (f *fakeManager) ListServices(_ context.Context) ([]platform.ServiceInfo, error) {
	return nil, nil
}
func (f *fakeManager) GetStatus(_ context.Context, _ string) (platform.ServiceStatus, error) {
	return platform.ServiceStatus{}, nil
}
func (f *fakeManager) Describe(_ context.Context, _ string) (platform.ServiceDetail, error) {
	return platform.ServiceDetail{}, nil
}
func (f *fakeManager) GetLogs(_ context.Context, _ string, _ int) ([]string, error) {
	return f.logs, f.err
}
func (f *fakeManager) Start(_ context.Context, _ string) error   { return nil }
func (f *fakeManager) Stop(_ context.Context, _ string) error    { return nil }
func (f *fakeManager) Restart(_ context.Context, _ string) error { return nil }
func (f *fakeManager) Reload(_ context.Context, _ string) error  { return nil }

func TestPlugin_Metadata(t *testing.T) {
	t.Parallel()

	p := plugin.New(&fakeManager{})
	if p.Name() != "get_service_logs" {
		t.Errorf("Name: got %q, want %q", p.Name(), "get_service_logs")
	}
	if p.Description() == "" {
		t.Error("Description must not be empty")
	}
}

func TestPlugin_Execute_Success(t *testing.T) {
	t.Parallel()

	mgr := &fakeManager{
		logs: []string{
			"Apr 01 10:00:00 nginx[1]: started",
			"Apr 01 10:00:01 nginx[1]: listening",
		},
	}
	p := plugin.New(mgr)

	result, err := p.Execute(context.Background(), map[string]any{
		"service_name": "nginx.service",
		"lines":        float64(50),
	})
	if err != nil {
		t.Fatalf("Execute: unexpected error: %v", err)
	}
	out, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("result type %T, want map[string]any", result)
	}
	if out["line_count"].(int) != 2 {
		t.Errorf("line_count: got %v, want 2", out["line_count"])
	}
}

func TestPlugin_Execute_DefaultLines(t *testing.T) {
	t.Parallel()

	mgr := &fakeManager{logs: []string{"line1"}}
	p := plugin.New(mgr)

	// lines not provided — should use a default
	result, err := p.Execute(context.Background(), map[string]any{
		"service_name": "nginx.service",
	})
	if err != nil {
		t.Fatalf("Execute with no lines: unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("result must not be nil")
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
		"service_name": "bad name!",
	})
	if err == nil {
		t.Fatal("Execute: expected validation error, got nil")
	}
}

func TestPlugin_Execute_InvalidLines(t *testing.T) {
	t.Parallel()

	p := plugin.New(&fakeManager{})
	_, err := p.Execute(context.Background(), map[string]any{
		"service_name": "nginx.service",
		"lines":        float64(0),
	})
	if err == nil {
		t.Fatal("Execute: expected error for lines=0, got nil")
	}
}

func TestPlugin_Execute_ManagerError(t *testing.T) {
	t.Parallel()

	mgr := &fakeManager{err: errors.New("not supported")}
	p := plugin.New(mgr)

	_, err := p.Execute(context.Background(), map[string]any{
		"service_name": "nginx.service",
	})
	if err == nil {
		t.Fatal("Execute: expected error from manager, got nil")
	}
}
