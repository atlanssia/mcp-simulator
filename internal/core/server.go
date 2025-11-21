package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/atlanssia/mcp-simulator/internal/infra/transport"
)

// MockGenerator interface for generating mock data (avoids circular dependency)
type MockGenerator interface {
	GenerateMockResponseWithParams(
		ctx context.Context,
		toolName, toolDescription string,
		inputSchema, sampleParams map[string]interface{},
		params GenerationParams,
	) (interface{}, error)
	GetDefaultModel() string
}

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ToolsCallParams represents parameters for tools/call
type ToolsCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// BaseVirtualServer implements VirtualServer.
type BaseVirtualServer struct {
	config    ServerConfig
	registry  Registry
	server    *http.Server
	status    string
	mu        sync.RWMutex
	generator MockGenerator
}

func NewBaseVirtualServer(config ServerConfig, registry Registry, generator MockGenerator) *BaseVirtualServer {
	return &BaseVirtualServer{
		config:    config,
		registry:  registry,
		status:    "stopped",
		generator: generator,
	}
}

func (s *BaseVirtualServer) ID() string {
	return s.config.ID
}

func (s *BaseVirtualServer) Config() ServerConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	config := s.config
	config.Status = s.status
	return config
}

func (s *BaseVirtualServer) Status() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

func (s *BaseVirtualServer) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status == "running"
}

func (s *BaseVirtualServer) GetRegistry() Registry {
	return s.registry
}

func (s *BaseVirtualServer) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.status == "running" {
		s.mu.Unlock()
		return fmt.Errorf("server already running")
	}
	s.status = "starting"
	s.mu.Unlock()

	mux := http.NewServeMux()
	mux.HandleFunc("/sse", s.handleSSE)
	mux.HandleFunc("/messages", s.handleMessages)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Port),
		Handler: mux,
	}

	log.Printf("Starting Virtual Server %s on port %d", s.config.Name, s.config.Port)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server %s error: %v", s.config.Name, err)
			s.mu.Lock()
			s.status = "error"
			s.mu.Unlock()
		}
	}()

	s.mu.Lock()
	s.status = "running"
	s.mu.Unlock()
	return nil
}

func (s *BaseVirtualServer) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server != nil {
		err := s.server.Shutdown(ctx)
		s.status = "stopped"
		return err
	}
	s.status = "stopped"
	return nil
}

func (s *BaseVirtualServer) handleSSE(w http.ResponseWriter, r *http.Request) {
	adapter := transport.NewSSEAdapter()
	adapter.HandleSSE(w, r)
}

func (s *BaseVirtualServer) handleMessages(w http.ResponseWriter, r *http.Request) {
	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, nil, -32700, "Parse error")
		return
	}

	// Route to appropriate method
	var result interface{}
	var err error

	switch req.Method {
	case "tools/list":
		result, err = s.handleToolsList(r.Context())
	case "tools/call":
		result, err = s.handleToolsCall(r.Context(), req.Params)
	default:
		s.sendError(w, req.ID, -32601, fmt.Sprintf("Method not found: %s", req.Method))
		return
	}

	if err != nil {
		s.sendError(w, req.ID, -32603, err.Error())
		return
	}

	s.sendResult(w, req.ID, result)
}

// handleToolsList returns list of available tools
func (s *BaseVirtualServer) handleToolsList(ctx context.Context) (interface{}, error) {
	tools := s.registry.ListTools()

	// Convert to MCP format
	mcpTools := make([]map[string]interface{}, len(tools))
	for i, tool := range tools {
		mcpTools[i] = map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		}
	}

	return map[string]interface{}{
		"tools": mcpTools,
	}, nil
}

// handleToolsCall invokes a tool and generates mock data
func (s *BaseVirtualServer) handleToolsCall(ctx context.Context, paramsRaw json.RawMessage) (interface{}, error) {
	var params ToolsCallParams
	if err := json.Unmarshal(paramsRaw, &params); err != nil {
		return nil, fmt.Errorf("invalid params: %v", err)
	}

	// Get tool from registry
	tool, ok := s.registry.GetTool(params.Name)
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", params.Name)
	}

	// Generate mock data using AI
	mockData, err := s.generator.GenerateMockResponseWithParams(
		ctx,
		tool.Name,
		tool.Description,
		tool.InputSchema,
		params.Arguments,
		GenerationParams{
			Model:        s.generator.GetDefaultModel(),
			Temperature:  0.7,
			SystemPrompt: "",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate mock data: %v", err)
	}

	// Return in MCP format
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": toJSONString(mockData),
			},
		},
	}, nil
}

func (s *BaseVirtualServer) sendResult(w http.ResponseWriter, id interface{}, result interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *BaseVirtualServer) sendError(w http.ResponseWriter, id interface{}, code int, message string) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // JSON-RPC errors still return 200
	json.NewEncoder(w).Encode(resp)
}

func toJSONString(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}
