package validations

import (
	"testing"

	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/ticket-service/internal/models"
	"github.com/google/uuid"
)

func TestValidateCreateHoldRequestRejectsInvalidBatches(t *testing.T) {
	valid := uuid.New()
	tests := []struct {
		name string
		req  models.CreateHoldRequest
	}{
		{"empty", models.CreateHoldRequest{}},
		{"duplicate seat", models.CreateHoldRequest{SeatIDs: []uuid.UUID{valid, valid}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateCreateHoldRequest(tt.req); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
