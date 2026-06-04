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
	Query string `json:"query"`
}

type SearchResponse struct {
	Answer  string   `json:"answer"`
	Sources []string `json:"sources"`
}

type SeedArticle struct {
	Title  string `json:"title"`
	Date   string `json:"date"`
	Text   string `json:"text"`
	Author string `json:"author"`
}
