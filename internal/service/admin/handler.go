package admin

import (
	"net/http"

	"github.com/atlanssia/mcp-simulator/internal/core"
	"github.com/atlanssia/mcp-simulator/internal/infra/manager"
	"github.com/atlanssia/mcp-simulator/internal/infra/storage"
	"github.com/atlanssia/mcp-simulator/internal/service/ai"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	manager   *manager.ServerManager
	generator *ai.Generator
	storage   storage.Storage
	logger    *zap.Logger
}

func NewHandler(manager *manager.ServerManager, generator *ai.Generator, logger *zap.Logger) *Handler {
	return &Handler{
		manager:   manager,
		generator: generator,
		logger:    logger,
	}
}

// SetStorage sets the storage backend for persistence
func (h *Handler) SetStorage(s storage.Storage) {
	h.storage = s
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.GET("/servers", h.ListServers)
		api.POST("/servers", h.CreateServer)
		api.POST("/servers/:id/start", h.StartServer)
		api.POST("/servers/:id/stop", h.StopServer)

		// Tool management endpoints
		api.GET("/servers/:id/tools", h.ListTools)
		api.POST("/servers/:id/tools", h.CreateTool)
		api.PUT("/servers/:id/tools/:toolName", h.UpdateTool)
		api.DELETE("/servers/:id/tools/:toolName", h.DeleteTool)
		api.POST("/servers/:id/tools/:toolName/generate-mock", h.GenerateToolMockResponse)

		// LLM configuration endpoints
		api.GET("/config/llm", h.GetLLMConfig)
		api.POST("/config/llm", h.UpdateLLMConfig)
		api.GET("/config/llm/providers", h.ListProviders)
		api.GET("/config/llm/models", h.ListModels)
		api.GET("/config/llm/models/dynamic", h.ListDynamicModels)

		// AI generation
		api.POST("/ai/generate", h.GenerateMock)
		api.POST("/ai/generate-schema", h.GenerateSchema)
	}
}

func (h *Handler) ListServers(c *gin.Context) {
	servers := h.manager.ListServers()
	configs := make([]core.ServerConfig, len(servers))
	for i, s := range servers {
		configs[i] = s.Config()
	}
	c.JSON(http.StatusOK, configs)
}

func (h *Handler) CreateServer(c *gin.Context) {
	var config core.ServerConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use ServerManager's CreateServer method
	_, err := h.manager.CreateServer(config)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Persist to storage
	if h.storage != nil {
		if err := h.storage.SaveServer(config); err != nil {
			// Log error but don't fail the request
			c.Header("X-Warning", "Failed to persist server: "+err.Error())
		}
	}

	c.JSON(http.StatusCreated, config)
}

func (h *Handler) StartServer(c *gin.Context) {
	id := c.Param("id")
	if err := h.manager.StartServer(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "started"})
}

func (h *Handler) StopServer(c *gin.Context) {
	id := c.Param("id")
	if err := h.manager.StopServer(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "stopped"})
}

type GenerateRequest struct {
	Prompt string                 `json:"prompt"`
	Schema map[string]interface{} `json:"schema"`
}

