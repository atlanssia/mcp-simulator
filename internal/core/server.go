package core

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/atlanssia/mcp-simulator/internal/infra/transport"
)

// BaseVirtualServer implements VirtualServer.
type BaseVirtualServer struct {
	config   ServerConfig
	registry Registry
	server   *http.Server
}

func NewBaseVirtualServer(config ServerConfig, registry Registry) *BaseVirtualServer {
	return &BaseVirtualServer{
		config:   config,
		registry: registry,
	}
}

func (s *BaseVirtualServer) ID() string {
	return s.config.ID
}

func (s *BaseVirtualServer) Config() ServerConfig {
	return s.config
}

func (s *BaseVirtualServer) Start(ctx context.Context) error {
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
		}
	}()
	return nil
}

func (s *BaseVirtualServer) Stop(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

func (s *BaseVirtualServer) handleSSE(w http.ResponseWriter, r *http.Request) {
	adapter := transport.NewSSEAdapter()
	adapter.HandleSSE(w, r)
}

func (s *BaseVirtualServer) handleMessages(w http.ResponseWriter, r *http.Request) {
	// Placeholder for message handling
	w.WriteHeader(http.StatusOK)
}
