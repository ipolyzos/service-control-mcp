// Package runtime wires the plugin registry to the MCP stdio transport.
package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"

	"github.com/ipolyzos/service-control-mcp/internal/plugin"
)

const serverName = "service-control-mcp"

// Server wraps the MCP server and the plugin registry.
type Server struct {
	mcp      *mcpserver.MCPServer
	registry *plugin.Registry
}

// New creates an MCP server and registers all tools from the given registry.
// version is injected at build time via ldflags (-X main.version=…).
func New(reg *plugin.Registry, version string) *Server {
	s := &Server{
		mcp:      mcpserver.NewMCPServer(serverName, version),
		registry: reg,
	}
	for _, t := range reg.List() {
		s.register(t)
	}
	return s
}

// Serve starts the stdio transport and blocks until ctx is cancelled or the
// client disconnects.
func (s *Server) Serve(ctx context.Context) error {
	stdio := mcpserver.NewStdioServer(s.mcp)
	return stdio.Listen(ctx, os.Stdin, os.Stdout)
}

// ExecuteTool invokes a registered tool by name with the given arguments.
// Tool-level errors are returned as an isError result, not a Go error.
// A Go error is returned only when the tool name is unknown or result
// marshalling fails.
func (s *Server) ExecuteTool(ctx context.Context, name string, args map[string]any) (*mcp.CallToolResult, error) {
	t, err := s.registry.Lookup(name)
	if err != nil {
		return nil, fmt.Errorf("runtime: %w", err)
	}
	result, err := t.Execute(ctx, args)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(err.Error())},
			IsError: true,
		}, nil
	}
	return encodeResult(result)
}

func (s *Server) register(t plugin.Tool) {
	tool := mcp.NewToolWithRawSchema(t.Name(), t.Description(), t.InputSchema())
	s.mcp.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		if args == nil {
			args = map[string]any{}
		}
		return s.ExecuteTool(ctx, t.Name(), args)
	})
}

func encodeResult(v any) (*mcp.CallToolResult, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("runtime: marshal result: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(string(b))},
	}, nil
}
