package core

// LLMConfig holds the configuration for AI/LLM provider
type LLMConfig struct {
	Provider     string  `json:"provider"`      // "siliconflow", "modelscope", "openai", etc.
	APIKey       string  `json:"api_key"`       // API key (will be encrypted in storage)
	BaseURL      string  `json:"base_url"`      // Custom endpoint URL
	Model        string  `json:"model"`         // Selected model name
	Temperature  float64 `json:"temperature"`   // 0.0 - 2.0
	SystemPrompt string  `json:"system_prompt"` // Custom prompt template
}

// DefaultLLMConfig returns a default configuration
func DefaultLLMConfig() LLMConfig {
	return LLMConfig{
		Provider:    "openai",
		BaseURL:     "https://api.openai.com/v1",
		Model:       "gpt-4o-mini",
		Temperature: 0.7,
		SystemPrompt: `You are an expert API designer. Generate a JSON schema for the following tool description.

Tool Description: {description}

Requirements:
1. Return ONLY valid JSON schema (no markdown, no explanations)
2. Use "type": "object" as the root
3. Define "properties" with appropriate types (string, number, boolean, array, object)
4. Include "description" for each property
5. Specify "required" array for mandatory fields
6. Use clear, descriptive property names

Example format:
{
  "type": "object",
  "properties": {
    "param1": {
      "type": "string",
      "description": "Description of param1"
    }
  },
  "required": ["param1"]
}`,
	}
}
