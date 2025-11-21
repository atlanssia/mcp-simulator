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
	"go.uber.org/zap"
)

type Generator struct {
	config core.LLMConfig
	client *http.Client
	mu     sync.RWMutex
	logger *zap.Logger
}

func NewGenerator(config core.LLMConfig, logger *zap.Logger) *Generator {
	// If config is empty (default), try to load from env
	if config.APIKey == "" {
		config.APIKey = os.Getenv("OPENAI_API_KEY")
	}
	if config.Provider == "" {
		config = core.DefaultLLMConfig()
		if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
			config.APIKey = apiKey
		}
	}

	return &Generator{
		config: config,
		client: &http.Client{},
		logger: logger,
	}
}

// UpdateConfig updates the LLM configuration
func (g *Generator) UpdateConfig(config core.LLMConfig) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.config = config
}

// GetDefaultModel returns the configured model
func (g *Generator) GetDefaultModel() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.config.Model
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

// DefaultSystemPrompt for backward compatibility
const (
	DefaultModel        = "gpt-4o-mini"
	DefaultTemperature  = 0.7
	DefaultSystemPrompt = `You are a professional API schema generator. Your ONLY job is to output valid JSON schema objects.

Rules:
1. NEVER use markdown code blocks (no triple backticks)
2. NEVER add explanations or conversational text
3. ALWAYS start with { and end with }
4. Output ONLY the raw JSON object

When given a tool description, analyze it and output a JSON schema that matches this exact format:
{
  "type": "object",
  "properties": {
    "parameter_name": {
      "type": "string",
      "description": "what this parameter does"
    }
  },
  "required": ["parameter_name"]
}`
)

// GenerateToolSchema generates a JSON schema for a tool using default parameters
// Deprecated: Use GenerateToolSchemaWithParams for more control
func (g *Generator) GenerateToolSchema(ctx context.Context, description string) (map[string]interface{}, error) {
	return g.GenerateToolSchemaWithParams(ctx, description, core.DefaultGenerationParams())
}

func (g *Generator) GenerateToolSchemaWithParams(ctx context.Context, description string, params core.GenerationParams) (map[string]interface{}, error) {
	g.mu.RLock()
	config := g.config
	g.mu.RUnlock()

	// System prompt is ALWAYS the fixed role definition
	systemPrompt := DefaultSystemPrompt

	// Construct user message with:
	// 1. Custom user prompt (if provided)
	// 2. Tool description
	// 3. JSON structure template
	userMessage := ""
	if params.SystemPrompt != "" {
		// User's custom instruction goes first
		userMessage = params.SystemPrompt + "\n\n"
	}

	userMessage += fmt.Sprintf(`Task: Generate a JSON schema for this tool.

Tool Description: %s

Output ONLY the JSON schema object. Start your response with { and end with }. No other text.

Required JSON structure:
{
  "type": "object",
  "properties": {
    "param_name": {
      "type": "string|number|boolean|array|object",
      "description": "parameter description"
    }
  },
  "required": ["param_name"]
}`, description)

	// For ModelScope compatibility: enable_thinking must be false for non-streaming calls
	enableThinking := false
	reqBody := CompletionRequest{
		Model: params.Model,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		},
		Temperature:    params.Temperature,
		EnableThinking: &enableThinking,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	// Log the request body
	if g.logger != nil {
		g.logger.Info("LLM API Request",
			zap.String("url", config.BaseURL+"/chat/completions"),
			zap.String("model", params.Model),
			zap.ByteString("request_body", jsonBody),
		)
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

	// Log the response body
	if g.logger != nil {
		g.logger.Info("LLM API Response",
			zap.Int("status_code", resp.StatusCode),
			zap.ByteString("response_body", body),
		)
	}

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

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse generated JSON: %w, content: %s", err, content)
	}

	return result, nil
}

