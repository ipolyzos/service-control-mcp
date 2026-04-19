// Package describe_service implements the describe_service MCP tool plugin.
package describe_service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ipolyzos/service-control-mcp/internal/platform"
	"github.com/ipolyzos/service-control-mcp/internal/security"
)

const (
	toolName    = "describe_service"
	description = "Return normalized metadata and manager-specific details for a service"
)

var inputSchema = json.RawMessage(`{
  "type": "object",
  "required": ["service_name"],
  "properties": {
    "service_name": {
      "type": "string",
      "description": "Name of the service to describe"
    }
  },
  "additionalProperties": false
}`)

// Manager is the subset of platform.Manager required by this plugin.
type Manager interface {
	Name() string
	Describe(ctx context.Context, name string) (platform.ServiceDetail, error)
}

// Plugin implements plugin.Tool for describe_service.
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

// Execute returns the full description of the requested service.
func (p *Plugin) Execute(ctx context.Context, args map[string]any) (any, error) {
	name, err := extractString(args, "service_name")
	if err != nil {
		return nil, fmt.Errorf("describe_service: %w", err)
	}
	err = security.ValidateServiceName(name)
	if err != nil {
		return nil, fmt.Errorf("describe_service: %w", err)
	}
	detail, err := p.mgr.Describe(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("describe_service: %w", err)
	}
	return detail, nil
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
