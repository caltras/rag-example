CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS articles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    date DATE NOT NULL,
    text TEXT NOT NULL,
    author TEXT NOT NULL,
    embedding VECTOR(1024),
    created_at TIMESTAMPTZ DEFAULT NOW()
);
