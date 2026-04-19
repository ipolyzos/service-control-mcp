// Package platform defines the Manager interface, common data types, and
// the Detect function that returns the appropriate adapter for the current OS.
package platform

import (
	"context"
	"fmt"
	"runtime"
)

// ServiceState represents the normalized lifecycle state of a service.
type ServiceState string

// Service lifecycle states returned by all platform adapters.
const (
	StateRunning  ServiceState = "running"
	StateStopped  ServiceState = "stopped"
	StateFailed   ServiceState = "failed"
	StateUnknown  ServiceState = "unknown"
	StateStarting ServiceState = "starting"
	StateStopping ServiceState = "stopping"
)

// ServiceInfo holds the minimal normalized descriptor for a listed service.
type ServiceInfo struct {
	Name        string       `json:"name"`
	Manager     string       `json:"manager"`
	State       ServiceState `json:"state"`
	Description string       `json:"description,omitempty"`
}

// ServiceStatus holds the current status of a single service.
type ServiceStatus struct {
	Name        string       `json:"name"`
	Manager     string       `json:"manager"`
	State       ServiceState `json:"state"`
	SubState    string       `json:"sub_state,omitempty"`
	LoadState   string       `json:"load_state,omitempty"`
	Description string       `json:"description,omitempty"`
	ActiveSince string       `json:"active_since,omitempty"`
	PID         int          `json:"pid,omitempty"`
}

// ServiceDetail holds rich metadata about a single service.
type ServiceDetail struct {
	Extra        map[string]string `json:"extra,omitempty"`
	Name         string            `json:"name"`
	Manager      string            `json:"manager"`
	State        ServiceState      `json:"state"`
	Description  string            `json:"description,omitempty"`
	UnitFilePath string            `json:"unit_file_path,omitempty"`
	ExecStart    string            `json:"exec_start,omitempty"`
}

// Manager is the interface that each platform adapter must implement.
// All methods receive a context that carries the execution deadline.
type Manager interface {
	// Name returns the human-readable manager identifier (e.g. "systemd").
	Name() string

	// ListServices enumerates services visible to the current user.
	ListServices(ctx context.Context) ([]ServiceInfo, error)

	// GetStatus returns the normalized status for a named service.
	GetStatus(ctx context.Context, name string) (ServiceStatus, error)

	// Describe returns metadata and manager-specific detail for a service.
	Describe(ctx context.Context, name string) (ServiceDetail, error)

	// GetLogs returns up to lines recent log entries for a service.
	GetLogs(ctx context.Context, name string, lines int) ([]string, error)

	// Start starts a stopped service.
	Start(ctx context.Context, name string) error

	// Stop stops a running service.
	Stop(ctx context.Context, name string) error

	// Restart stops and restarts a service.
	Restart(ctx context.Context, name string) error

	// Reload sends a configuration reload signal to a running service.
	Reload(ctx context.Context, name string) error
}

// UnsupportedPlatformError is returned by Detect when no adapter is
// available for the current operating system.
type UnsupportedPlatformError struct {
	OS string
}

func (e *UnsupportedPlatformError) Error() string {
	return fmt.Sprintf("unsupported platform: %q", e.OS)
}

// ErrUnsupportedPlatform creates an UnsupportedPlatformError for the given OS.
func ErrUnsupportedPlatform(os string) error {
	return &UnsupportedPlatformError{OS: os}
}

// registry holds adapters registered by platform-specific init functions.
var registry = map[string]func() (Manager, error){}

// Register records a factory function for the given GOOS value.
// It is called from each platform adapter's init().
func Register(goos string, factory func() (Manager, error)) {
	registry[goos] = factory
}

// Detect returns a Manager for the current runtime platform, or an
// UnsupportedPlatformError if no adapter is registered for the OS.
func Detect() (Manager, error) {
	return DetectFor(runtime.GOOS)
}

// DetectFor returns a Manager for the given GOOS value. It is exported
// primarily to support testing with synthetic OS names.
func DetectFor(goos string) (Manager, error) {
	factory, ok := registry[goos]
	if !ok {
		return nil, ErrUnsupportedPlatform(goos)
	}
	return factory()
}
