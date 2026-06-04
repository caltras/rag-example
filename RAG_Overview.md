# Retrieval-Augmented Generation (RAG) — Concepts & Implementation

## Table of Contents
1. [What is RAG?](#1-what-is-rag)
2. [Why RAG?](#2-why-rag)
3. [Core Architecture](#3-core-architecture)
4. [pgvector Query Operators](#4-pgvector-query-operators)
5. [Technology Stack](#5-technology-stack)
6. [Travel Agency Use Case](#6-travel-agency-use-case)
7. [Running the Project](#7-running-the-project)
8. [References](#8-references)

---

## 1. What is RAG?

**Retrieval-Augmented Generation** is an AI architecture that combines a **retrieval system** (search) with a **generative model** (LLM). Instead of asking the LLM to answer from its static training data alone, RAG first retrieves relevant documents from a knowledge base, then feeds them as context to the LLM.

```
User Query
    │
    ▼
┌─────────────────────┐
│  Embedding Model    │  converts query into a vector
│  (e.g., BGE-M3)     │
└─────────┬───────────┘
          │ vector
          ▼
┌─────────────────────┐
│  Vector Database    │  searches for similar documents
│  (PostgreSQL +      │  using cosine / L2 / inner product
│   pgvector)         │
└─────────┬───────────┘
          │ relevant passages
          ▼
┌─────────────────────┐
│  LLM (Chat Model)   │  generates answer grounded in
│  (OpenRouter /      │  retrieved context
│   Ollama)           │
└─────────┬───────────┘
          │ answer
          ▼
      User sees response with citations
```

### Key Insight

The LLM **never saw the retrieved documents during training**. RAG lets you plug in **any data** — company wikis, product catalogs, customer histories — without fine-tuning or retraining the model.

---

## 2. Why RAG?

| Problem | RAG Solution |
|---------|-------------|
| LLM knowledge is static (cutoff date) | Retrieves live/up-to-date information |
| LLM hallucinates facts | Grounds answers in retrieved documents |
| Can't access private/ proprietary data | Queries your own vector database |
| Expensive to retrain/fine-tune | No training needed — just index new docs |
| Hard to cite sources | Retrieved passages act as verifiable citations |

### When to Use RAG vs. Fine-Tuning

| Criterion | RAG | Fine-Tuning |
|-----------|-----|-------------|
| New knowledge | ✅ Instant (index a doc) | ❌ Requires retraining |
| Source attribution | ✅ Cites retrieved docs | ❌ Black-box weights |
| Cost | Low (just inference) | High (training compute) |
| Latency | Slightly higher (retrieval step) | Same as normal inference |
| Tone / style control | Prompt engineering | ✅ Learns style from training data |

---

## 3. Core Architecture

### 3.1 Embedding Pipeline

Documents are **chunked** (split into pieces) and each chunk is converted into a **dense vector** (embedding) using an embedding model. In this project, we use **BGE-M3** (BAAI General Embedding — Multilingual, Multi-function).

- **Model**: BGE-M3 via Ollama
- **Vector dimensions**: 1024
- **Distance metric**: Cosine similarity (`<=>` operator)
- **Languages**: 100+ languages supported

### 3.2 Vector Search

The query is embedded with the same model, then the vector database finds documents whose embeddings are closest to the query embedding.

### 3.3 Answer Generation

The retrieved documents are injected into a **system prompt** that instructs the LLM to answer based only on the provided context. The LLM never sees the raw database — only the curated excerpts.

---

## 4. pgvector Query Operators

pgvector provides three distance operators for vector similarity search:

### 4.1 Cosine Distance — `<=>`

```
SELECT * FROM articles ORDER BY embedding <=> $1 LIMIT 5;
```

- **Range**: [0, 2] (0 = identical direction, 1 = orthogonal, 2 = opposite)
- **Use case**: Text embeddings, semantic search
- **Why we use it**: Cosine similarity measures **angular** distance, ignoring vector magnitude. This is ideal for text because a long document and a short one can have the same "meaning" even if their embedding magnitudes differ.

### 4.2 L2 Distance — `<->` (Euclidean)

```
SELECT * FROM articles ORDER BY embedding <-> $1 LIMIT 5;
```

- **Range**: [0, ∞)
- **Use case**: Image similarity, recommendation systems where magnitude matters
- **Note**: Sensitive to vector scale — usually requires normalized vectors

### 4.3 Inner Product — `<#>` (Negative dot product)

```
SELECT * FROM articles ORDER BY embedding <#> $1 LIMIT 5;
```

- **Range**: (-∞, ∞)
- **Use case**: When you want to maximise the dot product (commonly used in some embedding models trained with inner product loss)
- **Note**: Returns the **negative** inner product so that smaller = more similar (consistent with `ORDER BY` semantics)

### 4.4 Index Types

| Index | Build Speed | Query Speed | Memory | Best For |
|-------|------------|-------------|--------|----------|
| **IVFFlat** | Fast | Moderate | Low | Balanced workloads, quick setup |
| **HNSW** | Slow | Fast | Higher | Read-heavy, low-latency required |

**IVFFlat** divides vectors into `lists` (clusters). At query time, it searches only the nearest `probes` clusters. The `lists` parameter is typically set to `sqrt(n)` where `n` is the number of rows.

**HNSW** builds a hierarchical navigable small-world graph. Parameters:
- `m` — number of bi-directional links per element (default 16)
- `ef_construction` — size of dynamic candidate list during construction (default 64)

---

## 5. Technology Stack

| Layer | Technology | Role |
|-------|-----------|------|
| Database | **PostgreSQL 16 + pgvector** | Relational storage + vector index |
| Embeddings | **Ollama + BGE-M3** | Converts text to 1024‑dim vectors |
| Chat (prod) | **OpenRouter** (gpt-oss-120b:free) | Answer generation via API |
| Chat (local) | **Ollama + Qwen 2.5 3B** | Offline fallback |
| Backend | **Go 1.26** | API server, orchestrates RAG pipeline |
| Frontend | **Vanilla JS + CSS** | Minimal chat UI |
| Container | **Docker Compose** | Orchestrates PostgreSQL + Ollama |

### Why These Choices?

- **PostgreSQL + pgvector**: No need for a separate vector database. Your relational data and vectors live in the same database — transactional consistency, backup, and querying with a single tool.
- **BGE-M3**: State-of-the-art multilingual embeddings, MIT license, runs efficiently on CPU.
- **OpenRouter**: Single API key gives access to dozens of models. Free tier available.
- **Go**: Excellent concurrency, fast startup, single-binary deployment.
- **Ollama**: Run models locally — no data leaves your machine.

---

## 6. Travel Agency Use Case

### 6.1 Scenario

A travel agency wants to build an AI assistant that helps agents prepare personalised travel plans for clients. The knowledge base contains:

- Destination guides (attractions, restaurants, culture, weather)
- Hotel catalogues (amenities, prices, reviews, locations)
- Flight schedules and policies
- Client profiles (past trips, preferences, budget, passport data)
- Local regulations (visa requirements, travel advisories)

### 6.2 How RAG Helps

Instead of an agent manually searching through multiple systems, RAG lets them ask natural questions and get grounded, cited answers:

#### Example Queries

| Query | Retrieved Context | Answer Includes |
|-------|------------------|----------------|
| "Find a family-friendly hotel in Tokyo under $200/night for 5 nights" | Hotel catalogue, client profile (family of 4) | Top 3 hotels, total cost, family amenities |
| "What's the weather like in Barcelona in March, and what should we pack?" | Destination guide, weather data | Average temps, rain probability, packing list |
| "Plan a 7-day itinerary for Paris focused on art museums" | Destination guide, attraction data | Day-by-day schedule, ticket prices, nearby restaurants |
| "Does Mr. Smith need a visa for Japan?" | Client passport data, visa regulations | Yes/no, processing time, required documents |

### 6.3 Sample Indexed Data Structure

```sql
CREATE TABLE destinations (
    id SERIAL PRIMARY KEY,
    city TEXT NOT NULL,
    country TEXT NOT NULL,
    category TEXT NOT NULL,        -- 'attraction', 'hotel', 'restaurant', 'tip'
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    metadata JSONB,                -- price, rating, season, etc.
    embedding VECTOR(1024)
);

CREATE TABLE client_profiles (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    preferences JSONB,             -- dietary, activity level, budget range
    past_trips TEXT[],             -- array of destination IDs
    passport_data JSONB,
    embedding VECTOR(1024)
);
```

### 6.4 RAG-Enhanced Workflow

```
Agent: "Plan a 10-day Japan trip for the Smith family (2 adults, 2 kids under 10)"
1. Embed the query ──────────────────────────► BGE-M3
2. Search destinations ─────────────────────► pgvector: "Tokyo family hotels, kid-friendly attractions"
3. Search client profile ───────────────────► pgvector: "Smith family preferences, past trips"
4. Search logistics ────────────────────────► pgvector: "Japan visa requirements for US citizens"
5. Send to LLM ────────────────────────────► OpenRouter/Ollama
6. Receive structured plan ──────────────────► Day-by-day itinerary with costs and citations
```

#### Example Output

> # Japan Family Trip — 10-Day Plan for the Smith Family
>
> ## Flight Recommendations
> - **Outbound**: JFK → NRT (Narita), Delta DL-275, departing 7:00 AM
> - **Inbound**: NRT → JFK, Delta DL-276, departing 4:30 PM
> - **Total**: ~$3,200 economy (based on [Delta Family Fare, src: flights_2026_03])
>
> ## Hotel: Shinjuku Granbell Hotel
> - 2 connecting rooms, 10 nights
> - $180/night per room → $3,600 total
> - Kids under 12 stay free (src: hotels_tokyo_2026)
>
> ## Itinerary Highlights
> - **Day 1–3**: Tokyo — DisneySea, Ghibli Museum, teamLab Planets
> - **Day 4–5**: Hakone — Ryokan stay, Mt. Fuji view, open-air onsen
> - **Day 6–8**: Kyoto — Fushimi Inari, bamboo grove, samurai experience
> - **Day 9–10**: Osaka — Universal Studios, Dotonbori food tour
>
> ## Visa
> ✅ US passport holders get 90-day tourist visa on arrival. No advance application needed.
> (src: visa_japan_2026)
>
> **Total Estimated Cost**: $8,900 (flights + hotel + activities + meals)

### 6.5 Benefits for the Agency

| Before RAG | After RAG |
|------------|-----------|
| Agent searches 4+ systems manually | One natural-language query |
| 30–45 minutes per plan | 2–3 minutes per plan |
| Information scattered across emails, PDFs, databases | All knowledge indexed and searchable |
| Inconsistent plan quality | Standardised, cited responses |
| Rely on senior agents' tribal knowledge | Junior agents produce expert-level plans |

---

## 7. Running the Project

### Prerequisites

- Docker & Docker Compose
- Go 1.26+ (for native backend)
- OpenRouter API key (free at https://openrouter.ai)

### Quick Start

```bash
# 1. Start PostgreSQL and Ollama
docker compose up -d db ollama

# 2. Build and run the backend natively (macOS — avoids corporate TLS issues)
cd backend && go build -o /tmp/rag-backend ./cmd/server
cd ..
DATABASE_URL="postgres://postgres:postgres@localhost:5432/ragexample?sslmode=disable" \
OLLAMA_URL="http://localhost:11434" \
LLM_PROVIDER=openrouter \
OPENROUTER_API_KEY="sk-or-v1-..." \
OPENROUTER_MODEL="openai/gpt-oss-120b:free" \
/tmp/rag-backend &

# 3. Seed the database
docker compose run --rm backend ./seed

# 4. Open http://localhost:8080
```

### API

```bash
curl -X POST http://localhost:8080/api/search \
  -H 'Content-Type: application/json' \
  -d '{"query": "what are vector databases?", "session_id": "my-session"}'
```

---

## 8. References

1. **Lewis, P., et al. (2020).** *Retrieval-Augmented Generation for Knowledge-Intensive NLP Tasks*. NeurIPS.  
   https://arxiv.org/abs/2005.11401

2. **BAAI (2024).** *BGE-M3: Multilingual Multi-Function Embeddings*.  
   https://huggingface.co/BAAI/bge-m3

3. **pgvector Documentation.** *Vector similarity search for PostgreSQL*.  
   https://github.com/pgvector/pgvector

4. **Ollama.** *Get up and running with large language models locally*.  
   https://ollama.ai

5. **OpenRouter.** *Unified API for LLMs*.  
   https://openrouter.ai

6. **Go Documentation.** *Build fast, reliable, and efficient software*.  
   https://go.dev/doc/

7. **Docker Compose Documentation.** *Define and run multi-container applications*.  
   https://docs.docker.com/compose/

8. **Johnson, J., Douze, M., & Jégou, H. (2019).** *Billion-scale similarity search with GPUs*. (HNSW algorithm reference).  
   https://arxiv.org/abs/1702.08734

9. **Malkov, Y. A., & Yashunin, D. A. (2016).** *Efficient and robust approximate nearest neighbor search using Hierarchical Navigable Small World graphs*.  
   https://arxiv.org/abs/1603.09320

10. **PostgreSQL Documentation.** *Chapter 11: Indexes*.  
    https://www.postgresql.org/docs/current/indexes.html
