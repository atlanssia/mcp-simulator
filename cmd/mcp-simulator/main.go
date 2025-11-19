package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/atlanssia/mcp-simulator/internal/infra/manager"
	"github.com/atlanssia/mcp-simulator/internal/service/admin"
	"github.com/atlanssia/mcp-simulator/internal/service/ai"
	"github.com/gin-gonic/gin"
)

//go:embed web/dist
var staticFiles embed.FS

func main() {
	// Initialize Server Manager
	serverManager := manager.NewServerManager()

	// Initialize AI Generator
	aiGenerator := ai.NewGenerator("") // Empty string to use env var

	// Initialize Admin Handler
	adminHandler := admin.NewHandler(serverManager, aiGenerator)

	// Setup Gin Router
	r := gin.Default()

	// Register Admin Routes
	adminHandler.RegisterRoutes(r)

	// Serve static files from embedded filesystem
	staticFS, err := fs.Sub(staticFiles, "web/dist")
	if err != nil {
		log.Fatalf("Failed to create sub filesystem: %v", err)
	}
	r.StaticFS("/assets", http.FS(staticFS))
	r.NoRoute(func(c *gin.Context) {
		data, err := staticFiles.ReadFile("web/dist/index.html")
		if err != nil {
			c.String(http.StatusNotFound, "404 page not found")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	// Start Admin Server
	log.Println("Starting Admin Server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start admin server: %v", err)
	}
}
