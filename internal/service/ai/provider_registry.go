package ai

// ModelInfo represents information about an LLM model
type ModelInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Free        bool   `json:"free"`
}

// ProviderInfo contains metadata about an LLM provider
type ProviderInfo struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	RequiresAPIKey bool   `json:"requires_api_key"`
	DefaultBaseURL string `json:"default_base_url"`
	SupportsModels bool   `json:"supports_models"` // Can fetch model list dynamically
	Description    string `json:"description"`
}

// ProviderRegistry contains metadata for all supported providers
var ProviderRegistry = map[string]ProviderInfo{
	"siliconflow": {
		ID:             "siliconflow",
		Name:           "SiliconFlow (硅基流动)",
		RequiresAPIKey: true,
		DefaultBaseURL: "https://api.siliconflow.cn/v1",
		SupportsModels: true,
		Description:    "提供多个免费模型，需要 API Key",
	},
	"modelscope": {
		ID:             "modelscope",
		Name:           "ModelScope (魔搭)",
		RequiresAPIKey: false,
		DefaultBaseURL: "https://api-inference.modelscope.cn/v1",
		SupportsModels: true,
		Description:    "无需 API Key，60+ 模型可用",
	},
	"openai": {
		ID:             "openai",
		Name:           "OpenAI",
		RequiresAPIKey: true,
		DefaultBaseURL: "https://api.openai.com/v1",
		SupportsModels: false,
		Description:    "GPT 系列模型",
	},
	"anthropic": {
		ID:             "anthropic",
		Name:           "Anthropic (Claude)",
		RequiresAPIKey: true,
		DefaultBaseURL: "https://api.anthropic.com/v1",
		SupportsModels: false,
		Description:    "Claude 系列模型",
	},
	"google": {
		ID:             "google",
		Name:           "Google (Gemini)",
		RequiresAPIKey: true,
		DefaultBaseURL: "https://generativelanguage.googleapis.com/v1",
		SupportsModels: false,
		Description:    "Gemini 系列模型",
	},
	"deepseek": {
		ID:             "deepseek",
		Name:           "DeepSeek",
		RequiresAPIKey: true,
		DefaultBaseURL: "https://api.deepseek.com/v1",
		SupportsModels: false,
		Description:    "DeepSeek 系列模型",
	},
	"zhipu": {
		ID:             "zhipu",
		Name:           "Zhipu (智谱 AI)",
		RequiresAPIKey: true,
		DefaultBaseURL: "https://open.bigmodel.cn/api/paas/v4",
		SupportsModels: false,
		Description:    "GLM 系列模型",
	},
	"kimi": {
		ID:             "kimi",
		Name:           "Kimi (月之暗面)",
		RequiresAPIKey: true,
		DefaultBaseURL: "https://api.moonshot.cn/v1",
		SupportsModels: false,
		Description:    "Moonshot 系列模型，长上下文",
	},
	"openrouter": {
		ID:             "openrouter",
		Name:           "OpenRouter",
		RequiresAPIKey: true,
		DefaultBaseURL: "https://openrouter.ai/api/v1",
		SupportsModels: false,
		Description:    "聚合多个模型，部分免费",
	},
	"minimax": {
		ID:             "minimax",
		Name:           "MiniMax",
		RequiresAPIKey: true,
		DefaultBaseURL: "https://api.minimax.chat/v1",
		SupportsModels: false,
		Description:    "ABAB 系列模型",
	},
	"baichuan": {
		ID:             "baichuan",
		Name:           "Baichuan (百川智能)",
		RequiresAPIKey: true,
		DefaultBaseURL: "https://api.baichuan-ai.com/v1",
		SupportsModels: false,
		Description:    "Baichuan 系列模型",
	},
	"custom": {
		ID:             "custom",
		Name:           "Custom (自定义)",
		RequiresAPIKey: true,
		DefaultBaseURL: "",
		SupportsModels: false,
		Description:    "自定义 OpenAI 兼容接口",
	},
}

// GetProviderInfo returns info for a specific provider
func GetProviderInfo(id string) (ProviderInfo, bool) {
	info, ok := ProviderRegistry[id]
	return info, ok
}

// ListProviders returns all provider information
func ListProviders() []ProviderInfo {
	providers := make([]ProviderInfo, 0, len(ProviderRegistry))
	for _, info := range ProviderRegistry {
		providers = append(providers, info)
	}
	return providers
}
