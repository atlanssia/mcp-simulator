package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Generator struct {
	apiKey string
	client *http.Client
}

func NewGenerator(apiKey string) *Generator {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	return &Generator{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

type CompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
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
	schemaBytes, _ := json.Marshal(schema)
	systemPrompt := fmt.Sprintf("You are a helpful assistant that generates JSON data based on a description and a JSON schema. Only return the JSON data, no markdown formatting. Schema: %s", string(schemaBytes))

	reqBody := CompletionRequest{
		Model: "gpt-4o",
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.apiKey)

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var completion CompletionResponse
	if err := json.Unmarshal(body, &completion); err != nil {
		return nil, err
	}

	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned")
	}

	content := completion.Choices[0].Message.Content
	var result interface{}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse generated JSON: %v", err)
	}

	return result, nil
}
