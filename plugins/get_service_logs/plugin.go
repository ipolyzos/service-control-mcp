// Package get_service_logs implements the get_service_logs MCP tool plugin.
package get_service_logs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ipolyzos/service-control-mcp/internal/security"
)

const (
	toolName     = "get_service_logs"
	description  = "Fetch recent bounded logs for a service where supported"
	defaultLines = 100
)

var inputSchema = json.RawMessage(`{
  "type": "object",
  "required": ["service_name"],
  "properties": {
    "service_name": {
      "type": "string",
      "description": "Name of the service whose logs to retrieve"
    },
    "lines": {
      "type": "integer",
      "description": "Number of log lines to return (1–500, default 100)",
      "minimum": 1,
      "maximum": 500
    }
  },
  "additionalProperties": false
}`)

// Manager is the subset of platform.Manager required by this plugin.
type Manager interface {
	Name() string
	GetLogs(ctx context.Context, name string, lines int) ([]string, error)
}

// Plugin implements plugin.Tool for get_service_logs.
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

// Execute retrieves recent log lines for the requested service.
func (p *Plugin) Execute(ctx context.Context, args map[string]any) (any, error) {
	name, err := extractString(args, "service_name")
	if err != nil {
		return nil, fmt.Errorf("get_service_logs: %w", err)
	}
	err = security.ValidateServiceName(name)
	if err != nil {
		return nil, fmt.Errorf("get_service_logs: %w", err)
	}

	lines := defaultLines
	if v, ok := args["lines"]; ok {
		n, ok := v.(float64) // JSON numbers arrive as float64
		if !ok {
			return nil, fmt.Errorf("get_service_logs: argument %q must be a number", "lines")
		}
		lines = int(n)
	}

	err = security.ValidateLogLines(lines)
	if err != nil {
		return nil, fmt.Errorf("get_service_logs: %w", err)
	}

	logLines, err := p.mgr.GetLogs(ctx, name, lines)
	if err != nil {
		return nil, fmt.Errorf("get_service_logs: %w", err)
	}
	if logLines == nil {
		logLines = []string{}
	}
	return map[string]any{
		"service_name": name,
		"manager":      p.mgr.Name(),
		"lines":        logLines,
		"line_count":   len(logLines),
	}, nil
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
