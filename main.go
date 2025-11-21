package main

import (
	"embed"
	"io/fs"
	stdLog "log"
	"net/http"

	"github.com/atlanssia/mcp-simulator/internal/infra/logger"
	"github.com/atlanssia/mcp-simulator/internal/infra/manager"
	"github.com/atlanssia/mcp-simulator/internal/infra/middleware"
	"github.com/atlanssia/mcp-simulator/internal/infra/storage"
	"github.com/atlanssia/mcp-simulator/internal/service/admin"
	"github.com/atlanssia/mcp-simulator/internal/service/ai"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

//go:embed web/dist
var staticFiles embed.FS

func main() {
	// Initialize Storage
	store, err := storage.NewFileStorage("./data")
	if err != nil {
		stdLog.Fatalf("Failed to initialize storage: %v", err)
	}
	stdLog.Println("Storage initialized in ./data directory")

	// Load LLM Config
	llmConfig, err := store.LoadLLMConfig()
	if err != nil {
		stdLog.Printf("Warning: Failed to load LLM config: %v", err)
	}

	// Initialize logger
	var loggerInstance *zap.Logger
	loggerInstance, err = logger.NewLogger("logs/admin.log")
	if err != nil {
		stdLog.Fatalf("Failed to initialize logger: %v", err)
	}
	defer loggerInstance.Sync()

	// Initialize AI Generator (needed by ServerManager)
	aiGenerator := ai.NewGenerator(llmConfig, loggerInstance)

	// Initialize Server Manager with AI generator
	serverManager := manager.NewServerManager(aiGenerator)

	// Load persisted servers and restore state
	servers, err := store.LoadServers()
	if err != nil {
		stdLog.Printf("Warning: Failed to load servers: %v", err)
	} else {
		stdLog.Printf("Loaded %d server(s) from storage", len(servers))
		for _, serverConfig := range servers {
			server, err := serverManager.CreateServer(serverConfig)
			if err != nil {
				stdLog.Printf("Warning: Failed to restore server %s: %v", serverConfig.ID, err)
				continue
			}

			// Load tools for this server
			tools, err := store.LoadTools(serverConfig.ID)
			if err != nil {
				stdLog.Printf("Warning: Failed to load tools for server %s: %v", serverConfig.ID, err)
			} else {
				stdLog.Printf("Loaded %d tool(s) for server %s", len(tools), serverConfig.ID)
				for _, tool := range tools {
					if err := server.GetRegistry().RegisterTool(tool); err != nil {
						stdLog.Printf("Warning: Failed to register tool %s: %v", tool.Name, err)
					}
				}
			}
		}
	}

	// Initialize Admin Handler
	adminHandler := admin.NewHandler(serverManager, aiGenerator, loggerInstance)

	// Setup router
	router := gin.New()
	router.Use(gin.Recovery())

	// Add Access Log middleware
	router.Use(func(c *gin.Context) {
		// Create a handler that wraps the next gin handler
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Next()
		})

		// Use the middleware
		middleware.AccessLog(loggerInstance)(next).ServeHTTP(c.Writer, c.Request)
	})

	// Register Admin Routes
	adminHandler.RegisterRoutes(router)

	// Serve static files from embedded filesystem
	staticFS, err := fs.Sub(staticFiles, "web/dist")
	if err != nil {
		loggerInstance.Fatal("Failed to create sub filesystem", zap.Error(err))
	}

	// Serve all static files (assets, vite.svg, etc.)
	router.Use(func(c *gin.Context) {
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
	loggerInstance.Info("Starting Admin Server on :8080")
	if err := router.Run(":8080"); err != nil {
		loggerInstance.Fatal("Failed to start server", zap.Error(err))
	}
}
