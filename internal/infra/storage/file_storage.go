package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/atlanssia/mcp-simulator/internal/core"
)

// Storage defines the interface for persisting application data
type Storage interface {
	// LLM Configuration
	SaveLLMConfig(config core.LLMConfig) error
	LoadLLMConfig() (core.LLMConfig, error)

	// Server Management
	SaveServer(server core.ServerConfig) error
	LoadServers() ([]core.ServerConfig, error)
	DeleteServer(id string) error

	// Tool Management (per-server)
	SaveTools(serverID string, tools []core.Tool) error
	LoadTools(serverID string) ([]core.Tool, error)
}

// FileStorage implements Storage using JSON files
type FileStorage struct {
	dataDir string
	mu      sync.RWMutex
}

// NewFileStorage creates a new file-based storage
func NewFileStorage(dataDir string) (*FileStorage, error) {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Create tools subdirectory
	toolsDir := filepath.Join(dataDir, "tools")
	if err := os.MkdirAll(toolsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create tools directory: %w", err)
	}

	return &FileStorage{
		dataDir: dataDir,
	}, nil
}

// SaveLLMConfig saves LLM configuration to config.json
func (fs *FileStorage) SaveLLMConfig(config core.LLMConfig) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	configPath := filepath.Join(fs.dataDir, "config.json")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadLLMConfig loads LLM configuration from config.json
func (fs *FileStorage) LoadLLMConfig() (core.LLMConfig, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	configPath := filepath.Join(fs.dataDir, "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return core.LLMConfig{
				Provider: "openai",
			}, nil
		}
		return core.LLMConfig{}, fmt.Errorf("failed to read config file: %w", err)
	}

	var config core.LLMConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return core.LLMConfig{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}

// SaveServer saves a single server configuration to servers.json
func (fs *FileStorage) SaveServer(server core.ServerConfig) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Load existing servers
	servers, _ := fs.loadServersUnlocked()

	// Update or append server
	found := false
	for i, s := range servers {
		if s.ID == server.ID {
			servers[i] = server
			found = true
			break
		}
	}
	if !found {
		servers = append(servers, server)
	}

	// Save back to file
	return fs.saveServersUnlocked(servers)
}

// LoadServers loads all server configurations from servers.json
func (fs *FileStorage) LoadServers() ([]core.ServerConfig, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	return fs.loadServersUnlocked()
}

// DeleteServer removes a server configuration from servers.json
func (fs *FileStorage) DeleteServer(id string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	servers, _ := fs.loadServersUnlocked()

	// Filter out the server to delete
	filtered := make([]core.ServerConfig, 0)
	for _, s := range servers {
		if s.ID != id {
			filtered = append(filtered, s)
		}
	}

	// Also delete the server's tools file
	toolsPath := filepath.Join(fs.dataDir, "tools", id+".json")
	_ = os.Remove(toolsPath) // Ignore error if file doesn't exist

	return fs.saveServersUnlocked(filtered)
}

// SaveTools saves tools for a specific server to tools/{serverID}.json
func (fs *FileStorage) SaveTools(serverID string, tools []core.Tool) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	toolsPath := filepath.Join(fs.dataDir, "tools", serverID+".json")
	data, err := json.MarshalIndent(tools, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tools: %w", err)
	}

	if err := os.WriteFile(toolsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write tools file: %w", err)
	}

	return nil
}

// LoadTools loads tools for a specific server from tools/{serverID}.json
func (fs *FileStorage) LoadTools(serverID string) ([]core.Tool, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	toolsPath := filepath.Join(fs.dataDir, "tools", serverID+".json")
	data, err := os.ReadFile(toolsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty slice if file doesn't exist
			return []core.Tool{}, nil
		}
		return nil, fmt.Errorf("failed to read tools file: %w", err)
	}

	var tools []core.Tool
	if err := json.Unmarshal(data, &tools); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tools: %w", err)
	}

	return tools, nil
}

// Helper methods (unlocked versions for internal use)

func (fs *FileStorage) loadServersUnlocked() ([]core.ServerConfig, error) {
	serversPath := filepath.Join(fs.dataDir, "servers.json")
	data, err := os.ReadFile(serversPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []core.ServerConfig{}, nil
		}
		return nil, fmt.Errorf("failed to read servers file: %w", err)
	}

	var servers []core.ServerConfig
	if err := json.Unmarshal(data, &servers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal servers: %w", err)
	}

	return servers, nil
}

func (fs *FileStorage) saveServersUnlocked(servers []core.ServerConfig) error {
	serversPath := filepath.Join(fs.dataDir, "servers.json")
	data, err := json.MarshalIndent(servers, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal servers: %w", err)
	}

	if err := os.WriteFile(serversPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write servers file: %w", err)
	}

	return nil
}
