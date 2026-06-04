package llm

import "context"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Embedder interface {
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
}

type ChatModel interface {
	Chat(ctx context.Context, messages []Message) (string, error)
}
