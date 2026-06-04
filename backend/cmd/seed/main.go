package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"rag-backend/internal/database"
	"rag-backend/internal/llm"
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

	embedder := llm.NewOllamaProvider(ollamaURL, "bge-m3", "")

	targetCount := len(seedArticles)
	count, err := db.ArticleCount(context.Background())
	if err != nil {
		log.Printf("Could not check article count: %v (continuing anyway)", err)
	}
	if count >= targetCount {
		log.Printf("Database already has %d articles (target: %d). Skipping articles seed.", count, targetCount)
	} else {
		log.Printf("Database has %d articles, inserting %d more.", count, targetCount-count)

		for i, article := range seedArticles {
		date, err := time.Parse("2006-01-02", article.Date)
		if err != nil {
			log.Fatalf("Invalid date %q: %v", article.Date, err)
		}

		textForEmbedding := article.Title + ". " + article.Text
		embedding, err := embedder.GenerateEmbedding(context.Background(), textForEmbedding)
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
		
		log.Println("Articles seed complete!")
	}

	seedTravel(db, embedder)
	log.Println("Seed complete!")
}

func seedTravel(db *database.DB, embedder *llm.OllamaProvider) {
	ctx := context.Background()

	count, err := db.HotelCount(ctx)
	if err == nil && count > 0 {
		log.Printf("Travel data already exists (%d hotels). Skipping travel seed.", count)
		return
	}

	seedFlights := []models.SeedFlight{
		{Airline: "Air Canada", FlightNumber: "AC-123", Origin: "Toronto (YYZ)", Destination: "Calgary (YYC)", DepartureTime: "07:00", ArrivalTime: "09:30", Price: 350.00, Currency: "CAD", Date: "2026-07-01", Class: "economy"},
		{Airline: "Air Canada", FlightNumber: "AC-124", Origin: "Toronto (YYZ)", Destination: "Calgary (YYC)", DepartureTime: "07:00", ArrivalTime: "09:30", Price: 550.00, Currency: "CAD", Date: "2026-07-01", Class: "business"},
		{Airline: "WestJet", FlightNumber: "WS-456", Origin: "Toronto (YYZ)", Destination: "Calgary (YYC)", DepartureTime: "10:00", ArrivalTime: "12:30", Price: 280.00, Currency: "CAD", Date: "2026-07-01", Class: "economy"},
		{Airline: "WestJet", FlightNumber: "WS-457", Origin: "Toronto (YYZ)", Destination: "Calgary (YYC)", DepartureTime: "10:00", ArrivalTime: "12:30", Price: 450.00, Currency: "CAD", Date: "2026-07-01", Class: "business"},
		{Airline: "Air Canada", FlightNumber: "AC-127", Origin: "Toronto (YYZ)", Destination: "Calgary (YYC)", DepartureTime: "14:00", ArrivalTime: "16:30", Price: 320.00, Currency: "CAD", Date: "2026-07-01", Class: "economy"},
		{Airline: "WestJet", FlightNumber: "WS-458", Origin: "Toronto (YYZ)", Destination: "Calgary (YYC)", DepartureTime: "18:00", ArrivalTime: "20:30", Price: 260.00, Currency: "CAD", Date: "2026-07-01", Class: "economy"},
		{Airline: "Flair Airlines", FlightNumber: "F8-301", Origin: "Toronto (YYZ)", Destination: "Calgary (YYC)", DepartureTime: "06:30", ArrivalTime: "09:00", Price: 179.00, Currency: "CAD", Date: "2026-07-01", Class: "economy"},
		{Airline: "Porter Airlines", FlightNumber: "PD-211", Origin: "Toronto (YTZ)", Destination: "Calgary (YYC)", DepartureTime: "08:00", ArrivalTime: "10:30", Price: 199.00, Currency: "CAD", Date: "2026-07-01", Class: "economy"},
		{Airline: "Air Canada", FlightNumber: "AC-130", Origin: "Calgary (YYC)", Destination: "Toronto (YYZ)", DepartureTime: "08:00", ArrivalTime: "12:30", Price: 330.00, Currency: "CAD", Date: "2026-07-07", Class: "economy"},
		{Airline: "WestJet", FlightNumber: "WS-460", Origin: "Calgary (YYC)", Destination: "Toronto (YYZ)", DepartureTime: "15:00", ArrivalTime: "19:30", Price: 250.00, Currency: "CAD", Date: "2026-07-07", Class: "economy"},
		{Airline: "Flair Airlines", FlightNumber: "F8-302", Origin: "Calgary (YYC)", Destination: "Toronto (YYZ)", DepartureTime: "11:00", ArrivalTime: "15:30", Price: 159.00, Currency: "CAD", Date: "2026-07-07", Class: "economy"},
		{Airline: "Air Canada", FlightNumber: "AC-100", Origin: "Toronto (YYZ)", Destination: "Vancouver (YVR)", DepartureTime: "07:30", ArrivalTime: "09:00", Price: 389.00, Currency: "CAD", Date: "2026-07-01", Class: "economy"},
		{Airline: "Air Canada", FlightNumber: "AC-101", Origin: "Toronto (YYZ)", Destination: "Vancouver (YVR)", DepartureTime: "07:30", ArrivalTime: "09:00", Price: 689.00, Currency: "CAD", Date: "2026-07-01", Class: "business"},
		{Airline: "WestJet", FlightNumber: "WS-101", Origin: "Toronto (YYZ)", Destination: "Vancouver (YVR)", DepartureTime: "09:00", ArrivalTime: "10:30", Price: 299.00, Currency: "CAD", Date: "2026-07-01", Class: "economy"},
		{Airline: "WestJet", FlightNumber: "WS-102", Origin: "Toronto (YYZ)", Destination: "Vancouver (YVR)", DepartureTime: "21:00", ArrivalTime: "22:30", Price: 219.00, Currency: "CAD", Date: "2026-07-01", Class: "economy"},
		{Airline: "Flair Airlines", FlightNumber: "F8-100", Origin: "Toronto (YYZ)", Destination: "Vancouver (YVR)", DepartureTime: "06:00", ArrivalTime: "07:30", Price: 159.00, Currency: "CAD", Date: "2026-07-01", Class: "economy"},
		{Airline: "Air Canada", FlightNumber: "AC-200", Origin: "Toronto (YYZ)", Destination: "Montreal (YUL)", DepartureTime: "08:00", ArrivalTime: "09:15", Price: 189.00, Currency: "CAD", Date: "2026-07-01", Class: "economy"},
		{Airline: "Air Canada", FlightNumber: "AC-201", Origin: "Toronto (YYZ)", Destination: "Montreal (YUL)", DepartureTime: "08:00", ArrivalTime: "09:15", Price: 389.00, Currency: "CAD", Date: "2026-07-01", Class: "business"},
		{Airline: "Porter Airlines", FlightNumber: "PD-100", Origin: "Toronto (YTZ)", Destination: "Montreal (YUL)", DepartureTime: "09:30", ArrivalTime: "10:45", Price: 139.00, Currency: "CAD", Date: "2026-07-01", Class: "economy"},
		{Airline: "WestJet", FlightNumber: "WS-200", Origin: "Toronto (YYZ)", Destination: "Montreal (YUL)", DepartureTime: "16:00", ArrivalTime: "17:15", Price: 159.00, Currency: "CAD", Date: "2026-07-01", Class: "economy"},
		{Airline: "Air Canada", FlightNumber: "AC-300", Origin: "Toronto (YYZ)", Destination: "New York (LGA)", DepartureTime: "07:00", ArrivalTime: "09:00", Price: 249.00, Currency: "USD", Date: "2026-07-01", Class: "economy"},
		{Airline: "Air Canada", FlightNumber: "AC-301", Origin: "Toronto (YYZ)", Destination: "New York (LGA)", DepartureTime: "07:00", ArrivalTime: "09:00", Price: 549.00, Currency: "USD", Date: "2026-07-01", Class: "business"},
		{Airline: "WestJet", FlightNumber: "WS-300", Origin: "Toronto (YYZ)", Destination: "New York (JFK)", DepartureTime: "11:00", ArrivalTime: "13:00", Price: 199.00, Currency: "USD", Date: "2026-07-01", Class: "economy"},
		{Airline: "Delta Air Lines", FlightNumber: "DL-500", Origin: "Toronto (YYZ)", Destination: "New York (JFK)", DepartureTime: "15:30", ArrivalTime: "17:30", Price: 279.00, Currency: "USD", Date: "2026-07-01", Class: "economy"},
		{Airline: "Air Canada", FlightNumber: "AC-150", Origin: "Calgary (YYC)", Destination: "Vancouver (YVR)", DepartureTime: "09:00", ArrivalTime: "09:30", Price: 189.00, Currency: "CAD", Date: "2026-07-02", Class: "economy"},
		{Airline: "WestJet", FlightNumber: "WS-150", Origin: "Calgary (YYC)", Destination: "Vancouver (YVR)", DepartureTime: "14:00", ArrivalTime: "14:30", Price: 149.00, Currency: "CAD", Date: "2026-07-02", Class: "economy"},
		{Airline: "Flair Airlines", FlightNumber: "F8-150", Origin: "Calgary (YYC)", Destination: "Vancouver (YVR)", DepartureTime: "07:00", ArrivalTime: "07:30", Price: 99.00, Currency: "CAD", Date: "2026-07-02", Class: "economy"},
		{Airline: "Air Canada", FlightNumber: "AC-160", Origin: "Calgary (YYC)", Destination: "Edmonton (YEG)", DepartureTime: "08:00", ArrivalTime: "09:00", Price: 129.00, Currency: "CAD", Date: "2026-07-02", Class: "economy"},
		{Airline: "WestJet", FlightNumber: "WS-160", Origin: "Calgary (YYC)", Destination: "Edmonton (YEG)", DepartureTime: "17:00", ArrivalTime: "18:00", Price: 99.00, Currency: "CAD", Date: "2026-07-02", Class: "economy"},
		{Airline: "Air Canada", FlightNumber: "AC-250", Origin: "Vancouver (YVR)", Destination: "Montreal (YUL)", DepartureTime: "08:30", ArrivalTime: "16:00", Price: 349.00, Currency: "CAD", Date: "2026-07-03", Class: "economy"},
		{Airline: "WestJet", FlightNumber: "WS-250", Origin: "Vancouver (YVR)", Destination: "Montreal (YUL)", DepartureTime: "13:00", ArrivalTime: "20:30", Price: 289.00, Currency: "CAD", Date: "2026-07-03", Class: "economy"},
		{Airline: "Air Canada", FlightNumber: "AC-105", Origin: "Vancouver (YVR)", Destination: "Toronto (YYZ)", DepartureTime: "10:00", ArrivalTime: "17:00", Price: 319.00, Currency: "CAD", Date: "2026-07-07", Class: "economy"},
		{Airline: "WestJet", FlightNumber: "WS-105", Origin: "Vancouver (YVR)", Destination: "Toronto (YYZ)", DepartureTime: "23:00", ArrivalTime: "06:00", Price: 239.00, Currency: "CAD", Date: "2026-07-07", Class: "economy"},
		{Airline: "Air Canada", FlightNumber: "AC-205", Origin: "Montreal (YUL)", Destination: "Toronto (YYZ)", DepartureTime: "12:00", ArrivalTime: "13:15", Price: 169.00, Currency: "CAD", Date: "2026-07-05", Class: "economy"},
		{Airline: "Porter Airlines", FlightNumber: "PD-105", Origin: "Montreal (YUL)", Destination: "Toronto (YTZ)", DepartureTime: "18:00", ArrivalTime: "19:15", Price: 119.00, Currency: "CAD", Date: "2026-07-05", Class: "economy"},
		{Airline: "Delta Air Lines", FlightNumber: "DL-505", Origin: "New York (JFK)", Destination: "Toronto (YYZ)", DepartureTime: "10:00", ArrivalTime: "12:00", Price: 229.00, Currency: "USD", Date: "2026-07-05", Class: "economy"},
		{Airline: "Air Canada", FlightNumber: "AC-305", Origin: "New York (LGA)", Destination: "Toronto (YYZ)", DepartureTime: "19:00", ArrivalTime: "21:00", Price: 189.00, Currency: "USD", Date: "2026-07-05", Class: "economy"},
	}

	seedHotels := []models.SeedHotel{
		{Name: "Cheapest Inn Calgary", City: "Calgary", Address: "123 Center St SW, Calgary, AB", PricePerNight: 89.00, Currency: "CAD", Rating: 2, Amenities: []string{"Free WiFi", "Parking", "Breakfast"}},
		{Name: "Bow River Motel", City: "Calgary", Address: "456 9th Ave SE, Calgary, AB", PricePerNight: 99.00, Currency: "CAD", Rating: 2, Amenities: []string{"Free WiFi", "Parking", "Kitchenette"}},
		{Name: "Mountain View Lodge", City: "Calgary", Address: "789 Bow River Rd, Calgary, AB", PricePerNight: 129.00, Currency: "CAD", Rating: 3, Amenities: []string{"Free WiFi", "Parking", "Restaurant", "Gym"}},
		{Name: "Calgary Downtown Hotel", City: "Calgary", Address: "456 8th Ave SW, Calgary, AB", PricePerNight: 159.00, Currency: "CAD", Rating: 3, Amenities: []string{"Free WiFi", "Restaurant", "Gym", "Pool", "Parking"}},
		{Name: "Suite Dreams Calgary", City: "Calgary", Address: "321 Stephen Ave SW, Calgary, AB", PricePerNight: 249.00, Currency: "CAD", Rating: 4, Amenities: []string{"Free WiFi", "Restaurant", "Gym", "Pool", "Spa", "Parking", "Room Service"}},
		{Name: "The Fairmont Palliser", City: "Calgary", Address: "133 9th Ave SW, Calgary, AB", PricePerNight: 399.00, Currency: "CAD", Rating: 5, Amenities: []string{"Free WiFi", "Restaurant", "Gym", "Pool", "Spa", "Parking", "Room Service", "Concierge", "Valet"}},
		{Name: "Toronto Budget Inn", City: "Toronto", Address: "123 Queen St W, Toronto, ON", PricePerNight: 109.00, Currency: "CAD", Rating: 2, Amenities: []string{"Free WiFi", "Parking", "Breakfast"}},
		{Name: "Toronto Midtown Hotel", City: "Toronto", Address: "456 Bloor St W, Toronto, ON", PricePerNight: 179.00, Currency: "CAD", Rating: 3, Amenities: []string{"Free WiFi", "Restaurant", "Gym", "Parking"}},
		{Name: "Delta Hotels Toronto", City: "Toronto", Address: "75 Lower Simcoe St, Toronto, ON", PricePerNight: 269.00, Currency: "CAD", Rating: 4, Amenities: []string{"Free WiFi", "Restaurant", "Gym", "Pool", "Parking", "Room Service"}},
		{Name: "Fairmont Royal York", City: "Toronto", Address: "100 Front St W, Toronto, ON", PricePerNight: 429.00, Currency: "CAD", Rating: 5, Amenities: []string{"Free WiFi", "Restaurant", "Gym", "Pool", "Spa", "Parking", "Room Service", "Concierge", "Valet"}},
		{Name: "The Ritz-Carlton Toronto", City: "Toronto", Address: "181 Wellington St W, Toronto, ON", PricePerNight: 599.00, Currency: "CAD", Rating: 5, Amenities: []string{"Free WiFi", "Restaurant", "Gym", "Pool", "Spa", "Parking", "Room Service", "Concierge", "Valet", "Butler"}},
		{Name: "Vancouver Downtown Hostel", City: "Vancouver", Address: "111 Burnaby St, Vancouver, BC", PricePerNight: 59.00, Currency: "CAD", Rating: 1, Amenities: []string{"Free WiFi", "Breakfast", "Lockers"}},
		{Name: "Granville Island Hotel", City: "Vancouver", Address: "1253 Johnston St, Vancouver, BC", PricePerNight: 189.00, Currency: "CAD", Rating: 3, Amenities: []string{"Free WiFi", "Restaurant", "Parking", "Pet Friendly"}},
		{Name: "The Westin Bayshore Vancouver", City: "Vancouver", Address: "1601 Bayshore Dr, Vancouver, BC", PricePerNight: 319.00, Currency: "CAD", Rating: 4, Amenities: []string{"Free WiFi", "Restaurant", "Gym", "Pool", "Parking", "Room Service", "Marina Access"}},
		{Name: "Fairmont Pacific Rim", City: "Vancouver", Address: "1038 Canada Pl, Vancouver, BC", PricePerNight: 479.00, Currency: "CAD", Rating: 5, Amenities: []string{"Free WiFi", "Restaurant", "Gym", "Pool", "Spa", "Parking", "Room Service", "Concierge", "Valet"}},
		{Name: "Hotel Montreal Prix Bas", City: "Montreal", Address: "1111 Rue St-Urbain, Montreal, QC", PricePerNight: 99.00, Currency: "CAD", Rating: 2, Amenities: []string{"Free WiFi", "Parking", "Breakfast"}},
		{Name: "Le Centre Sheraton Montreal", City: "Montreal", Address: "1201 Blvd Rene-Levesque W, Montreal, QC", PricePerNight: 219.00, Currency: "CAD", Rating: 4, Amenities: []string{"Free WiFi", "Restaurant", "Gym", "Pool", "Parking", "Room Service"}},
		{Name: "Hotel Le St-James", City: "Montreal", Address: "355 Rue St-Jacques, Montreal, QC", PricePerNight: 399.00, Currency: "CAD", Rating: 5, Amenities: []string{"Free WiFi", "Restaurant", "Gym", "Spa", "Parking", "Room Service", "Concierge"}},
		{Name: "Auberge Alternative Hostel", City: "Montreal", Address: "358 Rue St-Pierre, Montreal, QC", PricePerNight: 49.00, Currency: "CAD", Rating: 1, Amenities: []string{"Free WiFi", "Breakfast", "Lockers", "Common Kitchen"}},
		{Name: "The Manhattan Budget Stay", City: "New York", Address: "891 Amsterdam Ave, New York, NY", PricePerNight: 129.00, Currency: "USD", Rating: 2, Amenities: []string{"Free WiFi", "Breakfast"}},
		{Name: "Hilton Midtown Manhattan", City: "New York", Address: "1335 6th Ave, New York, NY", PricePerNight: 299.00, Currency: "USD", Rating: 4, Amenities: []string{"Free WiFi", "Restaurant", "Gym", "Room Service", "Business Center"}},
		{Name: "The Plaza Hotel", City: "New York", Address: "768 5th Ave, New York, NY", PricePerNight: 699.00, Currency: "USD", Rating: 5, Amenities: []string{"Free WiFi", "Restaurant", "Gym", "Spa", "Room Service", "Concierge", "Valet", "Butler"}},
		{Name: "The Jane Hotel", City: "New York", Address: "113 Jane St, New York, NY", PricePerNight: 159.00, Currency: "USD", Rating: 2, Amenities: []string{"Free WiFi", "Cabaret Lounge"}},
		{Name: "Pod 51 Hotel", City: "New York", Address: "230 E 51st St, New York, NY", PricePerNight: 139.00, Currency: "USD", Rating: 2, Amenities: []string{"Free WiFi", "Rooftop Deck", "Lounge"}},
	}

	seedDestinations := []models.SeedDestination{
		{City: "Calgary", Country: "Canada", Category: "landmark", Title: "Calgary Tower", Description: "Iconic 191-meter observation tower with panoramic views of the city and Rocky Mountains. Features a revolving restaurant.", Address: "101 9th Ave SW, Calgary, AB"},
		{City: "Calgary", Country: "Canada", Category: "park", Title: "Prince's Island Park", Description: "A beautiful 20-hectare urban park on an island in the Bow River. Features walking paths, gardens, and a dog park.", Address: "4th St & 1st Ave SW, Calgary, AB"},
		{City: "Calgary", Country: "Canada", Category: "museum", Title: "Glenbow Museum", Description: "Western Canada's largest museum with over a million artifacts covering art, culture, and history of the Canadian West.", Address: "130 9th Ave SE, Calgary, AB"},
		{City: "Calgary", Country: "Canada", Category: "attraction", Title: "Calgary Zoo", Description: "One of Canada's premier zoos with over 1,000 animals including pandas, penguins, and Canadian wildlife exhibits.", Address: "1300 Zoo Rd NE, Calgary, AB"},
		{City: "Calgary", Country: "Canada", Category: "attraction", Title: "TELUS Spark Science Centre", Description: "Interactive science museum with hands-on exhibits, an IMAX theatre, and a creative kids' museum.", Address: "220 St George's Dr NE, Calgary, AB"},
		{City: "Calgary", Country: "Canada", Category: "landmark", Title: "Stephen Avenue Walk", Description: "Pedestrian-only historic street with shops, restaurants, galleries, and street performers in downtown Calgary.", Address: "Stephen Ave SW, Calgary, AB"},
		{City: "Calgary", Country: "Canada", Category: "museum", Title: "Studio Bell Music Centre", Description: "Home of the National Music Centre with interactive exhibits on Canadian music history, recording studios, and performance spaces.", Address: "850 4th Ave SE, Calgary, AB"},
		{City: "Calgary", Country: "Canada", Category: "landmark", Title: "Heritage Park Historical Village", Description: "Canada's largest living history museum with over 180 exhibits, a steam train, paddlewheeler, and historic buildings from the 1860s to 1950s.", Address: "1900 Heritage Dr SW, Calgary, AB"},
		{City: "Banff", Country: "Canada", Category: "day_trip", Title: "Banff National Park", Description: "Canada's first national park in the Rocky Mountains. Features stunning mountain scenery, hot springs, hiking trails, and wildlife viewing. ~1.5 hour drive from Calgary.", Address: "Banff, AB"},
		{City: "Banff", Country: "Canada", Category: "landmark", Title: "Lake Louise", Description: "Famous turquoise glacial lake in Banff National Park with a historic chateau, hiking trails, and canoeing. ~2 hour drive from Calgary.", Address: "Lake Louise, AB"},
		{City: "Banff", Country: "Canada", Category: "landmark", Title: "Moraine Lake", Description: "Spectacular glacier-fed lake with vibrant blue water surrounded by the Valley of the Ten Peaks. One of the most photographed locations in Canada.", Address: "Moraine Lake Rd, Banff National Park, AB"},
		{City: "Banff", Country: "Canada", Category: "attraction", Title: "Banff Upper Hot Springs", Description: "Natural mineral hot springs with stunning mountain views. Features a large outdoor pool, spa services, and cafe.", Address: "1 Mountain Ave, Banff, AB"},
		{City: "Calgary", Country: "Canada", Category: "event", Title: "Calgary Stampede Grounds", Description: "Home of the world-famous Calgary Stampede rodeo and exhibition. Features the GMC Stadium, BMO Centre, and various event spaces.", Address: "1410 Olympic Way SE, Calgary, AB"},
		{City: "Calgary", Country: "Canada", Category: "day_trip", Title: "Drumheller & Royal Tyrrell Museum", Description: "World-renowned paleontology museum featuring one of the largest dinosaur fossil displays on Earth. ~1.5 hour drive from Calgary.", Address: "1500 N Dinosaur Trail, Drumheller, AB"},
		{City: "Banff", Country: "Canada", Category: "day_trip", Title: "Icefields Parkway", Description: "One of the world's most scenic mountain drives connecting Banff to Jasper through the heart of the Canadian Rockies. 232 km of glaciers, waterfalls, and wildlife.", Address: "Icefields Parkway (Hwy 93), AB"},
		{City: "Calgary", Country: "Canada", Category: "restaurant", Title: "Charcut Roast House", Description: "Award-winning farm-to-table restaurant specializing in house-made charcuterie and wood-fired meats in a rustic-modern setting.", Address: "899 Centre St SW, Calgary, AB"},
		{City: "Toronto", Country: "Canada", Category: "landmark", Title: "CN Tower", Description: "Iconic 553-metre communications and observation tower with a glass floor, revolving restaurant, and EdgeWalk. One of Canada's most recognizable landmarks.", Address: "301 Front St W, Toronto, ON"},
		{City: "Toronto", Country: "Canada", Category: "museum", Title: "Royal Ontario Museum", Description: "Canada's largest museum of natural history and world cultures with over 6 million specimens and exhibits including dinosaurs, minerals, and art.", Address: "100 Queen's Park, Toronto, ON"},
		{City: "Toronto", Country: "Canada", Category: "attraction", Title: "Ripley's Aquarium of Canada", Description: "Massive indoor aquarium featuring a 96-metre moving walkway through a shark tunnel, touch pools, and over 20,000 aquatic animals.", Address: "288 Bremner Blvd, Toronto, ON"},
		{City: "Toronto", Country: "Canada", Category: "park", Title: "High Park", Description: "Toronto's largest public park with 400 acres of trails, gardens, a zoo, sports facilities, and cherry blossoms in spring.", Address: "1873 Bloor St W, Toronto, ON"},
		{City: "Toronto", Country: "Canada", Category: "landmark", Title: "Distillery District", Description: "Historic pedestrian-only neighbourhood with Victorian industrial architecture, art galleries, boutiques, restaurants, and seasonal markets.", Address: "55 Mill St, Toronto, ON"},
		{City: "Toronto", Country: "Canada", Category: "landmark", Title: "St. Lawrence Market", Description: "A historic public market featuring over 120 vendors selling fresh produce, meat, cheese, baked goods, and prepared foods. Operating since 1803.", Address: "93 Front St E, Toronto, ON"},
		{City: "Toronto", Country: "Canada", Category: "attraction", Title: "Toronto Islands", Description: "A chain of small islands just offshore downtown Toronto with beaches, parks, bike paths, a small airport, and skyline views.", Address: "Ferry Docks, 9 Queens Quay W, Toronto, ON"},
		{City: "Niagara Falls", Country: "Canada", Category: "day_trip", Title: "Niagara Falls", Description: "World-famous waterfall on the Niagara River, featuring the Horseshoe Falls, boat tours behind the falls, and stunning illumination at night.", Address: "Niagara Falls, ON"},
		{City: "Toronto", Country: "Canada", Category: "restaurant", Title: "St. Lawrence Market Food Court", Description: "Diverse food court in the historic St. Lawrence Market with peameal bacon sandwiches, sushi, Chinese BBQ, and fresh seafood.", Address: "93 Front St E, Toronto, ON"},
		{City: "Vancouver", Country: "Canada", Category: "park", Title: "Stanley Park", Description: "A 405-hectare urban park with a famous seawall, ancient cedar forests, beaches, the Vancouver Aquarium, and stunning mountain and ocean views.", Address: "Vancouver, BC"},
		{City: "Vancouver", Country: "Canada", Category: "landmark", Title: "Granville Island", Description: "A vibrant peninsula with a public market, artisan studios, theatres, breweries, and restaurants beneath the Granville Street Bridge.", Address: "1669 Johnston St, Vancouver, BC"},
		{City: "Vancouver", Country: "Canada", Category: "attraction", Title: "Capilano Suspension Bridge", Description: "A 137-metre suspension bridge spanning 70 metres above the Capilano River, surrounded by old-growth rainforest with treetop walkways.", Address: "3735 Capilano Rd, North Vancouver, BC"},
		{City: "Vancouver", Country: "Canada", Category: "attraction", Title: "Grouse Mountain", Description: "Year-round mountain destination with skiing, hiking, zip-lining, a wildlife refuge with grizzly bears, and panoramic city views.", Address: "6400 Nancy Greene Way, North Vancouver, BC"},
		{City: "Vancouver", Country: "Canada", Category: "landmark", Title: "English Bay Beach", Description: "One of Vancouver's most popular urban beaches with sunset views, swimming, walking paths, and beachside restaurants and cafes.", Address: "Beach Ave, Vancouver, BC"},
		{City: "Vancouver", Country: "Canada", Category: "museum", Title: "Museum of Anthropology (MOA)", Description: "World-renowned museum at UBC showcasing Pacific Northwest First Nations art and culture, including totem poles, masks, and contemporary works.", Address: "6393 NW Marine Dr, Vancouver, BC"},
		{City: "Vancouver", Country: "Canada", Category: "day_trip", Title: "Whistler Blackcomb", Description: "One of North America's largest ski resorts with year-round activities including mountain biking, hiking, zip-lining, and peak-to-peak gondola.", Address: "Whistler, BC"},
		{City: "Montreal", Country: "Canada", Category: "landmark", Title: "Old Montreal (Vieux-Montreal)", Description: "Historic neighbourhood with cobblestone streets, 17th-century architecture, Notre-Dame Basilica, Place Jacques-Cartier, and charming cafes.", Address: "Old Montreal, QC"},
		{City: "Montreal", Country: "Canada", Category: "landmark", Title: "Notre-Dame Basilica", Description: "Stunning 19th-century Gothic Revival church with a breathtaking blue vaulted ceiling, intricate woodwork, and a spectacular light and sound show.", Address: "110 Rue Notre-Dame O, Montreal, QC"},
		{City: "Montreal", Country: "Canada", Category: "park", Title: "Mount Royal Park", Description: "Large urban park designed by Frederick Law Olmsted with hiking trails, a lake, lookout points, and the iconic Mount Royal Cross.", Address: "1260 Chem. Remembrance, Montreal, QC"},
		{City: "Montreal", Country: "Canada", Category: "museum", Title: "Montreal Museum of Fine Arts", Description: "One of Canada's largest art museums with a diverse collection of over 44,000 works spanning antiquities to contemporary art.", Address: "1380 Rue Sherbrooke O, Montreal, QC"},
		{City: "Montreal", Country: "Canada", Category: "attraction", Title: "Montreal Botanical Garden", Description: "One of the world's largest botanical gardens with 75 hectares of themed gardens, greenhouses, and the famous Chinese and Japanese gardens.", Address: "4101 Rue Sherbrooke E, Montreal, QC"},
		{City: "Montreal", Country: "Canada", Category: "landmark", Title: "Plateau-Mont-Royal", Description: "Trendy neighbourhood known for its colourful row houses, independent boutiques, street art, and some of Montreal's best bagels and brunch spots.", Address: "Plateau-Mont-Royal, Montreal, QC"},
		{City: "Montreal", Country: "Canada", Category: "restaurant", Title: "Schwartz's Deli", Description: "Legendary Montreal deli famous for smoked meat sandwiches, serving since 1928. A quintessential Montreal food experience.", Address: "3895 Blvd St-Laurent, Montreal, QC"},
		{City: "New York", Country: "USA", Category: "landmark", Title: "Statue of Liberty & Ellis Island", Description: "Iconic American symbol and UNESCO World Heritage site. Ferry ride to Liberty Island includes museum access and panoramic NYC skyline views.", Address: "Liberty Island, New York, NY"},
		{City: "New York", Country: "USA", Category: "park", Title: "Central Park", Description: "World-famous 843-acre urban park featuring the Central Park Zoo, Bethesda Fountain, Bow Bridge, Strawberry Fields, and miles of walking paths.", Address: "Central Park, New York, NY"},
		{City: "New York", Country: "USA", Category: "museum", Title: "The Metropolitan Museum of Art", Description: "One of the world's largest and finest art museums with over 2 million works spanning 5,000 years of human history and culture.", Address: "1000 5th Ave, New York, NY"},
		{City: "New York", Country: "USA", Category: "landmark", Title: "Times Square", Description: "Iconic commercial and entertainment hub known for its dazzling digital billboards, Broadway theatres, and New Year's Eve ball drop.", Address: "Manhattan, NY"},
		{City: "New York", Country: "USA", Category: "landmark", Title: "Brooklyn Bridge", Description: "Historic hybrid cable-stayed/suspension bridge connecting Manhattan and Brooklyn. Offers a scenic pedestrian walkway with skyline views.", Address: "Brooklyn Bridge, New York, NY"},
		{City: "New York", Country: "USA", Category: "attraction", Title: "Broadway & Theatre District", Description: "World's leading theatre district centred around Times Square with dozens of historic theatres producing musicals, plays, and avant-garde performances.", Address: "Theatre District, New York, NY"},
		{City: "New York", Country: "USA", Category: "park", Title: "The High Line", Description: "A 1.45-mile elevated linear park built on a historic freight rail line above Manhattan's West Side. Features gardens, art installations, and food vendors.", Address: "New York, NY"},
		{City: "New York", Country: "USA", Category: "landmark", Title: "Empire State Building", Description: "Art Deco skyscraper with 86th and 102nd-floor observation decks offering sweeping 360-degree views of New York City.", Address: "350 5th Ave, New York, NY"},
	}

	for _, sf := range seedFlights {
		date, err := time.Parse("2006-01-02", sf.Date)
		if err != nil {
			log.Fatalf("Invalid flight date %q: %v", sf.Date, err)
		}
		textForEmbedding := fmt.Sprintf("Flight %s %s from %s to %s on %s %s class $%.2f",
			sf.Airline, sf.FlightNumber, sf.Origin, sf.Destination, sf.Date, sf.Class, sf.Price)
		emb, err := embedder.GenerateEmbedding(ctx, textForEmbedding)
		if err != nil {
			log.Fatalf("Failed to embed flight %s: %v", sf.FlightNumber, err)
		}
		f := &models.Flight{
			Airline:       sf.Airline,
			FlightNumber:  sf.FlightNumber,
			Origin:        sf.Origin,
			Destination:   sf.Destination,
			DepartureTime: sf.DepartureTime,
			ArrivalTime:   sf.ArrivalTime,
			Price:         sf.Price,
			Currency:      sf.Currency,
			Date:          date,
			Class:         sf.Class,
		}
		if err := db.InsertFlight(ctx, f, emb); err != nil {
			log.Fatalf("Failed to insert flight %s: %v", sf.FlightNumber, err)
		}
	}
	log.Printf("[%d] Flights inserted.", len(seedFlights))

	for _, sh := range seedHotels {
		textForEmbedding := fmt.Sprintf("Hotel %s in %s %d-star $%.2f per night amenities: %s",
			sh.Name, sh.City, sh.Rating, sh.PricePerNight, sh.Amenities)
		emb, err := embedder.GenerateEmbedding(ctx, textForEmbedding)
		if err != nil {
			log.Fatalf("Failed to embed hotel %s: %v", sh.Name, err)
		}
		h := &models.Hotel{
			Name:          sh.Name,
			City:          sh.City,
			Address:       sh.Address,
			PricePerNight: sh.PricePerNight,
			Currency:      sh.Currency,
			Rating:        sh.Rating,
			Amenities:     sh.Amenities,
		}
		if err := db.InsertHotel(ctx, h, emb); err != nil {
			log.Fatalf("Failed to insert hotel %s: %v", sh.Name, err)
		}
	}
	log.Printf("[%d] Hotels inserted.", len(seedHotels))

	for _, sd := range seedDestinations {
		textForEmbedding := fmt.Sprintf("%s in %s %s: %s", sd.Title, sd.City, sd.Category, sd.Description)
		emb, err := embedder.GenerateEmbedding(ctx, textForEmbedding)
		if err != nil {
			log.Fatalf("Failed to embed destination %s: %v", sd.Title, err)
		}
		d := &models.Destination{
			City:        sd.City,
			Country:     sd.Country,
			Category:    sd.Category,
			Title:       sd.Title,
			Description: sd.Description,
			Address:     sd.Address,
		}
		if err := db.InsertDestination(ctx, d, emb); err != nil {
			log.Fatalf("Failed to insert destination %s: %v", sd.Title, err)
		}
	}
	log.Printf("[%d] Destinations inserted.", len(seedDestinations))
}
