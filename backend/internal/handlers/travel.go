package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"rag-backend/internal/llm"
	"rag-backend/internal/models"
)

func (h *Handler) TravelSearch(w http.ResponseWriter, r *http.Request) {
	startTotal := time.Now()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.TravelSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Query) == "" {
		http.Error(w, "Query is required", http.StatusBadRequest)
		return
	}

	convo := h.getOrCreateConvo("travel_" + req.SessionID)
	sessionID := req.SessionID

	embedStart := time.Now()
	queryEmbedding, err := h.embedder.GenerateEmbedding(r.Context(), req.Query)
	if err != nil {
		log.Printf("travel embedding error: %v", err)
		http.Error(w, "Failed to generate embedding", http.StatusInternalServerError)
		return
	}
	embedTime := time.Since(embedStart)
	log.Printf("travel_embedding: completed in %v", embedTime)

	searchStart := time.Now()
	flights, err := h.db.SearchFlights(r.Context(), queryEmbedding, req.Query, 5)
	if err != nil {
		log.Printf("travel flight search error: %v", err)
	}
	hotels, err := h.db.SearchHotels(r.Context(), queryEmbedding, req.Query, 5)
	if err != nil {
		log.Printf("travel hotel search error: %v", err)
	}
	destinations, err := h.db.SearchDestinations(r.Context(), queryEmbedding, req.Query, 5)
	if err != nil {
		log.Printf("travel destination search error: %v", err)
	}
	searchTime := time.Since(searchStart)
	log.Printf("travel_search_similarity: found %d flights, %d hotels, %d destinations in %v",
		len(flights), len(hotels), len(destinations), searchTime)

	if len(flights) == 0 && len(hotels) == 0 && len(destinations) == 0 {
		totalTime := time.Since(startTotal)
		resp := models.TravelSearchResponse{
			Answer:          "No travel information found. Try asking about a specific destination.",
			TimingEmbedMs:   embedTime.Milliseconds(),
			TimingSearchMs:  searchTime.Milliseconds(),
			TimingTotalMs:   totalTime.Milliseconds(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	var contextBuilder strings.Builder

	contextBuilder.WriteString("## Available Flights\n\n")
	for _, f := range flights {
		contextBuilder.WriteString(fmt.Sprintf(
			"- %s %s: %s → %s, departs %s, arrives %s, $%.2f %s (%s)\n",
			f.Airline, f.FlightNumber, f.Origin, f.Destination,
			f.DepartureTime, f.ArrivalTime, f.Price, f.Currency, f.Class,
		))
	}

	contextBuilder.WriteString("\n## Available Hotels\n\n")
	for _, h := range hotels {
		am := strings.Join(h.Amenities, ", ")
		contextBuilder.WriteString(fmt.Sprintf(
			"- %s (%d-star) — $%.2f %s/night\n  Address: %s\n  Amenities: %s\n",
			h.Name, h.Rating, h.PricePerNight, h.Currency, h.Address, am,
		))
	}

	contextBuilder.WriteString("\n## Points of Interest & Destinations\n\n")
	for _, d := range destinations {
		contextBuilder.WriteString(fmt.Sprintf(
			"- %s (%s): %s\n  Address: %s\n",
			d.Title, d.Category, d.Description, d.Address,
		))
	}

	var answer string
	chatStart := time.Now()

	if req.Mode == "data" {
		var sb strings.Builder
		sb.WriteString("# Flights\n\n")
		for _, f := range flights {
			sb.WriteString(fmt.Sprintf("- **%s %s** %s → %s | %s–%s | $%.0f %s\n",
				f.Airline, f.FlightNumber, f.Origin, f.Destination,
				f.DepartureTime, f.ArrivalTime, f.Price, f.Currency))
		}
		sb.WriteString("\n# Hotels\n\n")
		for _, h := range hotels {
			sb.WriteString(fmt.Sprintf("- **%s** (%s) ★%d | $%.0f/night | %s\n",
				h.Name, h.City, h.Rating, h.PricePerNight, h.Address))
		}
		sb.WriteString("\n# Destinations\n\n")
		for _, d := range destinations {
			sb.WriteString(fmt.Sprintf("- **%s** (%s): %s\n  *%s*\n",
				d.Title, d.Category, d.Description, d.Address))
		}
		answer = sb.String()
	} else {
		systemPrompt := `You are a travel planning assistant for a RAG travel database. You MUST follow these rules strictly:

1. ONLY use the flight, hotel, and destination data provided below to create your travel plan.
2. If the data does not contain enough information to fulfill the request, say: "I cannot answer that based on the available travel data."
3. Do not answer questions unrelated to travel planning (e.g. general chat, opinions, calculations outside the data).
4. Cite specific names, prices, and addresses from the provided data.
5. When recommending a hotel, calculate and show the total cost for the stay duration.`

		userPrompt := fmt.Sprintf(
			"Here is the travel data retrieved:\n\n%s\n\nBased on this data, answer the traveller's request: %s",
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

		answer, err = h.chat.Chat(r.Context(), messages)
		if err != nil {
			log.Printf("travel chat error: %v", err)
			http.Error(w, "Failed to generate travel plan", http.StatusInternalServerError)
			return
		}
		log.Printf("travel_chat_completion: completed in %v", time.Since(chatStart))

		if convo != nil {
			h.appendHistory(convo, "user", req.Query)
			h.appendHistory(convo, "assistant", answer)
		}
	}

	chatTime := time.Since(chatStart)
	totalTime := time.Since(startTotal)

	resp := models.TravelSearchResponse{
		Answer:       answer,
		Flights:      flights,
		Hotels:       hotels,
		Destinations: destinations,
		SessionID:    sessionID,
		Mode:         req.Mode,
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
