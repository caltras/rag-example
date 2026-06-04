package main

import (
	"context"
	"log"
	"os"
	"time"

	"rag-backend/internal/database"
	"rag-backend/internal/embeddings"
	"rag-backend/internal/models"
)

var seedArticles = []models.SeedArticle{
	{
		Title:  "The Rise of Large Language Models in Enterprise Applications",
		Date:   "2024-11-15",
		Author: "Dr. Sarah Chen",
		Text:   `Large language models have revolutionized how enterprises approach natural language processing tasks. From automated customer support to code generation, LLMs demonstrate remarkable versatility. Companies like OpenAI, Google, and Anthropic have pushed the boundaries of what's possible with transformer architectures. The key challenges remain in hallucination reduction, context window limitations, and cost-effective deployment. Fine-tuning approaches like LoRA have made it feasible for organizations to adapt base models to domain-specific tasks without requiring massive computational resources.`,
	},
	{
		Title:  "Understanding Vector Databases and Semantic Search",
		Date:   "2024-10-20",
		Author: "Marcus Johnson",
		Text:   `Vector databases have emerged as a critical infrastructure component for modern AI applications. Unlike traditional keyword-based search, vector search enables semantic understanding by representing data as high-dimensional embeddings. pgvector extends PostgreSQL with vector similarity search, making it possible to combine relational data with vector operations in a single database system. The choice between IVFFlat and HNSW indexes depends on the trade-off between query speed and build time. Cosine similarity remains the most popular distance metric for text embeddings, as it measures angular distance rather than magnitude.`,
	},
	{
		Title:  "Retrieval-Augmented Generation: Bridging Knowledge Gaps",
		Date:   "2024-09-05",
		Author: "Dr. Elena Rodriguez",
		Text:   `Retrieval-Augmented Generation (RAG) addresses one of the fundamental limitations of large language models: their reliance on static training data. By combining a retrieval system with a generative model, RAG architectures can access up-to-date information and domain-specific knowledge at inference time. The typical RAG pipeline involves chunking documents, generating embeddings, storing them in a vector database, and retrieving relevant context for each query. Advanced techniques include hybrid search (combining vector and keyword search), re-ranking, and query expansion. RAG has proven particularly effective in customer support, legal research, and medical diagnosis applications.`,
	},
	{
		Title:  "Go Performance Patterns for High-Throughput API Services",
		Date:   "2024-08-12",
		Author: "Alex Thompson",
		Text:   `Go's concurrency model, built on goroutines and channels, makes it an excellent choice for building high-throughput API services. Key performance patterns include connection pooling for database access, worker pools for CPU-bound tasks, and proper use of context for cancellation and timeouts. The standard library's net/http package provides robust HTTP handling, while third-party routers like chi and gorilla/mux offer additional flexibility. For JSON serialization, encoding/json is sufficient for most use cases, but applications requiring maximum throughput may benefit from alternative serializers. Profiling with pprof is essential for identifying bottlenecks in production systems.`,
	},
	{
		Title:  "Docker Compose for Local AI Development Environments",
		Date:   "2024-07-28",
		Author: "Priya Patel",
		Text:   `Setting up local AI development environments has become significantly easier with Docker Compose. Running models locally with Ollama eliminates API costs and latency while providing full control over model selection. The key services in a typical AI stack include a vector database (PostgreSQL with pgvector), an embedding model server, and an LLM inference server. Docker Compose health checks ensure services start in the correct order. Volume mounts persist model data across restarts, while environment variables manage configuration. For Apple Silicon Macs, Rosetta 2 emulation may be needed for some containers, though many images now provide native ARM64 support.`,
	},
	{
		Title:  "The BGE-M3 Model: Multilingual Embeddings for Global Applications",
		Date:   "2024-06-15",
		Author: "Dr. Wei Zhang",
		Text:   `BGE-M3, developed by the Beijing Academy of Artificial Intelligence (BAAI), represents a significant advancement in multilingual embedding models. Supporting over 100 languages, it produces 1024-dimensional dense vectors that capture semantic meaning across linguistic boundaries. BGE-M3 supports multiple retrieval modes including dense retrieval, sparse retrieval (lexical matching), and multi-vector retrieval. Its training methodology combines contrastive learning with knowledge distillation, resulting in embeddings that outperform previous models on multilingual benchmarks. The model is available under the MIT license and runs efficiently on consumer GPUs via Ollama.`,
	},
	{
		Title:  "DeepSeek-R1: Open-Source Reasoning Models for Complex Tasks",
		Date:   "2025-01-10",
		Author: "Dr. James Miller",
		Text:   `DeepSeek-R1 represents a breakthrough in open-source reasoning models, demonstrating capabilities comparable to proprietary systems in mathematical reasoning, code generation, and logical analysis. The model uses a Mixture-of-Experts architecture with 8B active parameters out of 16B total, making it efficient for deployment. DeepSeek-R1 employs reinforcement learning from human feedback (RLHF) to improve reasoning chains. The model supports long context windows up to 128K tokens, making it suitable for processing large documents. Available under a permissive license, DeepSeek-R1 can be run locally via Ollama, enabling private AI-powered applications without cloud dependencies.`,
	},
	{
		Title:  "Optimizing PostgreSQL with pgvector for Production Workloads",
		Date:   "2024-05-20",
		Author: "Jordan Lee",
		Text:   "Deploying pgvector in production requires careful consideration of index strategies, connection pooling, and query optimization. The IVFFlat index provides a good balance of build time and search speed for most workloads, while HNSW offers faster search at the cost of longer index construction. Key configuration parameters include `lists` for IVFFlat (typically sqrt(n) where n is the number of rows) and `m` and `ef_construction` for HNSW. Query performance can be improved by using approximate search with a high-quality filter, combining vector search with metadata filtering, and monitoring index hit rates. Regular `VACUUM` and `ANALYZE` operations maintain query performance as data grows.",
	},
	{
		Title:  "Quantum Mechanics: The Mathematics of the Subatomic World",
		Date:   "2025-02-10",
		Author: "Prof. Lisa Yamamoto",
		Text:   `Quantum mechanics describes the behavior of matter and energy at atomic and subatomic scales. The theory is built on a mathematical framework where the state of a system is represented by a wavefunction, a complex-valued function that contains all measurable information about the system. The Schrödinger equation governs how wavefunctions evolve over time, while the Heisenberg uncertainty principle sets fundamental limits on measurement precision. Key mathematical tools include Hilbert spaces, linear operators, and eigenvalue equations. Quantum superposition allows particles to exist in multiple states simultaneously until measured, and entanglement creates correlations between particles that persist across arbitrary distances. These principles underpin modern technologies including semiconductor electronics, lasers, and quantum computing.`,
	},
	{
		Title:  "General Relativity: Einstein's Geometric Theory of Gravity",
		Date:   "2025-03-05",
		Author: "Prof. Lisa Yamamoto",
		Text:   `General relativity, published by Albert Einstein in 1915, revolutionized our understanding of gravity by describing it as the curvature of spacetime caused by mass and energy. The theory is encapsulated in the Einstein field equations, a set of ten coupled nonlinear partial differential equations that relate the geometry of spacetime to the distribution of matter. Key predictions include the bending of light around massive objects (gravitational lensing), the existence of black holes, gravitational time dilation, and gravitational waves. The mathematics of general relativity relies heavily on differential geometry, including Riemannian manifolds, metric tensors, and the Ricci curvature tensor. Experimental confirmations include the precession of Mercury's orbit, the 1919 solar eclipse observations, and the 2015 LIGO detection of gravitational waves.`,
	},
	{
		Title:  "Number Theory: The Queen of Mathematics",
		Date:   "2025-01-20",
		Author: "Dr. Anika Sharma",
		Text:   `Number theory, often called the queen of mathematics, studies the properties and relationships of integers. Fundamental topics include prime numbers, divisibility, modular arithmetic, and Diophantine equations. The Riemann Hypothesis, one of the seven Millennium Prize Problems, concerns the distribution of prime numbers and their connection to the zeros of the Riemann zeta function. Fermat's Last Theorem, proved by Andrew Wiles in 1994, states that no three positive integers a, b, c satisfy a^n + b^n = c^n for any integer n greater than 2. Modern applications of number theory are pervasive in cryptography, particularly in RSA encryption which relies on the difficulty of factoring large composite numbers. Elliptic curve cryptography provides stronger security with smaller key sizes and is widely used in blockchain technology.`,
	},
	{
		Title:  "Linear Algebra and Its Applications in Machine Learning",
		Date:   "2025-04-01",
		Author: "Dr. Anika Sharma",
		Text:   `Linear algebra forms the mathematical foundation of modern machine learning and deep learning. Core concepts include vectors, matrices, linear transformations, eigenvalues, and singular value decomposition. In neural networks, data flows through layers as matrix multiplications followed by nonlinear activation functions. Principal Component Analysis (PCA) uses eigendecomposition of the covariance matrix to reduce dimensionality while preserving maximal variance. Word embeddings represent words as dense vectors where semantic relationships correspond to vector arithmetic operations. The attention mechanism in transformer models computes weighted combinations of value vectors using query-key similarity scores. Gradient descent, the primary optimization algorithm for training neural networks, relies on computing gradients using the chain rule, which is efficiently implemented through backpropagation. Understanding linear algebra is essential for developing and debugging machine learning models.`,
	},
}

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

	ollamaClient := embeddings.NewClient(ollamaURL)

	targetCount := len(seedArticles)
	count, err := db.ArticleCount(context.Background())
	if err != nil {
		log.Printf("Could not check article count: %v (continuing anyway)", err)
	}
	if count >= targetCount {
		log.Printf("Database already has %d articles (target: %d). Skipping seed.", count, targetCount)
		return
	}
	log.Printf("Database has %d articles, inserting %d more.", count, targetCount-count)

	for i, article := range seedArticles {
		date, err := time.Parse("2006-01-02", article.Date)
		if err != nil {
			log.Fatalf("Invalid date %q: %v", article.Date, err)
		}

		textForEmbedding := article.Title + ". " + article.Text
		embedding, err := ollamaClient.GenerateEmbedding(context.Background(), textForEmbedding)
		if err != nil {
			log.Fatalf("Failed to generate embedding for article %q: %v", article.Title, err)
		}

		modelArticle := &models.Article{
			Title:  article.Title,
			Date:   date,
			Text:   article.Text,
			Author: article.Author,
		}

		if err := db.InsertArticle(context.Background(), modelArticle, embedding); err != nil {
			log.Fatalf("Failed to insert article %q: %v", article.Title, err)
		}

		log.Printf("[%d/%d] Inserted: %s", i+1, len(seedArticles), article.Title)
	}

	log.Println("Seed complete!")
}
