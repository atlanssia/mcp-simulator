package main

import (
	"log"

	"github.com/atlanssia/mcp-simulator/internal/infra/manager"
	"github.com/atlanssia/mcp-simulator/internal/service/admin"
	"github.com/atlanssia/mcp-simulator/internal/service/ai"
	"github.com/gin-gonic/gin"
)

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

	// Start Admin Server
	log.Println("Starting Admin Server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start admin server: %v", err)
	}
}
