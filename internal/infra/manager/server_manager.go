package manager

import (
	"context"
	"fmt"
	"sync"

	"github.com/atlanssia/mcp-simulator/internal/core"
)

type ServerManager struct {
	servers   map[string]core.VirtualServer
	mu        sync.RWMutex
	generator core.MockGenerator
}

func NewServerManager(generator core.MockGenerator) *ServerManager {
	return &ServerManager{
		servers:   make(map[string]core.VirtualServer),
		generator: generator,
	}
}

// CreateServer creates a new virtual server from config
func (m *ServerManager) CreateServer(config core.ServerConfig) (core.VirtualServer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if server already exists
	if _, ok := m.servers[config.ID]; ok {
		return nil, fmt.Errorf("server already exists: %s", config.ID)
	}

	// Create new server with in-memory registry and AI generator
	registry := core.NewInMemoryRegistry()
	server := core.NewBaseVirtualServer(config, registry, m.generator)
	m.servers[config.ID] = server

	return server, nil
}

// GetServer retrieves a server by ID
func (m *ServerManager) GetServer(id string) (core.VirtualServer, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	server, ok := m.servers[id]
	return server, ok
}

func (m *ServerManager) AddServer(server core.VirtualServer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.servers[server.ID()] = server
}

func (m *ServerManager) StartServer(ctx context.Context, id string) error {
	m.mu.RLock()
	server, ok := m.servers[id]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("server not found: %s", id)
	}

	return server.Start(ctx)
}

func (m *ServerManager) StopServer(ctx context.Context, id string) error {
	m.mu.RLock()
	server, ok := m.servers[id]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("server not found: %s", id)
	}

	return server.Stop(ctx)
}

func (m *ServerManager) ListServers() []core.VirtualServer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []core.VirtualServer
	for _, s := range m.servers {
		list = append(list, s)
	}
	return list
}