func (h *Handler) GenerateMock(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If schema is missing, we might want to provide a default or error.
	// For now, let's assume the user provides it or we pass nil (which might fail in generator).
	// The user request implies "Generate a list of 5 recent orders", so maybe the schema is implicit?
	// But the generator requires it. Let's assume the frontend sends it.

	result, err := h.generator.GenerateMockData(c.Request.Context(), req.Prompt, req.Schema)
	if err != nil {
		h.logger.Error("Failed to generate mock data", zap.Error(err), zap.String("prompt", req.Prompt))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

type GenerateSchemaRequest struct {
	Description string                `json:"description"`
	Params      core.GenerationParams `json:"params"`
}

func (h *Handler) GenerateSchema(c *gin.Context) {
	var req GenerateSchemaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use default params if not provided
	if req.Params.Model == "" {
		req.Params = core.DefaultGenerationParams()
	}

	// Log the request parameters for debugging
	h.logger.Info("Generating schema",
		zap.String("description", req.Description),
		zap.String("model", req.Params.Model),
		zap.Float64("temperature", req.Params.Temperature),
		zap.String("system_prompt", req.Params.SystemPrompt),
	)

	schema, err := h.generator.GenerateToolSchemaWithParams(c.Request.Context(), req.Description, req.Params)
	if err != nil {
		h.logger.Error("Failed to generate schema",
			zap.Error(err),
			zap.String("description", req.Description),
			zap.String("model", req.Params.Model),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, schema)
}

// ListTools returns all tools for a specific server
func (h *Handler) ListTools(c *gin.Context) {
	serverID := c.Param("id")

	servers := h.manager.ListServers()
	var targetServer core.VirtualServer
	for _, s := range servers {
		if s.ID() == serverID {
			targetServer = s
			break
		}
	}

	if targetServer == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
		return
	}

	tools := targetServer.GetRegistry().ListTools()
	c.JSON(http.StatusOK, tools)
}

// CreateTool adds a new tool to a server
func (h *Handler) CreateTool(c *gin.Context) {
	serverID := c.Param("id")

	var tool core.Tool
	if err := c.ShouldBindJSON(&tool); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	servers := h.manager.ListServers()
	var targetServer core.VirtualServer
	for _, s := range servers {
		if s.ID() == serverID {
			targetServer = s
			break
		}
	}

	if targetServer == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
		return
	}

	if err := targetServer.GetRegistry().RegisterTool(tool); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Persist tools to storage
	if h.storage != nil {
		tools := targetServer.GetRegistry().ListTools()
		if err := h.storage.SaveTools(serverID, tools); err != nil {
			c.Header("X-Warning", "Failed to persist tools: "+err.Error())
		}
	}

	// Notify server to update MCP tools
	targetServer.UpdateTools()

	c.JSON(http.StatusCreated, tool)
}

// UpdateTool updates an existing tool
func (h *Handler) UpdateTool(c *gin.Context) {
	serverID := c.Param("id")
	toolName := c.Param("toolName")

	var tool core.Tool
	if err := c.ShouldBindJSON(&tool); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure tool name matches URL parameter
	tool.Name = toolName

	servers := h.manager.ListServers()
	var targetServer core.VirtualServer
	for _, s := range servers {
		if s.ID() == serverID {
			targetServer = s
			break
		}
	}

	if targetServer == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
		return
	}

	// Check if tool exists
	if _, ok := targetServer.GetRegistry().GetTool(toolName); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "tool not found"})
		return
	}

	// Update by re-registering
	if err := targetServer.GetRegistry().UpdateTool(toolName, tool); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Persist tools to storage
	if h.storage != nil {
		tools := targetServer.GetRegistry().ListTools()
		if err := h.storage.SaveTools(serverID, tools); err != nil {
			c.Header("X-Warning", "Failed to persist tools: "+err.Error())
		}
	}

	// Notify server to update MCP tools
	targetServer.UpdateTools()

	c.JSON(http.StatusOK, tool)
}

// DeleteTool removes a tool from a server
func (h *Handler) DeleteTool(c *gin.Context) {
	serverID := c.Param("id")
	toolName := c.Param("toolName")

	servers := h.manager.ListServers()
	var targetServer core.VirtualServer
	for _, s := range servers {
		if s.ID() == serverID {
			targetServer = s
			break
		}
	}

	if targetServer == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
		return
	}

	// Check if tool exists
	if _, ok := targetServer.GetRegistry().GetTool(toolName); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "tool not found"})
		return
	}

	// Delete the tool
	if err := targetServer.GetRegistry().DeleteTool(toolName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Persist tools to storage
	if h.storage != nil {
		tools := targetServer.GetRegistry().ListTools()
		if err := h.storage.SaveTools(serverID, tools); err != nil {
			c.Header("X-Warning", "Failed to persist tools: "+err.Error())
		}
	}

	// Notify server to update MCP tools
	targetServer.UpdateTools()

	c.JSON(http.StatusOK, gin.H{"message": "tool deleted"})
}