type CompletionRequest struct {
	Model          string    `json:"model"`
	Messages       []Message `json:"messages"`
	Temperature    float64   `json:"temperature,omitempty"`
	EnableThinking *bool     `json:"enable_thinking,omitempty"` // For ModelScope: must be false for non-streaming
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

// GenerateMockData generates mock data using default parameters
// Deprecated: Use GenerateMockDataWithParams for more control
func (g *Generator) GenerateMockData(ctx context.Context, prompt string, schema map[string]interface{}) (interface{}, error) {
	g.mu.RLock()
	config := g.config
	g.mu.RUnlock()

	// Use default parameters
	systemPrompt := strings.ReplaceAll(DefaultSystemPrompt, "{description}", prompt)

	reqBody := CompletionRequest{
		Model: DefaultModel,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
		Temperature: DefaultTemperature,
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

// GenerateMockResponse generates realistic mock response data for a tool using default parameters
// Deprecated: Use GenerateMockResponseWithParams for more control
func (g *Generator) GenerateMockResponse(ctx context.Context, toolName, toolDescription string, inputSchema, sampleParams map[string]interface{}) (interface{}, error) {
	return g.GenerateMockResponseWithParams(ctx, toolName, toolDescription, inputSchema, sampleParams, core.DefaultGenerationParams())
}

// GenerateMockResponseWithParams generates realistic mock response data with custom parameters
func (g *Generator) GenerateMockResponseWithParams(ctx context.Context, toolName, toolDescription string, inputSchema, sampleParams map[string]interface{}, params core.GenerationParams) (interface{}, error) {
	g.mu.RLock()
	config := g.config
	g.mu.RUnlock()

	// Build prompt for generating mock response data
	schemaBytes, _ := json.Marshal(inputSchema)
	paramsBytes, _ := json.Marshal(sampleParams)

	// System prompt for mock data generation
	systemPrompt := `You are a professional mock data generator. Your ONLY job is to output valid JSON data.

Rules:
1. NEVER use markdown code blocks (no triple backticks)
2. NEVER add explanations or conversational text
3. ALWAYS start with { or [ and end with } or ]
4. Output ONLY the raw JSON data
5. Generate REALISTIC and VARIED data
6. For time-series or tabular data, generate multiple records (5-10 items)
7. Use appropriate data types and realistic values`

	// User message with task and examples
	userMessage := ""
	if params.SystemPrompt != "" {
		// User's custom instructions (e.g., "生成5条患者体征数据，包含体温、心率、血压")
		userMessage = params.SystemPrompt + "\n\n"
	}

	userMessage += fmt.Sprintf(`Task: Generate realistic mock response data for this MCP tool.

Tool Name: %s
Tool Description: %s
Input Schema: %s
Sample Parameters: %s

Output Requirements:
1. Return ONLY valid JSON (no markdown, no explanations)
2. Data should be realistic and match the tool's purpose
3. For time-series/tabular data, generate 5-10 records
4. Use proper data types (strings in quotes, numbers without quotes)

Example formats:
- Weather data: {"temperature": 22.5, "condition": "Partly Cloudy", "humidity": 65, "wind_speed": 12}
- Patient vitals (multiple records): {"records": [{"time": "2023-10-15 08:30", "item": "体温", "value": "38.5", "unit": "℃"}, ...]}
- Calculation result: {"result": 30, "operation": "sum"}

Now generate the mock data:`, toolName, toolDescription, string(schemaBytes), string(paramsBytes))

	// For ModelScope compatibility
	enableThinking := false
	reqBody := CompletionRequest{
		Model: params.Model,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		},
		Temperature:    params.Temperature,
		EnableThinking: &enableThinking,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	// Log the request
	if g.logger != nil {
		g.logger.Info("LLM API Request (Mock Data)",
			zap.String("url", config.BaseURL+"/chat/completions"),
			zap.String("model", params.Model),
			zap.ByteString("request_body", jsonBody),
		)
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

	// Log the response
	if g.logger != nil {
		g.logger.Info("LLM API Response (Mock Data)",
			zap.Int("status_code", resp.StatusCode),
			zap.ByteString("response_body", body),
		)
	}

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
