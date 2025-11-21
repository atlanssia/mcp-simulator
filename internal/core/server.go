package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/atlanssia/mcp-simulator/internal/infra/logger"
	"github.com/atlanssia/mcp-simulator/internal/infra/middleware"
	"go.uber.org/zap"
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
	mcpServer *MCPServerWrapper // NEW: mcp-go server wrapper
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

	// Create official MCP SDK server wrapper
	var err error
	s.mcpServer, err = NewMCPServerWrapper(
		s.config.Name,
		s.registry,
		s.generator,
	)
	if err != nil {
		s.mu.Lock()
		s.status = "error"
		s.mu.Unlock()
		return fmt.Errorf("failed to create MCP server: %w", err)
	}

	// Get HTTP handler from official MCP SDK
	// This automatically handles:
	// - GET / : discovery
	// - GET / with Accept:text/event-stream : SSE connection + endpoint event
	// - POST / : JSON-RPC requests (with SSE response if session exists)
	mcpHandler := s.mcpServer.GetHTTPHandler()

	// Create router
	mux := http.NewServeMux()

	// Mount official MCP SDK handler directly at root
	// No need for SSEWrapperHandler - SDK handles everything automatically
	mux.Handle("/", mcpHandler)

	// Optional: Keep /mcp/sse alias for backward compatibility
	mux.Handle("/mcp/sse", mcpHandler)
	mux.Handle("/messages", mcpHandler)

	// Initialize logger for this virtual server
	logFile := fmt.Sprintf("logs/%s.log", s.config.ID)
	logger, err := logger.NewLogger(logFile)
	if err != nil {
		log.Printf("Failed to initialize logger for server %s: %v", s.config.ID, err)
		// Fallback to default logger if file init fails
		logger, _ = zap.NewProduction()
	}

	// Apply Access Log middleware with this server's logger
	handler := middleware.AccessLog(logger)(mux)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Port),
		Handler: handler,
	}

	logger.Info("Starting Virtual Server",
		zap.String("name", s.config.Name),
		zap.Int("port", s.config.Port),
	)

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", zap.Error(err))
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

func (s *BaseVirtualServer) UpdateTools() {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.mcpServer != nil {
		s.mcpServer.UpdateTools()
	}
}
