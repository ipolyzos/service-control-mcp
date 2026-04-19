package launchd_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ipolyzos/service-control-mcp/internal/platform"
	"github.com/ipolyzos/service-control-mcp/internal/platform/launchd"
)

type fakeResponse struct {
	err    error
	stdout string
	stderr string
}

type fakeExec struct {
	responses map[string]fakeResponse
}

func (f *fakeExec) Run(_ context.Context, args []string) (string, string, error) {
	key := strings.Join(args, " ")
	if resp, ok := f.responses[key]; ok {
		return resp.stdout, resp.stderr, resp.err
	}
	return "", "", errors.New("fakeExec: no canned response for: " + key)
}

func newFake(responses map[string]fakeResponse) *fakeExec {
	return &fakeExec{responses: responses}
}

// ---- Name ----

func TestName(t *testing.T) {
	t.Parallel()

	mgr := launchd.NewWithExecutor(newFake(nil))
	if mgr.Name() != "launchd" {
		t.Errorf("Name: got %q, want %q", mgr.Name(), "launchd")
	}
}

// ---- GetStatus ----

func TestGetStatus_Running(t *testing.T) {
	t.Parallel()

	// launchctl print output for a running service
	output := `{
	"com.apple.logd" = {
		"pid" = 42;
		"state" = "running";
	};
}`
	exec := newFake(map[string]fakeResponse{
		"print system/com.apple.logd": {stdout: output},
	})
	mgr := launchd.NewWithExecutor(exec)

	status, err := mgr.GetStatus(context.Background(), "com.apple.logd")
	if err != nil {
		t.Fatalf("GetStatus: unexpected error: %v", err)
	}
	if status.Name != "com.apple.logd" {
		t.Errorf("Name: got %q, want %q", status.Name, "com.apple.logd")
	}
	if status.State != platform.StateRunning {
		t.Errorf("State: got %q, want %q", status.State, platform.StateRunning)
	}
	if status.Manager != "launchd" {
		t.Errorf("Manager: got %q, want %q", status.Manager, "launchd")
	}
}

func TestGetStatus_NotRunning(t *testing.T) {
	t.Parallel()

	output := `{
	"com.apple.logd" = {
		"pid" = -1;
		"state" = "not running";
	};
}`
	exec := newFake(map[string]fakeResponse{
		"print system/com.apple.logd": {stdout: output},
	})
	mgr := launchd.NewWithExecutor(exec)

	status, err := mgr.GetStatus(context.Background(), "com.apple.logd")
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if status.State != platform.StateStopped {
		t.Errorf("State: got %q, want %q", status.State, platform.StateStopped)
	}
}

func TestGetStatus_ExecError(t *testing.T) {
	t.Parallel()

	exec := newFake(map[string]fakeResponse{
		"print system/missing.service": {err: errors.New("service not found")},
	})
	mgr := launchd.NewWithExecutor(exec)

	_, err := mgr.GetStatus(context.Background(), "missing.service")
	if err == nil {
		t.Fatal("GetStatus: expected error, got nil")
	}
}

// ---- ListServices ----

func TestListServices_ParsesList(t *testing.T) {
	t.Parallel()

	// launchctl list output: PID, ExitCode, Label
	output := `42	0	com.apple.logd
-	0	com.apple.distnoted
99	0	com.apple.metadata.mds
`
	exec := newFake(map[string]fakeResponse{
		"list": {stdout: output},
	})
	mgr := launchd.NewWithExecutor(exec)

	services, err := mgr.ListServices(context.Background())
	if err != nil {
		t.Fatalf("ListServices: %v", err)
	}
	if len(services) != 3 {
		t.Fatalf("ListServices: got %d services, want 3", len(services))
	}
	var logd *platform.ServiceInfo
	for i := range services {
		if services[i].Name == "com.apple.logd" {
			logd = &services[i]
		}
	}
	if logd == nil {
		t.Fatal("com.apple.logd not found")
	}
	if logd.State != platform.StateRunning {
		t.Errorf("logd state: got %q, want %q", logd.State, platform.StateRunning)
	}
}

func TestListServices_Empty(t *testing.T) {
	t.Parallel()

	exec := newFake(map[string]fakeResponse{
		"list": {stdout: ""},
	})
	mgr := launchd.NewWithExecutor(exec)

	services, err := mgr.ListServices(context.Background())
	if err != nil {
		t.Fatalf("ListServices: %v", err)
	}
	if len(services) != 0 {
		t.Errorf("expected 0 services, got %d", len(services))
	}
}

func TestListServices_ExecError(t *testing.T) {
	t.Parallel()

	exec := newFake(map[string]fakeResponse{
		"list": {err: errors.New("launchctl error")},
	})
	mgr := launchd.NewWithExecutor(exec)

	_, err := mgr.ListServices(context.Background())
	if err == nil {
		t.Fatal("ListServices: expected error, got nil")
	}
}

