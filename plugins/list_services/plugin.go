// Package list_services implements the list_services MCP tool plugin.
package list_services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ipolyzos/service-control-mcp/internal/platform"
)

const (
	toolName    = "list_services"
	description = "Enumerate manageable services visible to the current user and platform adapter"
)

var inputSchema = json.RawMessage(`{
  "type": "object",
  "properties": {},
  "additionalProperties": false
}`)

// Manager is the subset of platform.Manager required by this plugin.
type Manager interface {
	Name() string
	ListServices(ctx context.Context) ([]platform.ServiceInfo, error)
}

// Plugin implements plugin.Tool for list_services.
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

// Execute lists all services visible to the platform adapter.
func (p *Plugin) Execute(ctx context.Context, _ map[string]any) (any, error) {
	services, err := p.mgr.ListServices(ctx)
	if err != nil {
		return nil, fmt.Errorf("list_services: %w", err)
	}
	if services == nil {
		services = []platform.ServiceInfo{}
	}
	return map[string]any{
		"services": services,
		"manager":  p.mgr.Name(),
		"count":    len(services),
	}, nil
}
