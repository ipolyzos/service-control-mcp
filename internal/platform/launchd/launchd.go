// Package launchd implements the platform.Manager interface for macOS systems
// managed by launchd. It wraps launchctl with an allowlisted subcommand set
// and a 30-second execution timeout. No privilege escalation is performed;
// the adapter operates under the invoking user's bootstrap context.
package launchd

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/ipolyzos/service-control-mcp/internal/platform"
)

const (
	execTimeout   = 30 * time.Second
	managerName   = "launchd"
	systemDomain  = "system"
	plistBasePath = "/Library/LaunchDaemons/"
)

// Executor abstracts command execution for testability.
type Executor interface {
	Run(ctx context.Context, args []string) (stdout, stderr string, err error)
}

// launchctlExec is the production Executor backed by launchctl.
type launchctlExec struct{}

func (l *launchctlExec) Run(ctx context.Context, args []string) (string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, execTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "launchctl", args...) //nolint:gosec // args are allowlisted by caller
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// Manager is the launchd platform adapter.
type Manager struct {
	exec Executor
}

// New returns a Manager backed by the real launchctl binary.
func New() *Manager {
	return &Manager{exec: &launchctlExec{}}
}

// NewWithExecutor returns a Manager using the supplied executor, used in tests.
func NewWithExecutor(e Executor) *Manager {
	return &Manager{exec: e}
}

// Name returns the identifier for this manager.
func (m *Manager) Name() string { return managerName }

// ListServices enumerates services visible via `launchctl list`.
func (m *Manager) ListServices(ctx context.Context) ([]platform.ServiceInfo, error) {
	stdout, _, err := m.exec.Run(ctx, []string{"list"})
	if err != nil {
		return nil, fmt.Errorf("launchd: list: %w", err)
	}
	return parseList(stdout), nil
}

// GetStatus returns the normalized status for a named service.
func (m *Manager) GetStatus(ctx context.Context, name string) (platform.ServiceStatus, error) {
	target := systemDomain + "/" + name
	stdout, _, err := m.exec.Run(ctx, []string{"print", target})
	if err != nil {
		return platform.ServiceStatus{}, fmt.Errorf("launchd: print %s: %w", name, err)
	}
	return parsePrint(name, stdout), nil
}

// Describe returns normalized metadata for a named service.
func (m *Manager) Describe(ctx context.Context, name string) (platform.ServiceDetail, error) {
	status, err := m.GetStatus(ctx, name)
	if err != nil {
		return platform.ServiceDetail{}, err
	}
	return platform.ServiceDetail{
		Name:    name,
		Manager: managerName,
		State:   status.State,
	}, nil
}

// GetLogs is not supported on macOS via launchctl; callers should use
// the system log or Console.app instead.
func (m *Manager) GetLogs(_ context.Context, name string, _ int) ([]string, error) {
	return nil, fmt.Errorf("launchd: get_logs for %q: %w", name, errNotSupported)
}

// Start loads and starts a service by bootstrapping its plist.
func (m *Manager) Start(ctx context.Context, name string) error {
	plist := plistBasePath + name + ".plist"
	_, _, err := m.exec.Run(ctx, []string{"bootstrap", systemDomain, plist})
	if err != nil {
		return fmt.Errorf("launchd: bootstrap %s: %w", name, err)
	}
	return nil
}

// Stop unloads a service from the system bootstrap context.
func (m *Manager) Stop(ctx context.Context, name string) error {
	target := systemDomain + "/" + name
	_, _, err := m.exec.Run(ctx, []string{"bootout", target})
	if err != nil {
		return fmt.Errorf("launchd: bootout %s: %w", name, err)
	}
	return nil
}

// Restart forcefully restarts a service using kickstart -k.
func (m *Manager) Restart(ctx context.Context, name string) error {
	return m.kickstart(ctx, name)
}

// Reload reloads a service by restarting it with kickstart -k.
// launchd has no native reload signal; kickstart is the closest equivalent.
func (m *Manager) Reload(ctx context.Context, name string) error {
	return m.kickstart(ctx, name)
}

func (m *Manager) kickstart(ctx context.Context, name string) error {
	target := systemDomain + "/" + name
	_, _, err := m.exec.Run(ctx, []string{"kickstart", "-k", target})
	if err != nil {
		return fmt.Errorf("launchd: kickstart %s: %w", name, err)
	}
	return nil
}

// errNotSupported is a sentinel for operations that have no launchd equivalent.
var errNotSupported = errors.New("not supported by launchd adapter")

// parseList parses `launchctl list` output.
// Each line: PID  ExitCode  Label
// A PID of "-" means the service is not running.
func parseList(output string) []platform.ServiceInfo {
	var services []platform.ServiceInfo
	for _, line := range splitLines(output) {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		pid := fields[0]
		label := fields[2]

		state := platform.StateStopped
		if pid != "-" && pid != "0" {
			state = platform.StateRunning
		}

		services = append(services, platform.ServiceInfo{
			Name:    label,
			Manager: managerName,
			State:   state,
		})
	}
	return services
}

// parsePrint parses `launchctl print system/<name>` output for key status fields.
func parsePrint(name, output string) platform.ServiceStatus {
	state := platform.StateUnknown

	for _, line := range splitLines(output) {
		line = strings.TrimSpace(line)
		if strings.Contains(line, `"state"`) {
			if strings.Contains(line, `"running"`) {
				state = platform.StateRunning
			} else {
				state = platform.StateStopped
			}
		}
	}

	return platform.ServiceStatus{
		Name:    name,
		Manager: managerName,
		State:   state,
	}
}

func splitLines(s string) []string {
	lines := strings.Split(s, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func init() {
	platform.Register("darwin", func() (platform.Manager, error) {
		return New(), nil
	})
}
