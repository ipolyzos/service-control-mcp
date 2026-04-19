// Package systemd implements the platform.Manager interface for Linux systems
// managed by systemd. It wraps systemctl with an allowlisted subcommand set,
// a 30-second execution timeout, and no privilege escalation.
package systemd

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/ipolyzos/service-control-mcp/internal/platform"
)

const (
	execTimeout = 30 * time.Second
	managerName = "systemd"
)

// Executor abstracts command execution so the adapter can be tested without
// running real systemctl.
type Executor interface {
	Run(ctx context.Context, args []string) (stdout, stderr string, err error)
}

// systemctlExec is the production Executor backed by systemctl.
type systemctlExec struct{}

func (s *systemctlExec) Run(ctx context.Context, args []string) (string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, execTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "systemctl", args...) //nolint:gosec // args are allowlisted by caller
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// journalctlExec is the production Executor backed by journalctl.
type journalctlExec struct{}

func (j *journalctlExec) Run(ctx context.Context, args []string) (string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, execTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "journalctl", args...) //nolint:gosec // args are allowlisted by caller
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// Manager is the systemd platform adapter.
type Manager struct {
	exec    Executor
	logExec Executor
}

// New returns a Manager backed by the real systemctl and journalctl binaries.
// It registers itself with the platform package for linux.
func New() *Manager {
	return &Manager{
		exec:    &systemctlExec{},
		logExec: &journalctlExec{},
	}
}

// NewWithExecutor returns a Manager using the supplied executor, used in tests.
func NewWithExecutor(e Executor) *Manager {
	return &Manager{exec: e, logExec: e}
}

// Name returns the identifier for this manager.
func (m *Manager) Name() string { return managerName }

// ListServices enumerates all service units visible to the current user.
func (m *Manager) ListServices(ctx context.Context) ([]platform.ServiceInfo, error) {
	args := []string{"list-units", "--type=service", "--no-pager", "--no-legend", "--all"}
	stdout, _, err := m.exec.Run(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("systemd: list-units: %w", err)
	}
	return parseListUnits(stdout), nil
}

// GetStatus returns the normalized status for a single service unit.
func (m *Manager) GetStatus(ctx context.Context, name string) (platform.ServiceStatus, error) {
	props := "ActiveState,SubState,LoadState,Description,MainPID,ActiveEnterTimestamp"
	args := []string{"show", "--no-pager", "--property=" + props, name}
	stdout, _, err := m.exec.Run(ctx, args)
	if err != nil {
		return platform.ServiceStatus{}, fmt.Errorf("systemd: show %s: %w", name, err)
	}
	return parseServiceStatus(name, stdout), nil
}

// Describe returns rich metadata for a single service unit.
func (m *Manager) Describe(ctx context.Context, name string) (platform.ServiceDetail, error) {
	status, err := m.GetStatus(ctx, name)
	if err != nil {
		return platform.ServiceDetail{}, err
	}

	args := []string{"show", "--no-pager", "--property=FragmentPath,ExecStart", name}
	stdout, _, err := m.exec.Run(ctx, args)
	if err != nil {
		return platform.ServiceDetail{}, fmt.Errorf("systemd: describe %s: %w", name, err)
	}
	kvs := parseProperties(stdout)

	return platform.ServiceDetail{
		Name:         name,
		Manager:      managerName,
		State:        status.State,
		Description:  status.Description,
		UnitFilePath: kvs["FragmentPath"],
		ExecStart:    kvs["ExecStart"],
	}, nil
}

// GetLogs returns up to lines recent journal entries for a service unit.
func (m *Manager) GetLogs(ctx context.Context, name string, lines int) ([]string, error) {
	args := []string{
		"--no-pager",
		"-u", name,
		"-n", strconv.Itoa(lines),
	}
	stdout, _, err := m.logExec.Run(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("systemd: journalctl %s: %w", name, err)
	}
	return splitLines(stdout), nil
}

// Start starts a service unit.
func (m *Manager) Start(ctx context.Context, name string) error {
	return m.runUnit(ctx, "start", name)
}

// Stop stops a service unit.
func (m *Manager) Stop(ctx context.Context, name string) error {
	return m.runUnit(ctx, "stop", name)
}

// Restart restarts a service unit.
func (m *Manager) Restart(ctx context.Context, name string) error {
	return m.runUnit(ctx, "restart", name)
}

// Reload sends a SIGHUP-equivalent reload to a service unit.
func (m *Manager) Reload(ctx context.Context, name string) error {
	return m.runUnit(ctx, "reload", name)
}

func (m *Manager) runUnit(ctx context.Context, action, name string) error {
	_, _, err := m.exec.Run(ctx, []string{action, name})
	if err != nil {
		return fmt.Errorf("systemd: %s %s: %w", action, name, err)
	}
	return nil
}

// parseListUnits parses `systemctl list-units --no-legend` output.
// Each line has the format:
//
//	<unit> <load> <active> <sub> <description...>
func parseListUnits(output string) []platform.ServiceInfo {
	var services []platform.ServiceInfo
	for _, line := range splitLines(output) {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		name := fields[0]
		active := fields[2]
		sub := fields[3]

		services = append(services, platform.ServiceInfo{
			Name:    name,
			Manager: managerName,
			State:   activeSubToState(active, sub),
		})
	}
	return services
}

// parseServiceStatus parses `systemctl show --property=...` output.
func parseServiceStatus(name, output string) platform.ServiceStatus {
	kvs := parseProperties(output)

	pid, _ := strconv.Atoi(kvs["MainPID"])
	return platform.ServiceStatus{
		Name:        name,
		Manager:     managerName,
		State:       activeSubToState(kvs["ActiveState"], kvs["SubState"]),
		SubState:    kvs["SubState"],
		LoadState:   kvs["LoadState"],
		Description: kvs["Description"],
		ActiveSince: kvs["ActiveEnterTimestamp"],
		PID:         pid,
	}
}

// parseProperties parses key=value lines from systemctl show output.
func parseProperties(output string) map[string]string {
	kvs := make(map[string]string)
	for _, line := range splitLines(output) {
		if idx := strings.IndexByte(line, '='); idx > 0 {
			kvs[line[:idx]] = line[idx+1:]
		}
	}
	return kvs
}

// activeSubToState maps systemd ActiveState/SubState pairs to ServiceState.
func activeSubToState(active, sub string) platform.ServiceState {
	switch active {
	case "active":
		if sub == "running" {
			return platform.StateRunning
		}
		return platform.StateStarting
	case "inactive", "deactivating":
		return platform.StateStopped
	case "failed":
		return platform.StateFailed
	case "activating":
		return platform.StateStarting
	default:
		return platform.StateUnknown
	}
}

// splitLines splits s on newlines, omitting trailing empty entries.
func splitLines(s string) []string {
	lines := strings.Split(s, "\n")
	// trim trailing empty line from a newline-terminated string
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func init() {
	platform.Register("linux", func() (platform.Manager, error) {
		return New(), nil
	})
}
