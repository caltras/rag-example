# RAG Search

A local Retrieval-Augmented Generation application with a chat interface, Go backend, PostgreSQL/pgvector, and Ollama for embeddings + OpenRouter for LLM inference.

Includes a **Travel Planning** module that searches flights, hotels, and destinations for multi-city trip planning.

## Architecture

```
User → Frontend (HTML/CSS/JS) → Go Backend → Ollama (BGE-M3 embeddings)
                                              → PostgreSQL/pgvector (similarity search)
                                              → OpenRouter (answer generation)
```

## Quick Start

```bash
# Start PostgreSQL and Ollama containers
docker compose up -d db ollama

# Seed database (articles + travel data)
docker compose run --rm backend ./seed

# Build and run the backend natively
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

## Services

| Service | Port | Description |
|---|---|---|
| Frontend + API | `8080` | Chat UI + `/api/search` and `/api/travel/search` endpoints |
| PostgreSQL/pgvector | `5432` | Vector database with pgvector extension |
| Ollama | `11434` | BGE-M3 embeddings (1024-dim) |

## Features

### Articles (Knowledge Base RAG)
- Ask questions about seeded articles (LLMs, vector databases, RAG, physics, math)
- Conversational history kept per session (last 6 messages)
- Sources panel shows which articles were retrieved

### Travel Planning
- Multi-table vector search across **flights**, **hotels**, and **destinations**
- Returns structured travel plans with flight options, hotel costs, and daily itineraries
- Seeded with data for Calgary, Toronto, Vancouver, Montreal, New York, Edmonton

### Chat / Data Mode Toggle
- **Chat** (default) — LLM generates a natural language answer / travel plan
- **Data** — returns raw retrieved data as structured markdown, skipping the LLM entirely — responses show a purple "Raw Data" badge

### Performance Timing
Each response includes millisecond-level timing breakdown:
- `timing_embed_ms` — BGE-M3 embedding generation
- `timing_search_ms` — pgvector similarity search
- `timing_chat_ms` — LLM chat completion (0 in Data mode)
- `timing_total_ms` — end-to-end request time

## Project Structure

```
├── docker-compose.yml            # Ollama + pgvector
├── backend/
│   ├── Dockerfile
│   ├── .env                      # API keys and provider config
│   ├── cmd/server/main.go        # HTTP server, provider wiring
│   ├── cmd/seed/main.go          # Database seeder (articles + travel)
│   ├── internal/
│   │   ├── database/database.go  # pgx pool + pgvector queries
│   │   ├── llm/                  # ChatModel + Embedder interfaces
│   │   │   ├── types.go          #   Message, Embedder, ChatModel
│   │   │   ├── ollama.go         #   Ollama provider (embed + chat)
│   │   │   └── openrouter.go     #   OpenRouter provider (chat)
│   │   ├── handlers/
│   │   │   ├── handlers.go       # /api/search (articles RAG)
│   │   │   └── travel.go         # /api/travel/search (travel RAG)
│   │   └── models/models.go      # Data types
│   └── migrations/
│       ├── 001_create_articles.sql
│       └── 002_create_travel.sql
├── frontend/
│   ├── index.html
│   ├── style.css
│   └── app.js
├── RAG_Overview.md               # Full RAG concepts + travel agency guide
└── slides.html                   # Presentation slides
```

## API

### Search Articles
```bash
curl -X POST http://localhost:8080/api/search \
  -H 'Content-Type: application/json' \
  -d '{"query":"What is quantum mechanics?", "session_id": "my-session", "mode": "chat"}'
```

### Search Travel
```bash
curl -X POST http://localhost:8080/api/travel/search \
  -H 'Content-Type: application/json' \
  -d '{"query":"Plan a trip to Calgary from Toronto for 5 days", "session_id": "my-trip", "mode": "chat"}'
```

Set `"mode": "data"` to skip LLM and return raw results.

## RAG Pipeline

1. Query is embedded with **BGE-M3** (1024-dim vector)
2. pgvector finds top-5 similar results via cosine distance (`<=>` operator)
3. Retrieved data + conversation history sent to LLM (OpenRouter)
4. Response returned with answer, source citations, and timing breakdown

## Re-seeding

```bash
# Articles only
docker compose exec db psql -U postgres -d ragexample -c "TRUNCATE articles;"
docker compose run --rm backend ./seed

# Full reset (articles + travel)
docker compose exec db psql -U postgres -d ragexample \
  -c "TRUNCATE articles, flights, hotels, destinations;"
docker compose run --rm backend ./seed
```

## Documentation

- [`RAG_Overview.md`](./RAG_Overview.md) — Comprehensive RAG guide: concepts, pgvector operators, tech stack, travel agency use case with examples
- [`slides.html`](./slides.html) — Presentation slides (navigate with arrow keys / click)
