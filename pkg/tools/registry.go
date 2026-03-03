package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

// ToolFunc is a function that executes a tool and returns a JSON result string.
type ToolFunc func(ctx context.Context, input json.RawMessage) (string, error)

// Registry maps tool names to their handler functions.
type Registry struct {
	handlers map[string]ToolFunc
}

func NewRegistry() *Registry {
	return &Registry{
		handlers: make(map[string]ToolFunc),
	}
}

func (r *Registry) Register(name string, fn ToolFunc) {
	r.handlers[name] = fn
}

func (r *Registry) Execute(ctx context.Context, name string, input json.RawMessage) (string, error) {
	fn, ok := r.handlers[name]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", name)
	}
	return fn(ctx, input)
}

func (r *Registry) Has(name string) bool {
	_, ok := r.handlers[name]
	return ok
}
