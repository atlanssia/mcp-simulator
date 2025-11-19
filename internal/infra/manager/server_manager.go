package manager

import (
	"context"
	"fmt"
	"sync"

	"github.com/atlanssia/mcp-simulator/internal/core"
)

type ServerManager struct {
	servers map[string]core.VirtualServer
	mu      sync.RWMutex
}

func NewServerManager() *ServerManager {
	return &ServerManager{
		servers: make(map[string]core.VirtualServer),
	}
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
