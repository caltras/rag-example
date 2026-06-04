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

func (db *DB) SearchSimilar(ctx context.Context, embedding []float32, limit int) ([]models.Article, error) {
	vec := pgvector.NewVector(embedding)
	rows, err := db.pool.Query(ctx,
		`SELECT id, title, date, text, author, created_at
		 FROM articles
		 ORDER BY embedding <=> $1
		 LIMIT $2`,
		vec, limit,
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
