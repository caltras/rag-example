ALTER TABLE articles ADD COLUMN IF NOT EXISTS text_search tsvector
  GENERATED ALWAYS AS (to_tsvector('english', coalesce(title, '') || ' ' || coalesce(text, ''))) STORED;

ALTER TABLE flights ADD COLUMN IF NOT EXISTS text_search tsvector
  GENERATED ALWAYS AS (to_tsvector('english', coalesce(airline, '') || ' ' || coalesce(flight_number, '') || ' ' || coalesce(origin, '') || ' ' || coalesce(destination, '') || ' ' || coalesce(class, ''))) STORED;

ALTER TABLE hotels ADD COLUMN IF NOT EXISTS text_search tsvector;

UPDATE hotels SET text_search = to_tsvector('english', coalesce(name, '') || ' ' || coalesce(city, '') || ' ' || coalesce(address, '') || ' ' || coalesce(array_to_string(amenities, ' '), ''));

CREATE OR REPLACE FUNCTION hotels_tsvector_trigger() RETURNS trigger AS $$
BEGIN
  NEW.text_search := to_tsvector('english', coalesce(NEW.name, '') || ' ' || coalesce(NEW.city, '') || ' ' || coalesce(NEW.address, '') || ' ' || coalesce(array_to_string(NEW.amenities, ' '), ''));
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_hotels_tsvector ON hotels;
CREATE TRIGGER trg_hotels_tsvector BEFORE INSERT OR UPDATE ON hotels
  FOR EACH ROW EXECUTE FUNCTION hotels_tsvector_trigger();

ALTER TABLE destinations ADD COLUMN IF NOT EXISTS text_search tsvector
  GENERATED ALWAYS AS (to_tsvector('english', coalesce(title, '') || ' ' || coalesce(city, '') || ' ' || coalesce(category, '') || ' ' || coalesce(description, ''))) STORED;

CREATE INDEX IF NOT EXISTS idx_articles_fts ON articles USING GIN(text_search);
CREATE INDEX IF NOT EXISTS idx_flights_fts ON flights USING GIN(text_search);
CREATE INDEX IF NOT EXISTS idx_hotels_fts ON hotels USING GIN(text_search);
CREATE INDEX IF NOT EXISTS idx_destinations_fts ON destinations USING GIN(text_search);