func TestListServices_ShortLines(t *testing.T) {
	t.Parallel()

	output := "too-short\n42\t0\tcom.apple.logd\n"
	exec := newFake(map[string]fakeResponse{
		"list": {stdout: output},
	})
	mgr := launchd.NewWithExecutor(exec)

	services, err := mgr.ListServices(context.Background())
	if err != nil {
		t.Fatalf("ListServices: %v", err)
	}
	// short line is skipped; only the valid line counts
	if len(services) != 1 {
		t.Errorf("expected 1 service, got %d", len(services))
	}
}

// ---- Describe ----

func TestDescribe_Success(t *testing.T) {
	t.Parallel()

	output := `{
	"com.apple.logd" = {
		"pid" = 42;
		"state" = "running";
	};
}`
	exec := newFake(map[string]fakeResponse{
		"print system/com.apple.logd": {stdout: output},
	})
	mgr := launchd.NewWithExecutor(exec)

	detail, err := mgr.Describe(context.Background(), "com.apple.logd")
	if err != nil {
		t.Fatalf("Describe: %v", err)
	}
	if detail.Name != "com.apple.logd" {
		t.Errorf("Name: got %q, want %q", detail.Name, "com.apple.logd")
	}
	if detail.Manager != "launchd" {
		t.Errorf("Manager: got %q, want %q", detail.Manager, "launchd")
	}
}

func TestDescribe_Error(t *testing.T) {
	t.Parallel()

	exec := newFake(map[string]fakeResponse{
		"print system/missing": {err: errors.New("not found")},
	})
	mgr := launchd.NewWithExecutor(exec)

	_, err := mgr.Describe(context.Background(), "missing")
	if err == nil {
		t.Fatal("Describe: expected error, got nil")
	}
}

// ---- Write operations ----

func TestStart_Success(t *testing.T) {
	t.Parallel()

	exec := newFake(map[string]fakeResponse{
		"bootstrap system /Library/LaunchDaemons/com.example.svc.plist": {},
	})
	mgr := launchd.NewWithExecutor(exec)

	if err := mgr.Start(context.Background(), "com.example.svc"); err != nil {
		t.Fatalf("Start: unexpected error: %v", err)
	}
}

func TestStop_Success(t *testing.T) {
	t.Parallel()

	exec := newFake(map[string]fakeResponse{
		"bootout system/com.example.svc": {},
	})
	mgr := launchd.NewWithExecutor(exec)

	if err := mgr.Stop(context.Background(), "com.example.svc"); err != nil {
		t.Fatalf("Stop: unexpected error: %v", err)
	}
}

func TestRestart_Success(t *testing.T) {
	t.Parallel()

	exec := newFake(map[string]fakeResponse{
		"kickstart -k system/com.example.svc": {},
	})
	mgr := launchd.NewWithExecutor(exec)

	if err := mgr.Restart(context.Background(), "com.example.svc"); err != nil {
		t.Fatalf("Restart: unexpected error: %v", err)
	}
}

func TestReload_Success(t *testing.T) {
	t.Parallel()

	exec := newFake(map[string]fakeResponse{
		"kickstart -k system/com.example.svc": {},
	})
	mgr := launchd.NewWithExecutor(exec)

	if err := mgr.Reload(context.Background(), "com.example.svc"); err != nil {
		t.Fatalf("Reload: unexpected error: %v", err)
	}
}

func TestStart_Error(t *testing.T) {
	t.Parallel()

	exec := newFake(map[string]fakeResponse{
		"bootstrap system /Library/LaunchDaemons/com.example.svc.plist": {err: errors.New("already loaded")},
	})
	mgr := launchd.NewWithExecutor(exec)

	if err := mgr.Start(context.Background(), "com.example.svc"); err == nil {
		t.Fatal("Start: expected error, got nil")
	}
}

func TestStop_Error(t *testing.T) {
	t.Parallel()

	exec := newFake(map[string]fakeResponse{
		"bootout system/com.example.svc": {err: errors.New("not loaded")},
	})
	mgr := launchd.NewWithExecutor(exec)

	if err := mgr.Stop(context.Background(), "com.example.svc"); err == nil {
		t.Fatal("Stop: expected error, got nil")
	}
}

// ---- GetLogs ----

func TestGetLogs_NotSupported(t *testing.T) {
	t.Parallel()

	mgr := launchd.NewWithExecutor(newFake(nil))
	_, err := mgr.GetLogs(context.Background(), "com.apple.logd", 50)
	if err == nil {
		t.Fatal("GetLogs: expected not-supported error, got nil")
	}
}

// ---- platform registration ----

func TestPlatformRegistration_Darwin(t *testing.T) {
	t.Parallel()

	mgr, err := platform.DetectFor("darwin")
	if err != nil {
		t.Fatalf("DetectFor(darwin): %v", err)
	}
	if mgr.Name() != "launchd" {
		t.Errorf("Name: got %q, want %q", mgr.Name(), "launchd")
	}
}
