CREATE TABLE IF NOT EXISTS auditoriums (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS showtimes (
    id UUID PRIMARY KEY,
    movie_id UUID NOT NULL REFERENCES movies(id),
    auditorium_id UUID NOT NULL REFERENCES auditoriums(id),
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    price_cents INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (ends_at > starts_at),
    CHECK (price_cents >= 0)
);