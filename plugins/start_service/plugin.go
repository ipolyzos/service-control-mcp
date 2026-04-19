// Package start_service implements the start_service MCP tool plugin.
package start_service

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
	toolName    = "start_service"
	description = "Start a stopped service"
)

var inputSchema = json.RawMessage(`{
  "type": "object",
  "required": ["service_name"],
  "properties": {
    "service_name": {
      "type": "string",
      "description": "Name of the service to start"
    }
  },
  "additionalProperties": false
}`)

// Manager is the subset of platform.Manager required by this plugin.
type Manager interface {
	Name() string
	Start(ctx context.Context, name string) error
}

// Plugin implements plugin.Tool for start_service.
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

// Execute starts the requested service and emits a structured audit entry.
func (p *Plugin) Execute(ctx context.Context, args map[string]any) (any, error) {
	name, err := extractString(args, "service_name")
	if err != nil {
		return nil, fmt.Errorf("start_service: %w", err)
	}
	err = security.ValidateServiceName(name)
	if err != nil {
		return nil, fmt.Errorf("start_service: %w", err)
	}

	err = p.mgr.Start(ctx, name)
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
		return nil, fmt.Errorf("start_service: %w", err)
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
