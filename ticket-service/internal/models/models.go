package models

import (
	"time"

	"github.com/google/uuid"
)

type CreateHoldRequest struct {
	SeatIDs []uuid.UUID `json:"seat_ids"`
}

type HoldDetails struct {
	HoldID     uuid.UUID   `json:"hold_id"`
	ShowtimeID uuid.UUID   `json:"showtime_id"`
	SeatIDs    []uuid.UUID `json:"seat_ids"`
	ExpiresAt  time.Time   `json:"expires_at"`
}

type Booking struct {
	ID         uuid.UUID `json:"id"`
	HoldID     uuid.UUID `json:"hold_id"`
	UserID     uuid.UUID `json:"user_id"`
	ShowtimeID uuid.UUID `json:"showtime_id"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type BookingSeat struct {
	BookingID  uuid.UUID `json:"booking_id"`
	ShowtimeID uuid.UUID `json:"showtime_id"`
	SeatID     uuid.UUID `json:"seat_id"`
	PriceCents int       `json:"price_cents"`
}
