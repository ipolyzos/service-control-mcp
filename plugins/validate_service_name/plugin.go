// Package validate_service_name implements the validate_service_name MCP tool plugin.
package validate_service_name

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ipolyzos/service-control-mcp/internal/security"
)

const (
	toolName    = "validate_service_name"
	description = "Validate and normalize service identifiers before write operations"
)

var inputSchema = json.RawMessage(`{
  "type": "object",
  "required": ["service_name"],
  "properties": {
    "service_name": {
      "type": "string",
      "description": "Service name to validate"
    }
  },
  "additionalProperties": false
}`)

// Plugin implements plugin.Tool for validate_service_name.
type Plugin struct{}

// New returns a Plugin.
func New() *Plugin { return &Plugin{} }

// Name returns the MCP tool name.
func (p *Plugin) Name() string { return toolName }

// Description returns the human-readable tool summary.
func (p *Plugin) Description() string { return description }

// InputSchema returns the JSON Schema for this tool's inputs.
func (p *Plugin) InputSchema() json.RawMessage { return inputSchema }

// Execute validates the service name and reports whether it is acceptable.
// Unlike other tools, this tool does not return an error on invalid names;
// instead it returns a structured result with valid=false and the reason.
func (p *Plugin) Execute(_ context.Context, args map[string]any) (any, error) {
	v, ok := args["service_name"]
	if !ok {
		return nil, fmt.Errorf("validate_service_name: missing required argument %q", "service_name")
	}
	name, ok := v.(string)
	if !ok {
		return nil, fmt.Errorf("validate_service_name: argument %q must be a string, got %T", "service_name", v)
	}

	if err := security.ValidateServiceName(name); err != nil {
		return map[string]any{
			"service_name": name,
			"valid":        false,
			"error":        err.Error(),
		}, nil
	}
	return map[string]any{
		"service_name": name,
		"valid":        true,
		"error":        "",
	}, nil
}
