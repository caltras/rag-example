CREATE TABLE IF NOT EXISTS flights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    airline TEXT NOT NULL,
    flight_number TEXT NOT NULL,
    origin TEXT NOT NULL,
    destination TEXT NOT NULL,
    departure_time TIME NOT NULL,
    arrival_time TIME NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'CAD',
    date DATE NOT NULL,
    class TEXT NOT NULL DEFAULT 'economy',
    embedding VECTOR(1024),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS hotels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    city TEXT NOT NULL,
    address TEXT NOT NULL,
    price_per_night DECIMAL(10,2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'CAD',
    rating INTEGER NOT NULL DEFAULT 3,
    amenities TEXT[] DEFAULT '{}',
    embedding VECTOR(1024),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS destinations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    city TEXT NOT NULL,
    country TEXT NOT NULL,
    category TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    address TEXT DEFAULT '',
    embedding VECTOR(1024),
    created_at TIMESTAMPTZ DEFAULT NOW()
);
