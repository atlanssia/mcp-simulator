package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/atlanssia/mcp-simulator/internal/core"
)

type Generator struct {
	config core.LLMConfig
	client *http.Client
	mu     sync.RWMutex
}

func NewGenerator(apiKey string) *Generator {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}

	config := core.DefaultLLMConfig()
	if apiKey != "" {
		config.APIKey = apiKey
	}

	return &Generator{
		config: config,
		client: &http.Client{},
	}
}

// UpdateConfig updates the LLM configuration
func (g *Generator) UpdateConfig(config core.LLMConfig) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.config = config
}

// GetConfig returns the current LLM configuration
func (g *Generator) GetConfig() core.LLMConfig {
	g.mu.RLock()
	defer g.mu.RUnlock()
	// Return a copy with masked API key
	config := g.config
	if len(config.APIKey) > 8 {
		config.APIKey = config.APIKey[:4] + "****" + config.APIKey[len(config.APIKey)-4:]
	}
	return config
}

type CompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CompletionResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func (g *Generator) GenerateMockData(ctx context.Context, prompt string, schema map[string]interface{}) (interface{}, error) {
	g.mu.RLock()
	config := g.config
	g.mu.RUnlock()

	// Replace {description} placeholder in system prompt
	systemPrompt := strings.ReplaceAll(config.SystemPrompt, "{description}", prompt)

	reqBody := CompletionRequest{
		Model: config.Model,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
		Temperature: config.Temperature,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	// Use configured base URL
	url := config.BaseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var completion CompletionResponse
	if err := json.Unmarshal(body, &completion); err != nil {
		return nil, err
	}

	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned")
	}

	content := completion.Choices[0].Message.Content

	// Try to extract JSON from markdown code blocks if present
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	var result interface{}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse generated JSON: %v, content: %s", err, content)
	}

	return result, nil
}

// GenerateMockResponse generates realistic mock response data for a tool
// based on its description and input schema
func (g *Generator) GenerateMockResponse(ctx context.Context, toolName, toolDescription string, inputSchema, sampleParams map[string]interface{}) (interface{}, error) {
	g.mu.RLock()
	config := g.config
	g.mu.RUnlock()

	// Build prompt for generating mock response data
	schemaBytes, _ := json.Marshal(inputSchema)
	paramsBytes, _ := json.Marshal(sampleParams)

	prompt := fmt.Sprintf(`Generate realistic mock response data for the following MCP tool.

Tool Name: %s
Tool Description: %s
Input Schema: %s
Sample Parameters: %s

Generate a realistic JSON response that this tool would return when called with the sample parameters.
The response should be realistic and match the tool's purpose.

Examples:
- For get_weather(city="Beijing"): return weather data with temperature, condition, humidity, etc.
- For get_patient_vitals(patient_id="12345", count=5): return 5 records of temperature/blood pressure readings
- For calculate_sum(a=10, b=20): return {"result": 30}

Return ONLY the JSON response data, no markdown formatting or explanations.`,
		toolName, toolDescription, string(schemaBytes), string(paramsBytes))

	reqBody := CompletionRequest{
		Model: config.Model,
		Messages: []Message{
			{Role: "system", Content: "You are a helpful assistant that generates realistic mock data for API tools. Return only valid JSON, no markdown."},
			{Role: "user", Content: prompt},
		},
		Temperature: config.Temperature,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := config.BaseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var completion CompletionResponse
	if err := json.Unmarshal(body, &completion); err != nil {
		return nil, err
	}

	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned")
	}

	content := completion.Choices[0].Message.Content

	// Try to extract JSON from markdown code blocks if present
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	var result interface{}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse generated JSON: %v, content: %s", err, content)
	}

	return result, nil
}
