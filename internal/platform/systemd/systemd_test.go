package systemd_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ipolyzos/service-control-mcp/internal/platform"
	"github.com/ipolyzos/service-control-mcp/internal/platform/systemd"
)

// fakeExec is an executor that returns canned output for specific arguments.
type fakeExec struct {
	responses map[string]fakeResponse
}

type fakeResponse struct {
	err    error
	stdout string
	stderr string
}

func (f *fakeExec) Run(_ context.Context, args []string) (stdout, stderr string, err error) {
	key := strings.Join(args, " ")
	if resp, ok := f.responses[key]; ok {
		return resp.stdout, resp.stderr, resp.err
	}
	return "", "", errors.New("fakeExec: no canned response for: " + key)
}

func newFake(responses map[string]fakeResponse) *fakeExec {
	return &fakeExec{responses: responses}
}

// ---- GetStatus ----

func TestGetStatus_Running(t *testing.T) {
	t.Parallel()

	exec := newFake(map[string]fakeResponse{
		"show --no-pager --property=ActiveState,SubState,LoadState,Description,MainPID,ActiveEnterTimestamp nginx.service": {
			stdout: "ActiveState=active\nSubState=running\nLoadState=loaded\nDescription=A high performance web server\nMainPID=1234\nActiveEnterTimestamp=Sat 2025-01-01 00:00:00 UTC\n",
		},
	})
	mgr := systemd.NewWithExecutor(exec)

	status, err := mgr.GetStatus(context.Background(), "nginx.service")
	if err != nil {
		t.Fatalf("GetStatus: unexpected error: %v", err)
	}
	if status.State != platform.StateRunning {
		t.Errorf("State: got %q, want %q", status.State, platform.StateRunning)
	}
	if status.SubState != "running" {
		t.Errorf("SubState: got %q, want %q", status.SubState, "running")
	}
	if status.PID != 1234 {
		t.Errorf("PID: got %d, want %d", status.PID, 1234)
	}
}

func TestGetStatus_Stopped(t *testing.T) {
	t.Parallel()

	exec := newFake(map[string]fakeResponse{
		"show --no-pager --property=ActiveState,SubState,LoadState,Description,MainPID,ActiveEnterTimestamp nginx.service": {
			stdout: "ActiveState=inactive\nSubState=dead\nLoadState=loaded\nDescription=nginx\nMainPID=0\nActiveEnterTimestamp=\n",
		},
	})
	mgr := systemd.NewWithExecutor(exec)

	status, err := mgr.GetStatus(context.Background(), "nginx.service")
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if status.State != platform.StateStopped {
		t.Errorf("State: got %q, want %q", status.State, platform.StateStopped)
	}
}

func TestGetStatus_Failed(t *testing.T) {
	t.Parallel()

	exec := newFake(map[string]fakeResponse{
		"show --no-pager --property=ActiveState,SubState,LoadState,Description,MainPID,ActiveEnterTimestamp nginx.service": {
			stdout: "ActiveState=failed\nSubState=failed\nLoadState=loaded\nDescription=nginx\nMainPID=0\nActiveEnterTimestamp=\n",
		},
	})
	mgr := systemd.NewWithExecutor(exec)

	status, err := mgr.GetStatus(context.Background(), "nginx.service")
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if status.State != platform.StateFailed {
		t.Errorf("State: got %q, want %q", status.State, platform.StateFailed)
	}
}

func TestGetStatus_ExecError(t *testing.T) {
	t.Parallel()

	exec := newFake(map[string]fakeResponse{
		"show --no-pager --property=ActiveState,SubState,LoadState,Description,MainPID,ActiveEnterTimestamp unknown.service": {
			err: errors.New("unit not found"),
		},
	})
	mgr := systemd.NewWithExecutor(exec)

	_, err := mgr.GetStatus(context.Background(), "unknown.service")
	if err == nil {
		t.Fatal("GetStatus: expected error, got nil")
	}
}

// ---- Start / Stop / Restart / Reload ----

func TestStart_Success(t *testing.T) {
	t.Parallel()

	exec := newFake(map[string]fakeResponse{
		"start nginx.service": {},
	})
	mgr := systemd.NewWithExecutor(exec)

	if err := mgr.Start(context.Background(), "nginx.service"); err != nil {
		t.Fatalf("Start: unexpected error: %v", err)
	}
}

func TestStop_Success(t *testing.T) {
	t.Parallel()

	exec := newFake(map[string]fakeResponse{
		"stop nginx.service": {},
	})
	mgr := systemd.NewWithExecutor(exec)

	if err := mgr.Stop(context.Background(), "nginx.service"); err != nil {
		t.Fatalf("Stop: unexpected error: %v", err)
	}
}

