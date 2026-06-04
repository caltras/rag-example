# RAG Search

A local Retrieval-Augmented Generation application with a chat interface, Go backend, PostgreSQL/pgvector, and Ollama for embeddings + LLM inference.

## Architecture

```
User → Frontend (HTML/CSS/JS) → Go Backend → Ollama (BGE-M3 embeddings)
                                              → PostgreSQL/pgvector (similarity search)
                                              → Ollama / OpenRouter (answer generation)
```

## Quick Start

```bash
# Start PostgreSQL and Ollama containers
docker compose up -d db ollama

# Seed the database with 12 articles (AI, physics, math)
docker compose run --rm backend ./seed

# Build and run the backend natively (or use: docker compose up -d backend)
cd backend && go build -o /tmp/rag-backend ./cmd/server && cd ..
DATABASE_URL="postgres://postgres:postgres@localhost:5432/ragexample?sslmode=disable" \
OLLAMA_URL="http://localhost:11434" \
LLM_PROVIDER=openrouter \
OPENROUTER_API_KEY="sk-or-v1-..." \
OPENROUTER_MODEL="openai/gpt-oss-120b:free" \
/tmp/rag-backend &

# Open the UI
open http://localhost:8080
```

> **Note**: Native backend on macOS avoids corporate firewall TLS issues.
> Alternatively: `docker compose up -d` to run everything in Docker
> (may need Fortinet CA cert installed in containers).

## Services

| Service | Port | Description |
|---|---|---|
| Frontend + API | `8080` | Chat UI + `/api/search` endpoint |
| PostgreSQL/pgvector | `5432` | Vector database with pgvector extension |
| Ollama | `11434` | BGE-M3 (embeddings) + local chat models |

## Project Structure

```
├── docker-compose.yml            # Ollama + pgvector + backend
├── backend/
│   ├── Dockerfile
│   ├── cmd/server/main.go        # HTTP server
│   ├── cmd/seed/main.go          # Database seeder (12 articles)
│   ├── internal/
│   │   ├── database/database.go  # pgx pool + pgvector queries
│   │   ├── llm/                  # ChatModel + Embedder interfaces
│   │   │   ├── types.go          #   Message, Embedder, ChatModel interfaces
│   │   │   ├── ollama.go         #   Ollama provider (embed + chat)
│   │   │   └── openrouter.go     #   OpenRouter provider (chat)
│   │   ├── handlers/handlers.go  # Search endpoint (RAG pipeline + history)
│   │   └── models/models.go      # Data types
│   └── migrations/001_create_articles.sql  # DB schema
├── frontend/
│   ├── index.html
│   ├── style.css
│   └── app.js
├── RAG_Overview.md               # Full RAG concepts + travel agency guide
└── slides.html                   # Presentation slides
```

## API

```bash
curl -X POST http://localhost:8080/api/search \
  -H 'Content-Type: application/json' \
  -d '{"query":"What is quantum mechanics?", "session_id": "my-session"}'
```

## RAG Pipeline

1. Query is embedded with **BGE-M3** (1024-dim vector)
2. pgvector finds top-5 similar articles via cosine distance (`<=>` operator)
3. Retrieved articles + conversation history sent to LLM (OpenRouter or Ollama)
4. Response returned with answer and source citations

## Re-seeding

```bash
docker compose exec db psql -U postgres -d ragexample -c "TRUNCATE articles;"
docker compose run --rm backend ./seed
```

## Documentation

- [`RAG_Overview.md`](./RAG_Overview.md) — Comprehensive RAG guide: concepts, pgvector operators, tech stack, travel agency use case with examples
- [`slides.html`](./slides.html) — Presentation slides (navigate with arrow keys / click)
