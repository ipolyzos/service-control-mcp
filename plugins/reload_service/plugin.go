// Package reload_service implements the reload_service MCP tool plugin.
package reload_service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ipolyzos/service-control-mcp/internal/audit"
	"github.com/ipolyzos/service-control-mcp/internal/security"
)

const (
	toolName    = "reload_service"
	description = "Send a configuration reload signal to a running service"
)

var inputSchema = json.RawMessage(`{
  "type": "object",
  "required": ["service_name"],
  "properties": {
    "service_name": {
      "type": "string",
      "description": "Name of the service to reload"
    }
  },
  "additionalProperties": false
}`)

// Manager is the subset of platform.Manager required by this plugin.
type Manager interface {
	Name() string
	Reload(ctx context.Context, name string) error
}

// Plugin implements plugin.Tool for reload_service.
type Plugin struct {
	mgr Manager
	log *audit.Logger
}

// New returns a Plugin backed by the given Manager and audit Logger.
func New(mgr Manager, log *audit.Logger) *Plugin {
	return &Plugin{mgr: mgr, log: log}
}

// Name returns the MCP tool name.
func (p *Plugin) Name() string { return toolName }

// Description returns the human-readable tool summary.
func (p *Plugin) Description() string { return description }

// InputSchema returns the JSON Schema for this tool's inputs.
func (p *Plugin) InputSchema() json.RawMessage { return inputSchema }

// Execute sends a reload signal to the requested service.
func (p *Plugin) Execute(ctx context.Context, args map[string]any) (any, error) {
	name, err := extractString(args, "service_name")
	if err != nil {
		return nil, fmt.Errorf("reload_service: %w", err)
	}
	err = security.ValidateServiceName(name)
	if err != nil {
		return nil, fmt.Errorf("reload_service: %w", err)
	}

	err = p.mgr.Reload(ctx, name)
	outcome := audit.OutcomeSuccess
	errMsg := ""
	if err != nil {
		outcome = audit.OutcomeFailure
		errMsg = err.Error()
	}
	_ = p.log.Write(audit.Entry{
		AuditID:     newAuditID(),
		Timestamp:   time.Now().UTC(),
		Action:      toolName,
		ServiceName: name,
		Manager:     p.mgr.Name(),
		Outcome:     outcome,
		Error:       errMsg,
	})
	if err != nil {
		return nil, fmt.Errorf("reload_service: %w", err)
	}
	return map[string]any{
		"service_name": name,
		"manager":      p.mgr.Name(),
		"outcome":      string(audit.OutcomeSuccess),
	}, nil
}

func newAuditID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func extractString(args map[string]any, key string) (string, error) {
	v, ok := args[key]
	if !ok {
		return "", fmt.Errorf("missing required argument %q", key)
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("argument %q must be a string, got %T", key, v)
	}
	return s, nil
}
