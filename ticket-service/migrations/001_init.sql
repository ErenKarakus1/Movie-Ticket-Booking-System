CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY,
    hold_id UUID NOT NULL UNIQUE,
    user_id UUID NOT NULL,
    showtime_id UUID NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (status IN ('confirmed', 'cancelled'))
);

CREATE TABLE IF NOT EXISTS booking_seats (
    booking_id UUID NOT NULL REFERENCES bookings(id),
    showtime_id UUID NOT NULL,
    seat_id UUID NOT NULL,
    price_cents INT NOT NULL,
    cancelled_at TIMESTAMPTZ,
    PRIMARY KEY (booking_id, seat_id),
    CHECK (price_cents >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS booking_seats_active_showtime_seat_unique
    ON booking_seats (showtime_id, seat_id)
    WHERE cancelled_at IS NULL;
