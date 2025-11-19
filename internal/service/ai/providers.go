package ai

// ModelInfo represents information about an LLM model
type ModelInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Free        bool   `json:"free"`
}

// ProviderConfig represents a provider's configuration
type ProviderConfig struct {
	Name    string      `json:"name"`
	BaseURL string      `json:"base_url"`
	Models  []ModelInfo `json:"models"`
}

// ProviderPresets contains preset configurations for popular LLM providers
var ProviderPresets = map[string]ProviderConfig{
	"siliconflow": {
		Name:    "SiliconFlow (硅基流动)",
		BaseURL: "https://api.siliconflow.cn/v1",
		Models: []ModelInfo{
			{Name: "Qwen/Qwen2.5-7B-Instruct", DisplayName: "Qwen 2.5 7B (免费)", Free: true},
			{Name: "deepseek-ai/DeepSeek-V2.5", DisplayName: "DeepSeek V2.5 (免费)", Free: true},
			{Name: "THUDM/glm-4-9b-chat", DisplayName: "GLM-4 9B (免费)", Free: true},
			{Name: "Qwen/Qwen2.5-72B-Instruct", DisplayName: "Qwen 2.5 72B", Free: false},
			{Name: "Pro/Qwen/Qwen2.5-Coder-32B-Instruct", DisplayName: "Qwen 2.5 Coder 32B", Free: false},
		},
	},
	"modelscope": {
		Name:    "ModelScope (魔搭)",
		BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",
		Models: []ModelInfo{
			{Name: "qwen-turbo", DisplayName: "Qwen Turbo (便宜)", Free: false},
			{Name: "qwen-plus", DisplayName: "Qwen Plus (平衡)", Free: false},
			{Name: "qwen-max", DisplayName: "Qwen Max (最强)", Free: false},
		},
	},
	"openai": {
		Name:    "OpenAI",
		BaseURL: "https://api.openai.com/v1",
		Models: []ModelInfo{
			{Name: "gpt-4o-mini", DisplayName: "GPT-4o Mini (便宜)", Free: false},
			{Name: "gpt-4o", DisplayName: "GPT-4o (强大)", Free: false},
			{Name: "gpt-4-turbo", DisplayName: "GPT-4 Turbo", Free: false},
			{Name: "gpt-3.5-turbo", DisplayName: "GPT-3.5 Turbo (便宜)", Free: false},
		},
	},
	"deepseek": {
		Name:    "DeepSeek",
		BaseURL: "https://api.deepseek.com/v1",
		Models: []ModelInfo{
			{Name: "deepseek-chat", DisplayName: "DeepSeek Chat", Free: false},
			{Name: "deepseek-coder", DisplayName: "DeepSeek Coder", Free: false},
		},
	},
	"zhipu": {
		Name:    "智谱 AI",
		BaseURL: "https://open.bigmodel.cn/api/paas/v4",
		Models: []ModelInfo{
			{Name: "glm-4", DisplayName: "GLM-4", Free: false},
			{Name: "glm-4-flash", DisplayName: "GLM-4 Flash (便宜)", Free: false},
			{Name: "glm-3-turbo", DisplayName: "GLM-3 Turbo", Free: false},
		},
	},
	"anthropic": {
		Name:    "Anthropic",
		BaseURL: "https://api.anthropic.com/v1",
		Models: []ModelInfo{
			{Name: "claude-3-5-sonnet-20241022", DisplayName: "Claude 3.5 Sonnet", Free: false},
			{Name: "claude-3-opus-20240229", DisplayName: "Claude 3 Opus", Free: false},
			{Name: "claude-3-haiku-20240307", DisplayName: "Claude 3 Haiku (便宜)", Free: false},
		},
	},
	"google": {
		Name:    "Google",
		BaseURL: "https://generativelanguage.googleapis.com/v1beta",
		Models: []ModelInfo{
			{Name: "gemini-1.5-pro", DisplayName: "Gemini 1.5 Pro", Free: false},
			{Name: "gemini-1.5-flash", DisplayName: "Gemini 1.5 Flash (便宜)", Free: false},
		},
	},
	"kimi": {
		Name:    "Kimi (月之暗面)",
		BaseURL: "https://api.moonshot.cn/v1",
		Models: []ModelInfo{
			{Name: "moonshot-v1-8k", DisplayName: "Moonshot v1 8K", Free: false},
			{Name: "moonshot-v1-32k", DisplayName: "Moonshot v1 32K", Free: false},
			{Name: "moonshot-v1-128k", DisplayName: "Moonshot v1 128K", Free: false},
		},
	},
	"openrouter": {
		Name:    "OpenRouter",
		BaseURL: "https://openrouter.ai/api/v1",
		Models: []ModelInfo{
			{Name: "anthropic/claude-3.5-sonnet", DisplayName: "Claude 3.5 Sonnet", Free: false},
			{Name: "google/gemini-2.0-flash-exp:free", DisplayName: "Gemini 2.0 Flash (免费)", Free: true},
			{Name: "meta-llama/llama-3.2-3b-instruct:free", DisplayName: "Llama 3.2 3B (免费)", Free: true},
			{Name: "qwen/qwen-2.5-7b-instruct:free", DisplayName: "Qwen 2.5 7B (免费)", Free: true},
		},
	},
	"minimax": {
		Name:    "MiniMax",
		BaseURL: "https://api.minimax.chat/v1",
		Models: []ModelInfo{
			{Name: "abab6.5s-chat", DisplayName: "ABAB 6.5s Chat", Free: false},
			{Name: "abab6.5-chat", DisplayName: "ABAB 6.5 Chat", Free: false},
			{Name: "abab6.5g-chat", DisplayName: "ABAB 6.5g Chat", Free: false},
		},
	},
	"baichuan": {
		Name:    "百川智能",
		BaseURL: "https://api.baichuan-ai.com/v1",
		Models: []ModelInfo{
			{Name: "Baichuan4", DisplayName: "Baichuan 4", Free: false},
			{Name: "Baichuan3-Turbo", DisplayName: "Baichuan 3 Turbo", Free: false},
			{Name: "Baichuan3-Turbo-128k", DisplayName: "Baichuan 3 Turbo 128K", Free: false},
		},
	},
	"custom": {
		Name:    "Custom OpenAI-Compatible",
		BaseURL: "",
		Models:  []ModelInfo{}, // User provides custom models
	},
}

// GetProviderConfig returns the configuration for a given provider
func GetProviderConfig(provider string) (ProviderConfig, bool) {
	config, ok := ProviderPresets[provider]
	return config, ok
}

// ListProviders returns a list of all available providers
func ListProviders() []string {
	providers := make([]string, 0, len(ProviderPresets))
	for key := range ProviderPresets {
		providers = append(providers, key)
	}
	return providers
}

// FilterFreeModels returns only free models from a provider
func FilterFreeModels(provider string) []ModelInfo {
	config, ok := ProviderPresets[provider]
	if !ok {
		return []ModelInfo{}
	}

	freeModels := make([]ModelInfo, 0)
	for _, model := range config.Models {
		if model.Free {
			freeModels = append(freeModels, model)
		}
	}
	return freeModels
}
