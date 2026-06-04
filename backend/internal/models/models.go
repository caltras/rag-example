package models

import "time"

type Article struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Date      time.Time `json:"date"`
	Text      string    `json:"text"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

type SearchRequest struct {
	Query     string `json:"query"`
	SessionID string `json:"session_id"`
	Mode      string `json:"mode"`
}

type SearchResponse struct {
	Answer          string        `json:"answer"`
	Sources         []string      `json:"sources"`
	SessionID       string        `json:"session_id"`
	TimingEmbedMs   int64         `json:"timing_embed_ms"`
	TimingSearchMs  int64         `json:"timing_search_ms"`
	TimingChatMs    int64         `json:"timing_chat_ms"`
	TimingTotalMs   int64         `json:"timing_total_ms"`
	Mode            string        `json:"mode"`
	DataOnly        string        `json:"data_only,omitempty"`
}

type SeedArticle struct {
	Title  string `json:"title"`
	Date   string `json:"date"`
	Text   string `json:"text"`
	Author string `json:"author"`
}

type Flight struct {
	ID            string    `json:"id"`
	Airline       string    `json:"airline"`
	FlightNumber  string    `json:"flight_number"`
	Origin        string    `json:"origin"`
	Destination   string    `json:"destination"`
	DepartureTime string    `json:"departure_time"`
	ArrivalTime   string    `json:"arrival_time"`
	Price         float64   `json:"price"`
	Currency      string    `json:"currency"`
	Date          time.Time `json:"date"`
	Class         string    `json:"class"`
	CreatedAt     time.Time `json:"created_at"`
}

type Hotel struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	City          string    `json:"city"`
	Address       string    `json:"address"`
	PricePerNight float64   `json:"price_per_night"`
	Currency      string    `json:"currency"`
	Rating        int       `json:"rating"`
	Amenities     []string  `json:"amenities"`
	CreatedAt     time.Time `json:"created_at"`
}

type Destination struct {
	ID          string    `json:"id"`
	City        string    `json:"city"`
	Country     string    `json:"country"`
	Category    string    `json:"category"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Address     string    `json:"address"`
	CreatedAt   time.Time `json:"created_at"`
}

type TravelSearchRequest struct {
	Query     string `json:"query"`
	SessionID string `json:"session_id"`
	Mode      string `json:"mode"`
}

type ConnectingFlight struct {
	From     Flight `json:"from"`
	To       Flight `json:"to"`
	Hub      string `json:"hub"`
	TotalUSD float64 `json:"total_usd"`
}

type TravelSearchResponse struct {
	Answer            string             `json:"answer"`
	Flights           []Flight           `json:"flights"`
	Hotels            []Hotel            `json:"hotels"`
	Destinations      []Destination      `json:"destinations"`
	ConnectingRoutes  []ConnectingFlight `json:"connecting_routes,omitempty"`
	SessionID         string             `json:"session_id"`
	Mode              string             `json:"mode"`
	TimingEmbedMs     int64              `json:"timing_embed_ms"`
	TimingSearchMs    int64              `json:"timing_search_ms"`
	TimingChatMs      int64              `json:"timing_chat_ms"`
	TimingTotalMs     int64              `json:"timing_total_ms"`
}

type SeedFlight struct {
	Airline       string
	FlightNumber  string
	Origin        string
	Destination   string
	DepartureTime string
	ArrivalTime   string
	Price         float64
	Currency      string
	Date          string
	Class         string
}

type SeedHotel struct {
	Name          string
	City          string
	Address       string
	PricePerNight float64
	Currency      string
	Rating        int
	Amenities     []string
}

type SeedDestination struct {
	City        string
	Country     string
	Category    string
	Title       string
	Description string
	Address     string
}
