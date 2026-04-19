package platform_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ipolyzos/service-control-mcp/internal/platform"
)

// fakeManager is a no-op Manager used only in tests.
type fakeManager struct{ name string }

func (f *fakeManager) Name() string { return f.name }
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
	return nil, nil
}
func (f *fakeManager) Start(_ context.Context, _ string) error   { return nil }
func (f *fakeManager) Stop(_ context.Context, _ string) error    { return nil }
func (f *fakeManager) Restart(_ context.Context, _ string) error { return nil }
func (f *fakeManager) Reload(_ context.Context, _ string) error  { return nil }

func TestServiceInfoFields(t *testing.T) {
	t.Parallel()

	info := platform.ServiceInfo{
		Name:    "nginx",
		Manager: "systemd",
		State:   platform.StateRunning,
	}
	if info.Name != "nginx" {
		t.Errorf("Name: got %q, want %q", info.Name, "nginx")
	}
	if info.Manager != "systemd" {
		t.Errorf("Manager: got %q, want %q", info.Manager, "systemd")
	}
	if info.State != platform.StateRunning {
		t.Errorf("State: got %q, want %q", info.State, platform.StateRunning)
	}
}

func TestServiceState_String(t *testing.T) {
	t.Parallel()

	cases := []struct {
		state platform.ServiceState
		want  string
	}{
		{platform.StateRunning, "running"},
		{platform.StateStopped, "stopped"},
		{platform.StateFailed, "failed"},
		{platform.StateUnknown, "unknown"},
	}
	for _, tc := range cases {
		if string(tc.state) != tc.want {
			t.Errorf("ServiceState %v: got %q, want %q", tc.state, string(tc.state), tc.want)
		}
	}
}

func TestErrUnsupportedPlatform(t *testing.T) {
	t.Parallel()

	err := platform.ErrUnsupportedPlatform("freebsd")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	var unsupported *platform.UnsupportedPlatformError
	if !errors.As(err, &unsupported) {
		t.Errorf("expected UnsupportedPlatformError, got %T", err)
	}
	if unsupported.OS != "freebsd" {
		t.Errorf("OS: got %q, want %q", unsupported.OS, "freebsd")
	}
}

func TestFakeManager_ImplementsInterface(t *testing.T) {
	t.Parallel()

	// Compile-time assertion that fakeManager satisfies Manager.
	var _ platform.Manager = (*fakeManager)(nil)
}
