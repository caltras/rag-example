package main

import (
	"log"
	"net/http"
	"os"

	"rag-backend/internal/database"
	"rag-backend/internal/embeddings"
	"rag-backend/internal/handlers"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/ragexample?sslmode=disable"
	}

	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}

	db, err := database.New(dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Connected to database")

	ollamaClient := embeddings.NewClient(ollamaURL)
	h := handlers.New(db, ollamaClient)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/search", h.Search)

	fs := http.FileServer(http.Dir("./frontend"))
	mux.Handle("/", fs)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
