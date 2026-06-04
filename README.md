# RAG Search

A local Retrieval-Augmented Generation application with a chat interface, Go backend, PostgreSQL/pgvector, and Ollama for embeddings + LLM inference.

## Architecture

```
User → Frontend (HTML/CSS/JS) → Go Backend → Ollama (BGE-M3 embeddings)
                                              → PostgreSQL/pgvector (similarity search)
                                              → Ollama (Qwen2.5:3b answer generation)
```

## Quick Start

```bash
# Start all services
docker compose up -d

# Seed the database with 12 articles (AI, physics, math)
docker compose exec backend ./seed

# Open the UI
open http://localhost:8080
```

First startup pulls Docker images and ML models (~2 GB), which takes a few minutes.

## Services

| Service | Port | Description |
|---|---|---|
| Frontend + API | `8080` | Chat UI + `/api/search` endpoint |
| PostgreSQL/pgvector | `5432` | Vector database with pgvector extension |
| Ollama | `11434` | BGE-M3 (embeddings) + Qwen2.5:3b (chat) |

## Project Structure

```
├── docker-compose.yml            # Ollama + pgvector + backend
├── backend/
│   ├── Dockerfile
│   ├── cmd/server/main.go        # HTTP server
│   ├── cmd/seed/main.go          # Database seeder (12 articles)
│   ├── internal/
│   │   ├── database/database.go  # pgx pool + pgvector queries
│   │   ├── embeddings/embeddings.go  # Ollama API client
│   │   ├── handlers/handlers.go  # Search endpoint (RAG pipeline)
│   │   └── models/models.go      # Data types
│   └── migrations/001_create_articles.sql  # DB schema
└── frontend/
    ├── index.html
    ├── style.css
    └── app.js
```

## API

```bash
curl -X POST http://localhost:8080/api/search \
  -H 'Content-Type: application/json' \
  -d '{"query":"What is quantum mechanics?"}'
```

## RAG Pipeline

1. Query is embedded with **BGE-M3** (1024-dim vector)
2. pgvector finds top-5 similar articles via cosine distance
3. Retrieved articles + query sent to **Qwen2.5:3b** for answer
4. Response returned with answer and source citations

## Re-seeding

```bash
docker compose exec db psql -U postgres -d ragexample -c "TRUNCATE articles;"
docker compose exec backend ./seed
```
