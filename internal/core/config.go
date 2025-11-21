package core

// LLMConfig holds the configuration for AI/LLM provider
// LLMConfig contains the essential provider configuration
// Generation parameters (temperature, system_prompt, model) are now
// specified per-generation in the generation context
type LLMConfig struct {
	Provider string `json:"provider"` // Provider ID (e.g., "siliconflow", "openai")
	APIKey   string `json:"api_key"`  // API key (optional, depends on provider)
	BaseURL  string `json:"base_url"` // API endpoint (optional, uses provider default if empty)
	Model    string `json:"model"`
}

// DefaultLLMConfig returns a default configuration
func DefaultLLMConfig() LLMConfig {
	return LLMConfig{
		Provider: "openai",
		BaseURL:  "https://api.openai.com/v1",
	}
}
