// Package get_service_status implements the get_service_status MCP tool plugin.
package get_service_status

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ipolyzos/service-control-mcp/internal/platform"
	"github.com/ipolyzos/service-control-mcp/internal/security"
)

const (
	toolName    = "get_service_status"
	description = "Return normalized status for a single service"
)

var inputSchema = json.RawMessage(`{
  "type": "object",
  "required": ["service_name"],
  "properties": {
    "service_name": {
      "type": "string",
      "description": "Name of the service to query"
    }
  },
  "additionalProperties": false
}`)

// Manager is the subset of platform.Manager required by this plugin.
type Manager interface {
	Name() string
	GetStatus(ctx context.Context, name string) (platform.ServiceStatus, error)
}

// Plugin implements plugin.Tool for get_service_status.
type Plugin struct {
	mgr Manager
}

// New returns a Plugin backed by the given Manager.
func New(mgr Manager) *Plugin {
	return &Plugin{mgr: mgr}
}

// Name returns the MCP tool name.
func (p *Plugin) Name() string { return toolName }

// Description returns the human-readable tool summary.
func (p *Plugin) Description() string { return description }

// InputSchema returns the JSON Schema for this tool's inputs.
func (p *Plugin) InputSchema() json.RawMessage { return inputSchema }

// Execute returns the normalized status of the requested service.
func (p *Plugin) Execute(ctx context.Context, args map[string]any) (any, error) {
	name, err := extractString(args, "service_name")
	if err != nil {
		return nil, fmt.Errorf("get_service_status: %w", err)
	}
	err = security.ValidateServiceName(name)
	if err != nil {
		return nil, fmt.Errorf("get_service_status: %w", err)
	}
	status, err := p.mgr.GetStatus(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("get_service_status: %w", err)
	}
	return status, nil
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