func TestRestart_Success(t *testing.T) {
	t.Parallel()

	exec := newFake(map[string]fakeResponse{
		"restart nginx.service": {},
	})
	mgr := systemd.NewWithExecutor(exec)

	if err := mgr.Restart(context.Background(), "nginx.service"); err != nil {
		t.Fatalf("Restart: unexpected error: %v", err)
	}
}

func TestReload_Success(t *testing.T) {
	t.Parallel()

	exec := newFake(map[string]fakeResponse{
		"reload nginx.service": {},
	})
	mgr := systemd.NewWithExecutor(exec)

	if err := mgr.Reload(context.Background(), "nginx.service"); err != nil {
		t.Fatalf("Reload: unexpected error: %v", err)
	}
}

func TestStop_ExecError(t *testing.T) {
	t.Parallel()

	exec := newFake(map[string]fakeResponse{
		"stop nginx.service": {
			err: errors.New("access denied"),
		},
	})
	mgr := systemd.NewWithExecutor(exec)

	if err := mgr.Stop(context.Background(), "nginx.service"); err == nil {
		t.Fatal("Stop: expected error, got nil")
	}
}

// ---- ListServices ----

func TestListServices_ParsesOutput(t *testing.T) {
	t.Parallel()

	// systemctl list-units --type=service --no-pager --no-legend --all output
	output := `  nginx.service    loaded active   running A high performance web server
  sshd.service     loaded active   running OpenSSH server daemon
  redis.service    loaded inactive dead    Redis in-memory data store
`
	exec := newFake(map[string]fakeResponse{
		"list-units --type=service --no-pager --no-legend --all": {stdout: output},
	})
	mgr := systemd.NewWithExecutor(exec)

	services, err := mgr.ListServices(context.Background())
	if err != nil {
		t.Fatalf("ListServices: %v", err)
	}
	if len(services) != 3 {
		t.Fatalf("ListServices: got %d services, want 3", len(services))
	}

	// nginx should be running
	var nginx *platform.ServiceInfo
	for i := range services {
		if services[i].Name == "nginx.service" {
			nginx = &services[i]
			break
		}
	}
	if nginx == nil {
		t.Fatal("nginx.service not found in list")
	}
	if nginx.State != platform.StateRunning {
		t.Errorf("nginx state: got %q, want %q", nginx.State, platform.StateRunning)
	}
}

func TestListServices_EmptyOutput(t *testing.T) {
	t.Parallel()

	exec := newFake(map[string]fakeResponse{
		"list-units --type=service --no-pager --no-legend --all": {stdout: ""},
	})
	mgr := systemd.NewWithExecutor(exec)

	services, err := mgr.ListServices(context.Background())
	if err != nil {
		t.Fatalf("ListServices: %v", err)
	}
	if len(services) != 0 {
		t.Errorf("expected 0 services, got %d", len(services))
	}
}

// ---- Describe ----

func TestDescribe_ParsesOutput(t *testing.T) {
	t.Parallel()

	exec := newFake(map[string]fakeResponse{
		"show --no-pager --property=ActiveState,SubState,LoadState,Description,MainPID,ActiveEnterTimestamp nginx.service": {
			stdout: "ActiveState=active\nSubState=running\nLoadState=loaded\nDescription=nginx web server\nMainPID=9999\nActiveEnterTimestamp=\n",
		},
		"show --no-pager --property=FragmentPath,ExecStart nginx.service": {
			stdout: "FragmentPath=/usr/lib/systemd/system/nginx.service\nExecStart={ path=/usr/sbin/nginx ; argv[]=/usr/sbin/nginx -g daemon off; ; ...}\n",
		},
	})
	mgr := systemd.NewWithExecutor(exec)

	detail, err := mgr.Describe(context.Background(), "nginx.service")
	if err != nil {
		t.Fatalf("Describe: %v", err)
	}
	if detail.Name != "nginx.service" {
		t.Errorf("Name: got %q, want %q", detail.Name, "nginx.service")
	}
	if detail.UnitFilePath == "" {
		t.Error("UnitFilePath must not be empty")
	}
}

// ---- GetLogs ----

func TestGetLogs_ReturnsLines(t *testing.T) {
	t.Parallel()

	logOutput := "Apr 01 10:00:00 nginx[1]: started\nApr 01 10:00:01 nginx[1]: listening\n"
	exec := newFake(map[string]fakeResponse{
		"--no-pager -u nginx.service -n 50": {stdout: logOutput},
	})
	mgr := systemd.NewWithExecutor(exec)

	lines, err := mgr.GetLogs(context.Background(), "nginx.service", 50)
	if err != nil {
		t.Fatalf("GetLogs: %v", err)
	}
	if len(lines) != 2 {
		t.Errorf("GetLogs: got %d lines, want 2", len(lines))
	}
}

// ---- Name ----

func TestName(t *testing.T) {
	t.Parallel()

	mgr := systemd.NewWithExecutor(newFake(nil))
	if mgr.Name() != "systemd" {
		t.Errorf("Name: got %q, want %q", mgr.Name(), "systemd")
	}
}