// GetLLMConfig returns the current LLM configuration
func (h *Handler) GetLLMConfig(c *gin.Context) {
	config := h.generator.GetConfig()
	c.JSON(http.StatusOK, config)
}

// UpdateLLMConfig updates the LLM configuration
func (h *Handler) UpdateLLMConfig(c *gin.Context) {
	var config core.LLMConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.generator.UpdateConfig(config)

	// Persist LLM config to storage
	if h.storage != nil {
		if err := h.storage.SaveLLMConfig(config); err != nil {
			c.Header("X-Warning", "Failed to persist LLM config: "+err.Error())
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "configuration updated"})
}

// ListProviders returns available LLM providers with metadata
func (h *Handler) ListProviders(c *gin.Context) {
	providers := ai.ListProviders()
	c.JSON(http.StatusOK, providers)
}

// ListModels returns models for a specific provider
// NOTE: This endpoint is deprecated in favor of dynamic model fetching
// Keeping for backward compatibility, but returns empty list
func (h *Handler) ListModels(c *gin.Context) {
	provider := c.Query("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider parameter required"})
		return
	}

	// Verify provider exists
	_, ok := ai.GetProviderInfo(provider)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}

	// Return empty list - clients should use dynamic model fetching
	c.JSON(http.StatusOK, []ai.ModelInfo{})
}

// GenerateToolMockResponse generates realistic mock response data for a tool
func (h *Handler) GenerateToolMockResponse(c *gin.Context) {
	serverID := c.Param("id")
	toolName := c.Param("toolName")

	var req struct {
		Params     map[string]interface{} `json:"params"`
		Generation core.GenerationParams  `json:"generation"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find the server
	servers := h.manager.ListServers()
	var targetServer core.VirtualServer
	for _, s := range servers {
		if s.Config().ID == serverID {
			targetServer = s
			break
		}
	}

	if targetServer == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
		return
	}

	// Get the tool
	tool, ok := targetServer.GetRegistry().GetTool(toolName)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "tool not found"})
		return
	}

	// Generate mock response
	// Use default params if not provided
	if req.Generation.Model == "" {
		req.Generation = core.DefaultGenerationParams()
	}

	mockData, err := h.generator.GenerateMockResponseWithParams(
		c.Request.Context(),
		tool.Name,
		tool.Description,
		tool.InputSchema,
		req.Params,
		req.Generation,
	)
	if err != nil {
		h.logger.Error("Failed to generate tool mock response",
			zap.Error(err),
			zap.String("tool", toolName),
			zap.String("model", req.Generation.Model),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mockData)
}

// ListDynamicModels returns models fetched dynamically from provider APIs
func (h *Handler) ListDynamicModels(c *gin.Context) {
	provider := c.Query("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider parameter required"})
		return
	}

	// Get current config to use API key
	config := h.generator.GetConfig()

	var models []ai.ModelInfo
	var err error

	// Get provider info to check if it supports dynamic models
	providerInfo, ok := ai.GetProviderInfo(provider)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}

	// Only fetch if provider supports dynamic models
	if !providerInfo.SupportsModels {
		c.JSON(http.StatusOK, []ai.ModelInfo{})
		return
	}

	switch provider {
	case "siliconflow":
		// Fetch from SiliconFlow API
		models, err = ai.FetchSiliconFlowModels(c.Request.Context(), config.APIKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	case "modelscope":
		// Fetch from ModelScope API (no auth required!)
		models, err = ai.FetchModelScopeModels(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	default:
		// Provider does not support dynamic fetching
		c.JSON(http.StatusOK, []ai.ModelInfo{})
		return
	}

	// Apply free filter if requested
	if c.Query("free") == "true" {
		filtered := make([]ai.ModelInfo, 0)
		for _, model := range models {
			if model.Free {
				filtered = append(filtered, model)
			}
		}
		models = filtered
	}

	c.JSON(http.StatusOK, models)
}
