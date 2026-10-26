package validations

import (
	"errors"

	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/ticket-service/internal/models"
	"github.com/google/uuid"
)

func ValidateCreateHoldRequest(req models.CreateHoldRequest) error {
	if len(req.SeatIDs) == 0 {
		return errors.New("at least 1 seat id is required")
	}

	if len(req.SeatIDs) > 500 {
		return errors.New("at most 500 seat ids at a time is allowed")
	}

	seen := make(map[uuid.UUID]struct{}, len(req.SeatIDs))
	for _, id := range req.SeatIDs {
		if _, exists := seen[id]; exists {
			return errors.New("duplicate seat id in request")
		}
		seen[id] = struct{}{}
	}

	return nil
}
