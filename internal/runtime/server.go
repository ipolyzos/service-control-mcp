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

const (
	serverName    = "service-control-mcp"
	serverVersion = "0.1.0"
)

// Server wraps the MCP server and the plugin registry.
type Server struct {
	mcp *mcpserver.MCPServer
}

// New creates an MCP server and registers all tools from the given registry.
func New(reg *plugin.Registry) *Server {
	s := &Server{
		mcp: mcpserver.NewMCPServer(serverName, serverVersion),
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

func (s *Server) register(t plugin.Tool) {
	tool := mcp.NewToolWithRawSchema(t.Name(), t.Description(), t.InputSchema())
	s.mcp.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		if args == nil {
			args = map[string]any{}
		}
		result, err := t.Execute(ctx, args)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.NewTextContent(err.Error())},
				IsError: true,
			}, nil
		}
		return encodeResult(result)
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
