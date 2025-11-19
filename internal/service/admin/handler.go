package admin

import (
	"encoding/json"
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

	// Wrap in MCP Tool Result structure
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal result"})
		return
	}

	toolResult := core.CallToolResult{
		Content: []core.Content{
			{
				Type: "text",
				Text: string(jsonBytes),
			},
		},
		IsError: false,
	}

	c.JSON(http.StatusOK, toolResult)
}
