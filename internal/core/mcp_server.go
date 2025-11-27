package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPServerWrapper wraps official MCP SDK server with our tool registry
type MCPServerWrapper struct {
	server       *mcp.Server
	registry     Registry
	generator    MockGenerator
	mockStrategy string // "llm" | "static" | "hybrid"
}

// NewMCPServerWrapper creates a new MCP server with tools from registry
func NewMCPServerWrapper(
	name string,
	registry Registry,
	gen MockGenerator,
	mockStrategy string,
) (*MCPServerWrapper, error) {
	// Create official MCP SDK server
	impl := &mcp.Implementation{
		Name:    name,
		Version: "1.0.0",
	}

	server := mcp.NewServer(impl, nil)

	wrapper := &MCPServerWrapper{
		server:       server,
		registry:     registry,
		generator:    gen,
		mockStrategy: mockStrategy,
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

		// Determine mock strategy (default to "hybrid" if empty)
		strategy := w.mockStrategy
		if strategy == "" {
			strategy = "hybrid"
		}

		var resultData interface{}
		var err error

		switch strategy {
		case "static":
			// Static-only mode: use tool-specific static data if available
			if tool.StaticMockData != nil {
				resultData = tool.StaticMockData
			} else {
				// Fallback to generic vitals for backward compatibility
				resultData = generateStaticVitals(inputParams)
			}

		case "llm":
			// LLM-only mode: only use LLM, fail if it fails
			genParams := DefaultGenerationParams()
			resultData, err = w.generator.GenerateMockResponseWithParams(
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
						&mcp.TextContent{Text: fmt.Sprintf("LLM generation failed: %v", err)},
					},
				}, nil, nil
			}

		case "hybrid":
			fallthrough
		default:
			// Hybrid mode (default): try LLM first, fallback to static
			genParams := DefaultGenerationParams()
			resultData, err = w.generator.GenerateMockResponseWithParams(
				ctx,
				tool.Name,
				tool.Description,
				tool.InputSchema,
				inputParams,
				genParams,
			)
			if err != nil {
				// Fallback to static vitals if LLM fails
				resultData = generateStaticVitals(inputParams)
			}
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

// generateStaticVitals returns realistic static vitals with properly structured data
// for multi-series charting. Compound vitals like BP are split into separate records.
func generateStaticVitals(params map[string]interface{}) interface{} {
	// Parse time range from params or use sensible defaults (last 4 hours)
	now := time.Now()

	var startTime, endTime time.Time

	// Try to parse start_time from params
	if st, ok := params["start_time"].(string); ok && st != "" {
		// Try parsing as time only (HH:MM)
		if parsed, err := time.Parse("15:04", st); err == nil {
			startTime = time.Date(now.Year(), now.Month(), now.Day(), parsed.Hour(), parsed.Minute(), 0, 0, now.Location())
		} else if parsed, err := time.Parse("2006-01-02 15:04", st); err == nil {
			startTime = parsed
		} else {
			// Fallback to 4 hours ago
			startTime = now.Add(-4 * time.Hour)
		}
	} else {
		startTime = now.Add(-4 * time.Hour)
	}

	// Try to parse end_time from params
	if et, ok := params["end_time"].(string); ok && et != "" {
		if parsed, err := time.Parse("15:04", et); err == nil {
			endTime = time.Date(now.Year(), now.Month(), now.Day(), parsed.Hour(), parsed.Minute(), 0, 0, now.Location())
		} else if parsed, err := time.Parse("2006-01-02 15:04", et); err == nil {
			endTime = parsed
		} else {
			endTime = now
		}
	} else {
		endTime = now
	}

	// Generate 5 time points evenly distributed
	duration := endTime.Sub(startTime)
	interval := duration / 4 // 5 points = 4 intervals

	timePoints := []string{}
	for i := 0; i < 5; i++ {
		t := startTime.Add(time.Duration(i) * interval)
		timePoints = append(timePoints, t.Format("15:04"))
	}

	records := []map[string]interface{}{}

	// Blood Pressure (split into SBP/DBP for separate series)
	bpValues := [][]int{
		{120, 80}, {118, 78}, {122, 82}, {119, 79}, {121, 81},
	}
	for i, ts := range timePoints {
		// Systolic BP (SBP)
		records = append(records, map[string]interface{}{
			"name":      "SBP",
			"value":     bpValues[i][0],
			"unit":      "mmHg",
			"group_key": "Blood Pressure",
			"ts":        ts,
		})
		// Diastolic BP (DBP)
		records = append(records, map[string]interface{}{
			"name":      "DBP",
			"value":     bpValues[i][1],
			"unit":      "mmHg",
			"group_key": "Blood Pressure",
			"ts":        ts,
		})
	}

	// Heart Rate
	hrValues := []int{75, 78, 72, 76, 74}
	for i, ts := range timePoints {
		records = append(records, map[string]interface{}{
			"name":      "HR",
			"value":     hrValues[i],
			"unit":      "bpm",
			"group_key": "Heart Rate",
			"ts":        ts,
		})
	}

	// Temperature
	tempValues := []float64{36.5, 36.7, 36.6, 36.8, 36.5}
	for i, ts := range timePoints {
		records = append(records, map[string]interface{}{
			"name":      "Temp",
			"value":     tempValues[i],
			"unit":      "°C",
			"group_key": "Temperature",
			"ts":        ts,
		})
	}

	// Blood Oxygen
	spo2Values := []int{98, 97, 98, 99, 98}
	for i, ts := range timePoints {
		records = append(records, map[string]interface{}{
			"name":      "SpO2",
			"value":     spo2Values[i],
			"unit":      "%",
			"group_key": "Blood Oxygen",
			"ts":        ts,
		})
	}

	return records
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
