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
