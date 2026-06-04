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

	// Build connecting routes from all available flights
	var connectingRoutes []models.ConnectingFlight
	allFlights, err := h.db.GetAllFlights(r.Context())
	if err == nil && len(allFlights) > 0 {
		// Build route graph: city -> list of (destination city, flight)
		type routeEntry struct {
			destCity string
			flight   models.Flight
		}
		routeGraph := make(map[string][]routeEntry)
		citiesSeen := make(map[string]bool)
		for _, f := range allFlights {
			oc := extractCity(f.Origin)
			dc := extractCity(f.Destination)
			routeGraph[oc] = append(routeGraph[oc], routeEntry{destCity: dc, flight: f})
			citiesSeen[oc] = true
			citiesSeen[dc] = true
		}

		// Collect city names from query and top flight results
		mentioned := make(map[string]bool)
		queryLower := strings.ToLower(req.Query)
		for city := range citiesSeen {
			if strings.Contains(queryLower, strings.ToLower(city)) {
				mentioned[city] = true
			}
		}
		for _, f := range flights {
			mentioned[extractCity(f.Origin)] = true
			mentioned[extractCity(f.Destination)] = true
		}

		// Find 1-stop connections between mentioned cities
		for origin := range mentioned {
			for dest := range mentioned {
				if origin == dest {
					continue
				}
				if hasDirectFlight(allFlights, origin, dest) {
					continue
				}
				for _, leg1 := range routeGraph[origin] {
					hub := leg1.destCity
					if hub == dest {
						continue
					}
					for _, leg2 := range routeGraph[hub] {
						if leg2.destCity == dest {
							total := leg1.flight.Price + leg2.flight.Price
							if leg1.flight.Currency == "USD" && leg2.flight.Currency == "CAD" {
								total = leg1.flight.Price + leg2.flight.Price*0.75
							} else if leg2.flight.Currency == "USD" && leg1.flight.Currency == "CAD" {
								total = leg1.flight.Price*0.75 + leg2.flight.Price
							}
							connectingRoutes = append(connectingRoutes, models.ConnectingFlight{
								From:     leg1.flight,
								To:       leg2.flight,
								Hub:      hub,
								TotalUSD: total,
							})
						}
					}
				}
			}
		}

		// Add connecting flight segments to flight list for context
		seen := make(map[string]bool)
		for _, f := range flights {
			seen[f.ID] = true
		}
		for _, cr := range connectingRoutes {
			if !seen[cr.From.ID] {
				flights = append(flights, cr.From)
				seen[cr.From.ID] = true
			}
			if !seen[cr.To.ID] {
				flights = append(flights, cr.To)
				seen[cr.To.ID] = true
			}
		}
	}

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

	if len(connectingRoutes) > 0 {
		contextBuilder.WriteString("\n## Connecting Flights (1-stop)\n\n")
		for _, cr := range connectingRoutes {
			contextBuilder.WriteString(fmt.Sprintf("- %s %s %s → **%s** → %s %s %s | $%.0f total\n  Leg 1: %s %s → %s\n  Leg 2: %s %s → %s\n",
				cr.From.Airline, cr.From.FlightNumber, cr.From.Origin,
				cr.Hub,
				cr.To.Airline, cr.To.FlightNumber, cr.To.Destination,
				cr.TotalUSD,
				cr.From.Airline, cr.From.Origin, cr.From.Destination,
				cr.To.Airline, cr.To.Origin, cr.To.Destination,
			))
		}
	}

	var answer string
	chatStart := time.Now()

	// Short-circuit: if the query asks about a direct flight between two cities, check explicitly
	if directOrigin, directDest, ok := parseDirectFlightQuery(req.Query); ok {
		found := false
		for _, f := range allFlights {
			if extractCity(f.Origin) == directOrigin && extractCity(f.Destination) == directDest {
				found = true
				break
			}
		}
		if !found {
			answer = fmt.Sprintf("There is no direct flight from %s to %s.", directOrigin, directDest)
			chatTime := time.Since(chatStart)
			totalTime := time.Since(startTotal)
			resp := models.TravelSearchResponse{
				Answer:    answer,
				Flights:   flights,
				Hotels:    hotels,
				Destinations: destinations,
				SessionID: sessionID,
				Mode:      req.Mode,
				TimingEmbedMs:  embedTime.Milliseconds(),
				TimingSearchMs: searchTime.Milliseconds(),
				TimingChatMs:   0,
				TimingTotalMs:  totalTime.Milliseconds(),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
	}

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
		if len(connectingRoutes) > 0 {
			sb.WriteString("\n# Connecting Flights (1-stop)\n\n")
			for _, cr := range connectingRoutes {
				sb.WriteString(fmt.Sprintf("- %s %s: %s → **%s** → %s | $%.0f total\n  *%s %s → %s then %s %s → %s*\n",
					cr.From.Airline, cr.From.FlightNumber, cr.From.Origin,
					cr.Hub, cr.To.Destination, cr.TotalUSD,
					cr.From.Airline, cr.From.Origin, cr.From.Destination,
					cr.To.Airline, cr.To.Origin, cr.To.Destination,
				))
			}
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
		Answer:           answer,
		Flights:          flights,
		Hotels:           hotels,
		Destinations:     destinations,
		ConnectingRoutes: connectingRoutes,
		SessionID:        sessionID,
		Mode:             req.Mode,
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

func extractCity(location string) string {
	idx := strings.LastIndex(location, " (")
	if idx > 0 {
		return location[:idx]
	}
	return location
}

func hasDirectFlight(flights []models.Flight, origin, dest string) bool {
	for _, f := range flights {
		if extractCity(f.Origin) == origin && extractCity(f.Destination) == dest {
			return true
		}
	}
	return false
}

// knownCities is the set of cities in our travel database
var knownCities = []string{"Calgary", "Toronto", "Vancouver", "Montreal", "New York", "Edmonton", "Banff", "Niagara Falls", "Whistler"}

func parseDirectFlightQuery(query string) (origin, dest string, ok bool) {
	lower := strings.ToLower(query)
	if !strings.Contains(lower, "direct flight") {
		return "", "", false
	}

	// Extract city names from the query that are in our known list
	var found []string
	queryLower := strings.ToLower(query)
	for _, city := range knownCities {
		if strings.Contains(queryLower, strings.ToLower(city)) {
			found = append(found, city)
		}
	}

	// Need exactly 2 cities found: origin and destination
	if len(found) < 2 {
		return "", "", false
	}

	// Try to determine order: look for "from X to Y" or "between X and Y"
	fromIdx := strings.Index(lower, "from ")
	toIdx := strings.Index(lower, " to ")
	if fromIdx >= 0 && toIdx > fromIdx {
		// "from X to Y" — X is origin, Y is destination
		for _, c := range found {
			cLower := strings.ToLower(c)
			ci := strings.Index(lower, cLower)
			if ci > fromIdx && ci < toIdx {
				origin = c
			} else if ci > toIdx {
				dest = c
			}
		}
		if origin != "" && dest != "" {
			return origin, dest, true
		}
	}

	// Fallback: first found = origin, second = dest
	return found[0], found[1], true
}
