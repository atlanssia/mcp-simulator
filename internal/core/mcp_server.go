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
) func(context.Context, *mcp.CallToolRequest, map[string]any) (*mcp.CallToolResult, map[string]any, error) {
	// Create a handler function that matches the signature expected by mcp.AddTool
	// We use map[string]any for arguments to support dynamic inputs
	// We use map[string]any for output to satisfy the SDK's requirement that output schema must be an object
	handler := func(ctx context.Context, request *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, map[string]any, error) {
		// Convert args to map[string]interface{} for the generator
		inputParams := make(map[string]interface{})
		for k, v := range args {
			inputParams[k] = v
		}

		// Generate mock response
		// Use default generation params if not provided in context (which they aren't here)
		genParams := DefaultGenerationParams()

		// TODO: In the future, we might want to pass generation params via context or other means

		resultData, err := w.generator.GenerateMockResponseWithParams(
			ctx,
			tool.Name,
			tool.Description,
			tool.InputSchema,
			inputParams,
			genParams,
		)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error generating response: %v", err)},
				},
			}, nil, nil
		}

		// Convert result to JSON string for text content
		resultJSON, err := json.MarshalIndent(resultData, "", "  ")
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error marshaling result: %v", err)},
				},
			}, nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(resultJSON)},
			},
		}, nil, nil
	}

	return handler
}

// GetHTTPHandler returns the official MCP SDK HTTP handler with SSE support
func (w *MCPServerWrapper) GetHTTPHandler() http.Handler {
	getServer := func(r *http.Request) *mcp.Server {
		return w.server
	}
	// Use NewSSEHandler for standard SSE support (GET-first) which Dify expects
	return mcp.NewSSEHandler(getServer, nil)
}

// UpdateTools refreshes tools from registry
func (w *MCPServerWrapper) UpdateTools() error {
	// Re-register all tools
	// Official SDK doesn't have dynamic tool removal, so we re-register
	return w.registerTools()
}
