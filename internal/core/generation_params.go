package core

// GenerationParams holds parameters for AI generation
type GenerationParams struct {
	Model        string
	Temperature  float64
	SystemPrompt string
	MaxTokens    int
}

// DefaultGenerationParams returns default generation parameters
func DefaultGenerationParams() GenerationParams {
	return GenerationParams{
		Model:        "",
		Temperature:  0.7,
		SystemPrompt: "",
		MaxTokens:    0,
	}
}
