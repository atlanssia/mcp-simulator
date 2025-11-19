package core

import "context"

// VirtualServer represents a simulated MCP server instance.
type VirtualServer interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	ID() string
	Config() ServerConfig
}

// ServerConfig holds configuration for a VirtualServer.
type ServerConfig struct {
	ID   string
	Port int
	Name string
}

// MockStrategy defines how to generate a response for a given request.
type MockStrategy interface {
	Generate(ctx context.Context, params map[string]interface{}) (interface{}, error)
}

// Registry manages the dynamic registration of Tools, Resources, and Prompts.
type Registry interface {
	RegisterTool(tool Tool) error
	RegisterResource(resource Resource) error
	RegisterPrompt(prompt Prompt) error
	GetTool(name string) (Tool, bool)
	// Add other lookup methods as needed
}

// Tool represents an MCP tool definition.
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// Resource represents an MCP resource definition.
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType"`
}

// Prompt represents an MCP prompt definition.
type Prompt struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CallToolResult represents the result of a tool execution.
type CallToolResult struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

// Content represents a content item in a tool result.
type Content struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}
