package core

import (
	"sync"
)

// InMemoryRegistry implements Registry with thread-safe maps.
type InMemoryRegistry struct {
	tools     map[string]Tool
	resources map[string]Resource
	prompts   map[string]Prompt
	mu        sync.RWMutex
}

func NewInMemoryRegistry() *InMemoryRegistry {
	return &InMemoryRegistry{
		tools:     make(map[string]Tool),
		resources: make(map[string]Resource),
		prompts:   make(map[string]Prompt),
	}
}

func (r *InMemoryRegistry) RegisterTool(tool Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[tool.Name] = tool
	return nil
}

func (r *InMemoryRegistry) RegisterResource(resource Resource) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.resources[resource.URI] = resource
	return nil
}

func (r *InMemoryRegistry) RegisterPrompt(prompt Prompt) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.prompts[prompt.Name] = prompt
	return nil
}

func (r *InMemoryRegistry) GetTool(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[name]
	return t, ok
}

// Add other getters as needed
