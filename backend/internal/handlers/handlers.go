package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"rag-backend/internal/database"
	"rag-backend/internal/embeddings"
	"rag-backend/internal/models"
)

type Handler struct {
	db     *database.DB
	ollama *embeddings.Client
}

func New(db *database.DB, ollama *embeddings.Client) *Handler {
	return &Handler{db: db, ollama: ollama}
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Query) == "" {
		http.Error(w, "Query is required", http.StatusBadRequest)
		return
	}

	queryEmbedding, err := h.ollama.GenerateEmbedding(r.Context(), req.Query)
	if err != nil {
		log.Printf("embedding generation error: %v", err)
		http.Error(w, "Failed to generate embedding", http.StatusInternalServerError)
		return
	}
	articles, err := h.db.SearchSimilar(r.Context(), queryEmbedding, 5)
	if err != nil {
		log.Printf("search error: %v", err)
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	if len(articles) == 0 {
		resp := models.SearchResponse{
			Answer:  "No relevant articles found.",
			Sources: []string{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	var contextBuilder strings.Builder
	for _, a := range articles {
		contextBuilder.WriteString(fmt.Sprintf(
			"Title: %s\nAuthor: %s\nDate: %s\n\n%s\n\n---\n\n",
			a.Title, a.Author, a.Date.Format("2006-01-02"), a.Text,
		))
	}

	systemPrompt := `You are a helpful research assistant. Use the provided article excerpts to answer the user's question. 
If the context does not contain enough information to answer, say so clearly.
Cite the title and author of the sources you use.`

	userPrompt := fmt.Sprintf(
		"Here are relevant article excerpts:\n\n%s\n\nBased on these articles, answer: %s",
		contextBuilder.String(), req.Query,
	)

	answer, err := h.ollama.Chat(r.Context(), systemPrompt, userPrompt)
	if err != nil {
		log.Printf("chat generation error: %v", err)
		http.Error(w, "Failed to generate answer", http.StatusInternalServerError)
		return
	}

	sources := make([]string, len(articles))
	for i, a := range articles {
		sources[i] = a.Title
	}

	resp := models.SearchResponse{
		Answer:  answer,
		Sources: sources,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