// ---- State mapping ----

func TestGetStatus_StateMapping(t *testing.T) {
	t.Parallel()

	propKey := "show --no-pager --property=ActiveState,SubState,LoadState,Description,MainPID,ActiveEnterTimestamp nginx.service"
	cases := []struct {
		active string
		sub    string
		want   platform.ServiceState
	}{
		{"active", "running", platform.StateRunning},
		{"active", "exited", platform.StateStarting},
		{"inactive", "dead", platform.StateStopped},
		{"deactivating", "stop-sigterm", platform.StateStopped},
		{"failed", "failed", platform.StateFailed},
		{"activating", "start", platform.StateStarting},
		{"maintenance", "", platform.StateUnknown},
	}
	for _, tc := range cases {
		exec := newFake(map[string]fakeResponse{
			propKey: {
				stdout: "ActiveState=" + tc.active + "\nSubState=" + tc.sub + "\nLoadState=loaded\nDescription=test\nMainPID=0\nActiveEnterTimestamp=\n",
			},
		})
		mgr := systemd.NewWithExecutor(exec)
		status, err := mgr.GetStatus(context.Background(), "nginx.service")
		if err != nil {
			t.Fatalf("GetStatus [%s/%s]: unexpected error: %v", tc.active, tc.sub, err)
		}
		if status.State != tc.want {
			t.Errorf("GetStatus [%s/%s]: got state %q, want %q", tc.active, tc.sub, status.State, tc.want)
		}
	}
}

// ---- Error paths ----

func TestListServices_ExecError(t *testing.T) {
	t.Parallel()

	exec := newFake(map[string]fakeResponse{
		"list-units --type=service --no-pager --no-legend --all": {err: errors.New("dbus error")},
	})
	mgr := systemd.NewWithExecutor(exec)

	_, err := mgr.ListServices(context.Background())
	if err == nil {
		t.Fatal("ListServices: expected error, got nil")
	}
}

func TestGetLogs_ExecError(t *testing.T) {
	t.Parallel()

	exec := newFake(map[string]fakeResponse{
		"--no-pager -u nginx.service -n 10": {err: errors.New("journald error")},
	})
	mgr := systemd.NewWithExecutor(exec)

	_, err := mgr.GetLogs(context.Background(), "nginx.service", 10)
	if err == nil {
		t.Fatal("GetLogs: expected error, got nil")
	}
}

func TestDescribe_GetStatusError(t *testing.T) {
	t.Parallel()

	exec := newFake(map[string]fakeResponse{
		"show --no-pager --property=ActiveState,SubState,LoadState,Description,MainPID,ActiveEnterTimestamp nginx.service": {
			err: errors.New("unit not found"),
		},
	})
	mgr := systemd.NewWithExecutor(exec)

	_, err := mgr.Describe(context.Background(), "nginx.service")
	if err == nil {
		t.Fatal("Describe: expected error when GetStatus fails, got nil")
	}
}

func TestDescribe_SecondCallError(t *testing.T) {
	t.Parallel()

	exec := newFake(map[string]fakeResponse{
		"show --no-pager --property=ActiveState,SubState,LoadState,Description,MainPID,ActiveEnterTimestamp nginx.service": {
			stdout: "ActiveState=active\nSubState=running\nLoadState=loaded\nDescription=nginx\nMainPID=0\nActiveEnterTimestamp=\n",
		},
		"show --no-pager --property=FragmentPath,ExecStart nginx.service": {
			err: errors.New("describe failed"),
		},
	})
	mgr := systemd.NewWithExecutor(exec)

	_, err := mgr.Describe(context.Background(), "nginx.service")
	if err == nil {
		t.Fatal("Describe: expected error from second show call, got nil")
	}
}

// ---- parseListUnits edge cases ----

func TestListServices_ShortLines(t *testing.T) {
	t.Parallel()

	// Lines with fewer than 4 fields should be silently skipped.
	output := "  only-three fields here\n  nginx.service loaded active running nginx\n"
	exec := newFake(map[string]fakeResponse{
		"list-units --type=service --no-pager --no-legend --all": {stdout: output},
	})
	mgr := systemd.NewWithExecutor(exec)

	services, err := mgr.ListServices(context.Background())
	if err != nil {
		t.Fatalf("ListServices: %v", err)
	}
	if len(services) != 1 {
		t.Errorf("expected 1 service (short line skipped), got %d", len(services))
	}
}

// ---- init / platform registration ----

func TestPlatformRegistration_Linux(t *testing.T) {
	t.Parallel()

	mgr, err := platform.DetectFor("linux")
	if err != nil {
		t.Fatalf("DetectFor(linux): unexpected error: %v", err)
	}
	if mgr.Name() != "systemd" {
		t.Errorf("Name: got %q, want %q", mgr.Name(), "systemd")
	}
}
