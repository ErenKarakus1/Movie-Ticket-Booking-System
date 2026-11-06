package validations

import (
	"errors"

	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/ticket-service/internal/models"
	"github.com/google/uuid"
)

func ValidateCreateHoldRequest(req models.CreateHoldRequest) error {
	if len(req.SeatIDs) == 0 {
		return errors.New("select at least one seat")
	}

	if len(req.SeatIDs) > 500 {
		return errors.New("you can request at most 500 seats at a time")
	}

	seen := make(map[uuid.UUID]struct{}, len(req.SeatIDs))
	for _, id := range req.SeatIDs {
		if _, exists := seen[id]; exists {
			return errors.New("the request contains the same seat more than once")
		}
		seen[id] = struct{}{}
	}

	return nil
}
