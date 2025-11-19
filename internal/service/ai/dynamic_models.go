package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// ModelCache caches fetched models to reduce API calls
type ModelCache struct {
	models    []ModelInfo
	timestamp time.Time
	mu        sync.RWMutex
}

var (
	siliconFlowCache = &ModelCache{}
	cacheDuration    = 5 * time.Minute
)

// SiliconFlow pricing map (maintained manually)
var siliconFlowFreePricing = map[string]bool{
	"Qwen/Qwen2.5-7B-Instruct":  true,
	"deepseek-ai/DeepSeek-V2.5": true,
	"THUDM/glm-4-9b-chat":       true,
}

// SiliconFlowModelResponse represents the API response
type SiliconFlowModelResponse struct {
	Object string `json:"object"`
	Data   []struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		OwnedBy string `json:"owned_by"`
	} `json:"data"`
}

// FetchSiliconFlowModels fetches models from SiliconFlow API
func FetchSiliconFlowModels(ctx context.Context, apiKey string) ([]ModelInfo, error) {
	// Check cache first
	siliconFlowCache.mu.RLock()
	if time.Since(siliconFlowCache.timestamp) < cacheDuration && len(siliconFlowCache.models) > 0 {
		models := siliconFlowCache.models
		siliconFlowCache.mu.RUnlock()
		return models, nil
	}
	siliconFlowCache.mu.RUnlock()

	// Fetch from API
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.siliconflow.cn/v1/models?sub_type=chat", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var apiResp SiliconFlowModelResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	// Convert to ModelInfo and merge with pricing data
	models := make([]ModelInfo, 0, len(apiResp.Data))
	for _, model := range apiResp.Data {
		isFree := siliconFlowFreePricing[model.ID]
		displayName := model.ID
		if isFree {
			displayName += " (免费)"
		}

		models = append(models, ModelInfo{
			Name:        model.ID,
			DisplayName: displayName,
			Free:        isFree,
		})
	}

	// Update cache
	siliconFlowCache.mu.Lock()
	siliconFlowCache.models = models
	siliconFlowCache.timestamp = time.Now()
	siliconFlowCache.mu.Unlock()

	return models, nil
}

// ModelScopeModelResponse represents the ModelScope API response
type ModelScopeModelResponse struct {
	Object string `json:"object"`
	Data   []struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		OwnedBy string `json:"owned_by"`
		Created int64  `json:"created"`
	} `json:"data"`
}

var (
	modelScopeCache = &ModelCache{}
)

// FetchModelScopeModels fetches models from ModelScope API (no auth required!)
func FetchModelScopeModels(ctx context.Context) ([]ModelInfo, error) {
	// Check cache first
	modelScopeCache.mu.RLock()
	if time.Since(modelScopeCache.timestamp) < cacheDuration && len(modelScopeCache.models) > 0 {
		models := modelScopeCache.models
		modelScopeCache.mu.RUnlock()
		return models, nil
	}
	modelScopeCache.mu.RUnlock()

	// Fetch from API (no auth required!)
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api-inference.modelscope.cn/v1/models", nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var apiResp ModelScopeModelResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	// Convert to ModelInfo
	// Note: ModelScope API doesn't include pricing info, all marked as paid
	models := make([]ModelInfo, 0, len(apiResp.Data))
	for _, model := range apiResp.Data {
		models = append(models, ModelInfo{
			Name:        model.ID,
			DisplayName: model.ID,
			Free:        false, // ModelScope doesn't provide pricing info
		})
	}

	// Update cache
	modelScopeCache.mu.Lock()
	modelScopeCache.models = models
	modelScopeCache.timestamp = time.Now()
	modelScopeCache.mu.Unlock()

	return models, nil
}

// ClearModelCache clears the model cache (useful for testing or manual refresh)
func ClearModelCache() {
	siliconFlowCache.mu.Lock()
	siliconFlowCache.models = nil
	siliconFlowCache.timestamp = time.Time{}
	siliconFlowCache.mu.Unlock()

	modelScopeCache.mu.Lock()
	modelScopeCache.models = nil
	modelScopeCache.timestamp = time.Time{}
	modelScopeCache.mu.Unlock()
}
