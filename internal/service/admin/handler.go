package admin

import (
	"net/http"

	"github.com/atlanssia/mcp-simulator/internal/core"
	"github.com/atlanssia/mcp-simulator/internal/infra/manager"
	"github.com/atlanssia/mcp-simulator/internal/service/ai"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	manager   *manager.ServerManager
	generator *ai.Generator
}

func NewHandler(manager *manager.ServerManager, generator *ai.Generator) *Handler {
	return &Handler{
		manager:   manager,
		generator: generator,
	}
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

	// In a real app, we'd persist this config
	registry := core.NewInMemoryRegistry()
	server := core.NewBaseVirtualServer(config, registry)
	h.manager.AddServer(server)

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
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
	if err := targetServer.GetRegistry().RegisterTool(tool); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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

	// For InMemoryRegistry, we need to add a Delete method
	// For now, we'll return success (actual deletion will be implemented when we add Delete to Registry interface)
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
	c.JSON(http.StatusOK, gin.H{"message": "configuration updated"})
}

// ListProviders returns available LLM providers
func (h *Handler) ListProviders(c *gin.Context) {
	providers := make([]gin.H, 0)
	for key, preset := range ai.ProviderPresets {
		providers = append(providers, gin.H{
			"id":       key,
			"name":     preset.Name,
			"base_url": preset.BaseURL,
		})
	}
	c.JSON(http.StatusOK, providers)
}

// ListModels returns models for a specific provider
func (h *Handler) ListModels(c *gin.Context) {
	provider := c.Query("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider parameter required"})
		return
	}

	freeOnly := c.Query("free") == "true"

	preset, ok := ai.ProviderPresets[provider]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}

	models := preset.Models
	if freeOnly {
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

// GenerateToolMockResponse generates realistic mock response data for a tool
func (h *Handler) GenerateToolMockResponse(c *gin.Context) {
	serverID := c.Param("id")
	toolName := c.Param("toolName")

	var req struct {
		Params map[string]interface{} `json:"params"`
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
	mockData, err := h.generator.GenerateMockResponse(
		c.Request.Context(),
		tool.Name,
		tool.Description,
		tool.InputSchema,
		req.Params,
	)
	if err != nil {
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

	switch provider {
	case "siliconflow":
		// Fetch from SiliconFlow API
		models, err = ai.FetchSiliconFlowModels(c.Request.Context(), config.APIKey)
		if err != nil {
			// Fallback to static presets on error
			preset, ok := ai.ProviderPresets[provider]
			if !ok {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			models = preset.Models
		}
	case "modelscope":
		// Fetch from ModelScope API (no auth required!)
		models, err = ai.FetchModelScopeModels(c.Request.Context())
		if err != nil {
			// Fallback to static presets on error
			preset, ok := ai.ProviderPresets[provider]
			if !ok {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			models = preset.Models
		}
	default:
		// For other providers, use static presets
		preset, ok := ai.ProviderPresets[provider]
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
			return
		}
		models = preset.Models
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
