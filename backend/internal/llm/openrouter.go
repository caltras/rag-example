package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type openRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterChoice struct {
	Message openRouterMessage `json:"message"`
}

type openRouterRequest struct {
	Model    string              `json:"model"`
	Messages []openRouterMessage `json:"messages"`
	Stream   bool                `json:"stream"`
}

type openRouterResponse struct {
	Choices []openRouterChoice `json:"choices"`
}

type OpenRouterProvider struct {
	apiKey string
	model  string
	http   *http.Client
}

func NewOpenRouterProvider(apiKey, model string) *OpenRouterProvider {
	return &OpenRouterProvider{
		apiKey: apiKey,
		model:  model,
		http:   &http.Client{},
	}
}

func (o *OpenRouterProvider) Chat(ctx context.Context, messages []Message) (string, error) {
	or := make([]openRouterMessage, len(messages))
	for i, m := range messages {
		or[i] = openRouterMessage{Role: m.Role, Content: m.Content}
	}
	body := openRouterRequest{
		Model:    o.model,
		Messages: or,
		Stream:   false,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal openrouter request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("create openrouter request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("openrouter request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openrouter returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result openRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode openrouter response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("openrouter returned no choices")
	}

	return result.Choices[0].Message.Content, nil
}
