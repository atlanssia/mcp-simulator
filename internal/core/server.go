package core

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/atlanssia/mcp-simulator/internal/infra/transport"
)

// BaseVirtualServer implements VirtualServer.
type BaseVirtualServer struct {
	config   ServerConfig
	registry Registry
	server   *http.Server
	status   string
	mu       sync.RWMutex
}

func NewBaseVirtualServer(config ServerConfig, registry Registry) *BaseVirtualServer {
	return &BaseVirtualServer{
		config:   config,
		registry: registry,
		status:   "stopped",
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
	// Placeholder for message handling
	w.WriteHeader(http.StatusOK)
}
