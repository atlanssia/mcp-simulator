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
	aiGenerator := ai.NewGenerator("")

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

	// Serve all static files (assets, vite.svg, etc.)
	r.Use(func(c *gin.Context) {
		path := c.Request.URL.Path

		// Skip API routes
		if len(path) >= 4 && path[:4] == "/api" {
			c.Next()
			return
		}

		// Try to serve the file from embedded FS
		if path == "/" {
			path = "/index.html"
		}

		// Remove leading slash for fs.FS
		filePath := path[1:]
		data, err := fs.ReadFile(staticFS, filePath)
		if err == nil {
			// Determine content type
			contentType := "text/html; charset=utf-8"
			if len(path) > 3 {
				ext := path[len(path)-3:]
				switch ext {
				case ".js":
					contentType = "application/javascript"
				case "css":
					contentType = "text/css"
				case "svg":
					contentType = "image/svg+xml"
				case "png":
					contentType = "image/png"
				case "jpg", "peg":
					contentType = "image/jpeg"
				}
			}
			c.Data(http.StatusOK, contentType, data)
			c.Abort()
			return
		}

		// File not found, serve index.html for SPA routing
		if data, err := fs.ReadFile(staticFS, "index.html"); err == nil {
			c.Data(http.StatusOK, "text/html; charset=utf-8", data)
			c.Abort()
			return
		}

		c.Next()
	})

	// Start Admin Server
	log.Println("Starting Admin Server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
