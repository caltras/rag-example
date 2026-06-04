package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"rag-backend/internal/database"
	"rag-backend/internal/llm"
	"rag-backend/internal/models"
)

type conversation struct {
	messages []llm.Message
}

type Handler struct {
	mu       sync.Mutex
	convos   map[string]*conversation
	maxHist  int
	db       *database.DB
	embedder llm.Embedder
	chat     llm.ChatModel
}

func New(db *database.DB, embedder llm.Embedder, chat llm.ChatModel) *Handler {
	return &Handler{
		convos:   make(map[string]*conversation),
		maxHist:  6,
		db:       db,
		embedder: embedder,
		chat:     chat,
	}
}

func (h *Handler) getOrCreateConvo(sessionID string) *conversation {
	if sessionID == "" {
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	c, ok := h.convos[sessionID]
	if !ok {
		c = &conversation{}
		h.convos[sessionID] = c
	}
	return c
}

func (h *Handler) appendHistory(c *conversation, role, content string) {
	c.messages = append(c.messages, llm.Message{Role: role, Content: content})
	if len(c.messages) > h.maxHist {
		c.messages = c.messages[len(c.messages)-h.maxHist:]
	}
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	startTotal := time.Now()

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

	convo := h.getOrCreateConvo(req.SessionID)
	sessionID := req.SessionID
	if convo != nil && sessionID == "" {
		sessionID = fmt.Sprintf("%p", convo)
	}

	embedStart := time.Now()
	queryEmbedding, err := h.embedder.GenerateEmbedding(r.Context(), req.Query)
	if err != nil {
		log.Printf("embedding generation error: %v", err)
		http.Error(w, "Failed to generate embedding", http.StatusInternalServerError)
		return
	}
	embedTime := time.Since(embedStart)
	log.Printf("embedding: completed in %v", embedTime)

	searchStart := time.Now()
	articles, err := h.db.SearchSimilar(r.Context(), queryEmbedding, 5)
	if err != nil {
		log.Printf("search error: %v", err)
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}
	searchTime := time.Since(searchStart)
	log.Printf("search_similarity: found %d articles in %v", len(articles), searchTime)

	if len(articles) == 0 {
		totalTime := time.Since(startTotal)
		resp := models.SearchResponse{
			Answer:         "No relevant articles found.",
			Sources:        []string{},
			TimingEmbedMs:  embedTime.Milliseconds(),
			TimingSearchMs: searchTime.Milliseconds(),
			TimingTotalMs:  totalTime.Milliseconds(),
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

	var dataOnly string
	var answer string
	var chatStart time.Time
	if req.Mode == "data" {
		for _, a := range articles {
			dataOnly += fmt.Sprintf("# %s\n**Author:** %s  \n**Date:** %s\n\n%s\n\n---\n\n",
				a.Title, a.Author, a.Date.Format("2006-01-02"), a.Text)
		}
	}

	if req.Mode != "data" {
		systemPrompt := `You are a helpful research assistant. Use the provided article excerpts to answer the user's question. 
If the context does not contain enough information to answer, say so clearly.
Cite the title and author of the sources you use.`

		userPrompt := fmt.Sprintf(
			"Here are relevant article excerpts:\n\n%s\n\nBased on these articles, answer: %s",
			contextBuilder.String(), req.Query,
		)

		messages := []llm.Message{{Role: "system", Content: systemPrompt}}

		if convo != nil {
			start := 0
			if len(convo.messages) > h.maxHist {
				start = len(convo.messages) - h.maxHist
			}
			for _, m := range convo.messages[start:] {
				messages = append(messages, m)
			}
		}

		messages = append(messages, llm.Message{Role: "user", Content: userPrompt})

		chatStart = time.Now()
		answer, err = h.chat.Chat(r.Context(), messages)
		if err != nil {
			log.Printf("chat generation error: %v", err)
			http.Error(w, "Failed to generate answer", http.StatusInternalServerError)
			return
		}
		log.Printf("chat_completion: completed in %v", time.Since(chatStart))

		if convo != nil {
			h.appendHistory(convo, "user", req.Query)
			h.appendHistory(convo, "assistant", answer)
		}
	} else {
		answer = dataOnly
	}

	sources := make([]string, len(articles))
	for i, a := range articles {
		sources[i] = a.Title
	}

	chatTime := time.Since(chatStart)
	totalTime := time.Since(startTotal)

	resp := models.SearchResponse{
		Answer:    answer,
		Sources:   sources,
		SessionID: sessionID,
		Mode:      req.Mode,
		TimingEmbedMs:  embedTime.Milliseconds(),
		TimingSearchMs: searchTime.Milliseconds(),
		TimingChatMs:   chatTime.Milliseconds(),
		TimingTotalMs:  totalTime.Milliseconds(),
	}
	if req.Mode == "data" {
		resp.TimingChatMs = 0
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
