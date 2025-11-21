package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPServerWrapper wraps official MCP SDK server with our tool registry
type MCPServerWrapper struct {
	server    *mcp.Server
	registry  Registry
	generator MockGenerator
}

// NewMCPServerWrapper creates a new MCP server with tools from registry
func NewMCPServerWrapper(
	name string,
	registry Registry,
	gen MockGenerator,
) (*MCPServerWrapper, error) {
	// Create official MCP SDK server
	impl := &mcp.Implementation{
		Name:    name,
		Version: "1.0.0",
	}

	server := mcp.NewServer(impl, nil)

	wrapper := &MCPServerWrapper{
		server:    server,
		registry:  registry,
		generator: gen,
	}

	// Register all tools from registry
	if err := wrapper.registerTools(); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}

	return wrapper, nil
}

// registerTools converts our tools to MCP SDK tools and registers them
func (w *MCPServerWrapper) registerTools() error {
	tools := w.registry.ListTools()

	for _, tool := range tools {
		// Convert our tool to mcp.Tool
		mcpTool := &mcp.Tool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: tool.InputSchema,
		}

		// Create handler that uses our generator
		handler := w.createToolHandler(tool)

		// Register with official MCP SDK server
		// Using map[string]any for dynamic arguments
		mcp.AddTool(w.server, mcpTool, handler)
	}

	return nil
}

// createToolHandler creates a handler function for the tool
// Official SDK expects: func(ctx, req, input) (*CallToolResult, output, error)
func (w *MCPServerWrapper) createToolHandler(
	tool Tool,
) func(context.Context, *mcp.CallToolRequest, map[string]any) (*mcp.CallToolResult, string, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, string, error) {
		// Generate mock data using our generator
		mockData, err := w.generator.GenerateMockResponseWithParams(
			ctx,
			tool.Name,
			tool.Description,
			tool.InputSchema,
			args,
			DefaultGenerationParams(),
		)
		if err != nil {
			// Return error via CallToolResult
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
				},
			}, "", nil
		}

		// Convert mockData to JSON string if it's not already
		var resultText string
		switch v := mockData.(type) {
		case string:
			resultText = v
		case []byte:
			resultText = string(v)
		default:
			// Marshal to JSON
			jsonData, err := json.Marshal(mockData)
			if err != nil {
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("failed to marshal response: %v", err)},
					},
				}, "", nil
			}
			resultText = string(jsonData)
		}

		// Return result - the second return value is the "output" that will be marshaled
		// We return nil CallToolResult to let SDK create it automatically from the string output
		return nil, resultText, nil
	}
}

// GetHTTPHandler returns the official MCP SDK HTTP handler with SSE support
func (w *MCPServerWrapper) GetHTTPHandler() http.Handler {
	// Use official SDK's NewStreamableHTTPHandler
	// The first argument is a function that returns the server for a given request
	// Since we have a single server instance, we return it for all requests
	getServer := func(r *http.Request) *mcp.Server {
		return w.server
	}

	// This automatically handles:
	// - GET / : discovery
	// - GET / with Accept:text/event-stream : SSE connection + endpoint event
	// - POST / : JSON-RPC requests (with SSE response if session exists)
	return mcp.NewStreamableHTTPHandler(getServer, nil)
}

// UpdateTools refreshes tools from registry
func (w *MCPServerWrapper) UpdateTools() error {
	// Re-register all tools
	// Official SDK doesn't have dynamic tool removal, so we re-register
	return w.registerTools()
}
