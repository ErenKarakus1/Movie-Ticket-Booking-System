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

func TestValidateCreateHoldRequestRejectsEmptyNonNilSlice(t *testing.T) {
	req := models.CreateHoldRequest{SeatIDs: []uuid.UUID{}}
	if err := ValidateCreateHoldRequest(req); err == nil {
		t.Fatal("expected validation error for empty seat ID list")
	}
}

func TestValidateCreateHoldRequestRejectsNonAdjacentDuplicateSeat(t *testing.T) {
	a, b, c := uuid.New(), uuid.New(), uuid.New()
	tests := []struct {
		name string
		ids  []uuid.UUID
	}{
		{"one apart", []uuid.UUID{a, b, a}},
		{"first and last", []uuid.UUID{a, b, c, a}},
		{"duplicate at end", []uuid.UUID{a, b, c, c}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := models.CreateHoldRequest{SeatIDs: tt.ids}
			if err := ValidateCreateHoldRequest(req); err == nil {
				t.Fatal("expected validation error for duplicate seat")
			}
		})
	}
}

func TestValidateCreateHoldRequestAcceptsSingleSeat(t *testing.T) {
	req := models.CreateHoldRequest{SeatIDs: []uuid.UUID{uuid.New()}}
	if err := ValidateCreateHoldRequest(req); err != nil {
		t.Fatalf("expected single-seat hold request to be accepted, got error: %v", err)
	}
}

func TestValidateCreateHoldRequestAcceptsMaximumBatch(t *testing.T) {
	seatIDs := make([]uuid.UUID, 500)
	for i := range seatIDs {
		seatIDs[i] = uuid.New()
	}
	if err := ValidateCreateHoldRequest(models.CreateHoldRequest{SeatIDs: seatIDs}); err != nil {
		t.Fatalf("expected 500-seat hold request to be accepted, got error: %v", err)
	}
}

func TestValidateCreateHoldRequestRejectsBatchOverMaximum(t *testing.T) {
	seatIDs := make([]uuid.UUID, 501)
	for i := range seatIDs {
		seatIDs[i] = uuid.New()
	}
	if err := ValidateCreateHoldRequest(models.CreateHoldRequest{SeatIDs: seatIDs}); err == nil {
		t.Fatal("expected hold request over 500 seats to be rejected")
	}
}
