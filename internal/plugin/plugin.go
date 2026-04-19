// Package plugin defines the Tool interface and the Registry that maps
// tool names to their implementations.
package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

// ErrToolNotFound is returned by Registry.Lookup when no tool is registered
// under the requested name.
var ErrToolNotFound = errors.New("tool not found")

// Tool is the interface that every MCP tool plugin must implement.
type Tool interface {
	// Name returns the unique tool identifier used in MCP requests.
	Name() string
	// Description is the human-readable summary exposed in the MCP tool list.
	Description() string
	// InputSchema returns the JSON Schema that describes valid inputs.
	InputSchema() json.RawMessage
	// Execute runs the tool with the given arguments and returns a result or error.
	Execute(ctx context.Context, args map[string]any) (any, error)
}

// Registry maps tool names to Tool implementations.
// It is safe for concurrent reads after all tools have been registered.
type Registry struct {
	tools map[string]Tool
	mu    sync.RWMutex
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]Tool)}
}

// Register adds p to the registry. It panics if p is nil and returns an error
// if a tool with the same name has already been registered.
func (r *Registry) Register(p Tool) error {
	if p == nil {
		panic("plugin.Register: nil Tool")
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	name := p.Name()
	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("plugin %q already registered", name)
	}
	r.tools[name] = p
	return nil
}

// Lookup returns the Tool registered under name. It returns ErrToolNotFound
// when no such tool exists.
func (r *Registry) Lookup(name string) (Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.tools[name]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrToolNotFound, name)
	}
	return p, nil
}

// List returns all registered tools in an unspecified order.
func (r *Registry) List() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Tool, 0, len(r.tools))
	for _, p := range r.tools {
		out = append(out, p)
	}
	return out
}
