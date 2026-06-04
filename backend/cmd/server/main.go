package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strings"

	"rag-backend/internal/database"
	"rag-backend/internal/handlers"
	"rag-backend/internal/llm"
)

func loadEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}

func main() {
	loadEnv(".env")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/ragexample?sslmode=disable"
	}

	db, err := database.New(dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Connected to database")

	provider := os.Getenv("LLM_PROVIDER")
	if provider == "" {
		provider = "ollama"
	}

	var embedder llm.Embedder
	var chat llm.ChatModel

	switch provider {
	case "openrouter":
		apiKey := os.Getenv("OPENROUTER_API_KEY")
		if apiKey == "" {
			log.Fatal("OPENROUTER_API_KEY is required for openrouter provider")
		}
		model := os.Getenv("OPENROUTER_MODEL")
		if model == "" {
			model = "openai/gpt-oss-120b:free"
		}
		chat = llm.NewOpenRouterProvider(apiKey, model)

		ollamaURL := os.Getenv("OLLAMA_URL")
		if ollamaURL == "" {
			ollamaURL = "http://localhost:11434"
		}
		embedder = llm.NewOllamaProvider(ollamaURL, "bge-m3", "")

	default:
		ollamaURL := os.Getenv("OLLAMA_URL")
		if ollamaURL == "" {
			ollamaURL = "http://localhost:11434"
		}
		embedModel := os.Getenv("OLLAMA_EMBED_MODEL")
		if embedModel == "" {
			embedModel = "bge-m3"
		}
		chatModel := os.Getenv("OLLAMA_CHAT_MODEL")
		if chatModel == "" {
			chatModel = "qwen2.5:3b"
		}
		ollama := llm.NewOllamaProvider(ollamaURL, embedModel, chatModel)
		embedder = ollama
		chat = ollama
	}

	h := handlers.New(db, embedder, chat)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/search", h.Search)
	mux.HandleFunc("/api/travel/search", h.TravelSearch)

	fs := http.FileServer(http.Dir("../frontend"))
	mux.Handle("/", fs)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
