CREATE TABLE IF NOT EXISTS seats (
    id UUID PRIMARY KEY,
    auditorium_id UUID NOT NULL REFERENCES auditoriums(id),
    row_label TEXT NOT NULL,
    seat_number INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (auditorium_id, row_label, seat_number)
);