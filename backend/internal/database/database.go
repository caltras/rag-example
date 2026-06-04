package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	pgxvector "github.com/pgvector/pgvector-go/pgx"

	"rag-backend/internal/models"
)

type DB struct {
	pool *pgxpool.Pool
}

func New(databaseURL string) (*DB, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database URL: %w", err)
	}
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return pgxvector.RegisterTypes(ctx, conn)
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	return &DB{pool: pool}, nil
}

func (db *DB) Close() {
	db.pool.Close()
}

func (db *DB) InsertArticle(ctx context.Context, article *models.Article, embedding []float32) error {
	vec := pgvector.NewVector(embedding)
	_, err := db.pool.Exec(ctx,
		`INSERT INTO articles (title, date, text, author, embedding)
		 VALUES ($1, $2, $3, $4, $5)`,
		article.Title, article.Date, article.Text, article.Author, vec,
	)
	return err
}

func (db *DB) SearchSimilar(ctx context.Context, embedding []float32, query string, limit int) ([]models.Article, error) {
	vec := pgvector.NewVector(embedding)
	rows, err := db.pool.Query(ctx,
		`SELECT id, title, date, text, author, created_at
		 FROM articles
		 ORDER BY (1 - (embedding <=> $1)) * 0.5 +
		   COALESCE(ts_rank(text_search, plainto_tsquery('english', $2)), 0) * 0.5 DESC
		 LIMIT $3`,
		vec, query, limit,
	)
	if err != nil {
		log.Printf("SQL error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var articles []models.Article
	for rows.Next() {
		var a models.Article
		if err := rows.Scan(&a.ID, &a.Title, &a.Date, &a.Text, &a.Author, &a.CreatedAt); err != nil {
			return nil, err
		}
		articles = append(articles, a)
	}
	return articles, nil
}

func (db *DB) ArticleCount(ctx context.Context) (int, error) {
	var count int
	err := db.pool.QueryRow(ctx, "SELECT COUNT(*) FROM articles").Scan(&count)
	return count, err
}

func (db *DB) HotelCount(ctx context.Context) (int, error) {
	var count int
	err := db.pool.QueryRow(ctx, "SELECT COUNT(*) FROM hotels").Scan(&count)
	return count, err
}

func (db *DB) InsertFlight(ctx context.Context, f *models.Flight, embedding []float32) error {
	vec := pgvector.NewVector(embedding)
	_, err := db.pool.Exec(ctx,
		`INSERT INTO flights (airline, flight_number, origin, destination, departure_time, arrival_time, price, currency, date, class, embedding)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		f.Airline, f.FlightNumber, f.Origin, f.Destination, f.DepartureTime, f.ArrivalTime, f.Price, f.Currency, f.Date, f.Class, vec,
	)
	return err
}

func (db *DB) InsertHotel(ctx context.Context, h *models.Hotel, embedding []float32) error {
	vec := pgvector.NewVector(embedding)
	_, err := db.pool.Exec(ctx,
		`INSERT INTO hotels (name, city, address, price_per_night, currency, rating, amenities, embedding)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		h.Name, h.City, h.Address, h.PricePerNight, h.Currency, h.Rating, h.Amenities, vec,
	)
	return err
}

func (db *DB) InsertDestination(ctx context.Context, d *models.Destination, embedding []float32) error {
	vec := pgvector.NewVector(embedding)
	_, err := db.pool.Exec(ctx,
		`INSERT INTO destinations (city, country, category, title, description, address, embedding)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		d.City, d.Country, d.Category, d.Title, d.Description, d.Address, vec,
	)
	return err
}

func (db *DB) SearchFlights(ctx context.Context, embedding []float32, query string, limit int) ([]models.Flight, error) {
	vec := pgvector.NewVector(embedding)
	rows, err := db.pool.Query(ctx,
		`SELECT id, airline, flight_number, origin, destination, departure_time, arrival_time, price, currency, date, class, created_at
		 FROM flights
		 ORDER BY (1 - (embedding <=> $1)) * 0.5 +
		   COALESCE(ts_rank(text_search, plainto_tsquery('english', $2)), 0) * 0.5 DESC
		 LIMIT $3`,
		vec, query, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var flights []models.Flight
	for rows.Next() {
		var f models.Flight
		if err := rows.Scan(&f.ID, &f.Airline, &f.FlightNumber, &f.Origin, &f.Destination, &f.DepartureTime, &f.ArrivalTime, &f.Price, &f.Currency, &f.Date, &f.Class, &f.CreatedAt); err != nil {
			return nil, err
		}
		flights = append(flights, f)
	}
	return flights, nil
}

func (db *DB) SearchHotels(ctx context.Context, embedding []float32, query string, limit int) ([]models.Hotel, error) {
	vec := pgvector.NewVector(embedding)
	rows, err := db.pool.Query(ctx,
		`SELECT id, name, city, address, price_per_night, currency, rating, amenities, created_at
		 FROM hotels
		 ORDER BY (1 - (embedding <=> $1)) * 0.5 +
		   COALESCE(ts_rank(text_search, plainto_tsquery('english', $2)), 0) * 0.5 DESC
		 LIMIT $3`,
		vec, query, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var hotels []models.Hotel
	for rows.Next() {
		var h models.Hotel
		if err := rows.Scan(&h.ID, &h.Name, &h.City, &h.Address, &h.PricePerNight, &h.Currency, &h.Rating, &h.Amenities, &h.CreatedAt); err != nil {
			return nil, err
		}
		hotels = append(hotels, h)
	}
	return hotels, nil
}

func (db *DB) SearchDestinations(ctx context.Context, embedding []float32, query string, limit int) ([]models.Destination, error) {
	vec := pgvector.NewVector(embedding)
	rows, err := db.pool.Query(ctx,
		`SELECT id, city, country, category, title, description, address, created_at
		 FROM destinations
		 ORDER BY (1 - (embedding <=> $1)) * 0.5 +
		   COALESCE(ts_rank(text_search, plainto_tsquery('english', $2)), 0) * 0.5 DESC
		 LIMIT $3`,
		vec, query, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var dests []models.Destination
	for rows.Next() {
		var d models.Destination
		if err := rows.Scan(&d.ID, &d.City, &d.Country, &d.Category, &d.Title, &d.Description, &d.Address, &d.CreatedAt); err != nil {
			return nil, err
		}
		dests = append(dests, d)
	}
	return dests, nil
}
