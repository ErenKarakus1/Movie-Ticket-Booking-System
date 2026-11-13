package validations

import (
	"strings"
	"testing"
	"time"

	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/movie-service/internal/models"
	"github.com/google/uuid"
)

func TestValidateAddMovieRequestRejectsInvalidInput(t *testing.T) {
	tests := []models.AddMovieRequest{
		{Description: "A description", DurationMinutes: 90, ReleaseDate: "2020-01-01"},
		{Title: "Film", Description: "x", DurationMinutes: 90, ReleaseDate: "2020-01-01"},
		{Title: "Film", Description: "A description", DurationMinutes: 0, ReleaseDate: "2020-01-01"},
		{Title: "Film", Description: "A description", DurationMinutes: 90, ReleaseDate: "2020-02-30"},
	}
	for _, req := range tests {
		if err := ValidateAddMovieRequest(req); err == nil {
			t.Fatalf("expected validation error for %#v", req)
		}
	}
}

func TestValidateAddShowtimeRequestRejectsInvalidInput(t *testing.T) {
	validFuture := time.Now().Add(time.Hour)
	tests := []models.AddShowtimeRequest{
		{AuditoriumID: uuid.New(), StartsAt: validFuture, PriceCents: -1},
		{StartsAt: validFuture, PriceCents: 100},
		{AuditoriumID: uuid.New(), StartsAt: time.Now().Add(-time.Hour), PriceCents: 100},
	}
	for _, req := range tests {
		if err := ValidateAddShowtimeRequest(req); err == nil {
			t.Fatalf("expected validation error for %#v", req)
		}
	}
}

func TestValidateCreateSeatsRejectsEmptyAndDuplicateSeats(t *testing.T) {
	if err := ValidateCreateSeatsRequest(nil); err == nil {
		t.Fatal("expected empty batch to be rejected")
	}
	duplicate := []models.CreateSeatRequest{{RowLabel: "A", SeatNumber: 1}, {RowLabel: "A", SeatNumber: 1}}
	if err := ValidateCreateSeatsRequest(duplicate); err == nil {
		t.Fatal("expected duplicate seat to be rejected")
	}
}

func TestValidateAddMovieRequestAcceptsValidRequest(t *testing.T) {
	req := models.AddMovieRequest{
		Title:           "Example Film",
		Description:     "A valid movie description",
		DurationMinutes: 120,
		ReleaseDate:     "2024-02-29",
	}
	if err := ValidateAddMovieRequest(req); err != nil {
		t.Fatalf("expected valid movie request, got error: %v", err)
	}
}

func TestValidateCreateSeatsRequestAcceptsMaximumBatch(t *testing.T) {
	seats := make([]models.CreateSeatRequest, 500)
	for i := range seats {
		seats[i] = models.CreateSeatRequest{RowLabel: "A", SeatNumber: i + 1}
	}
	if err := ValidateCreateSeatsRequest(seats); err != nil {
		t.Fatalf("expected 500-seat batch to be accepted, got error: %v", err)
	}
}

func TestValidateCreateSeatsRequestRejectsBatchOverMaximum(t *testing.T) {
	seats := make([]models.CreateSeatRequest, 501)
	for i := range seats {
		seats[i] = models.CreateSeatRequest{RowLabel: "A", SeatNumber: i + 1}
	}
	if err := ValidateCreateSeatsRequest(seats); err == nil {
		t.Fatal("expected batch larger than 500 seats to be rejected")
	}
}

func TestValidateAddShowtimeRequestAcceptsZeroPrice(t *testing.T) {
	req := models.AddShowtimeRequest{
		AuditoriumID: uuid.New(),
		StartsAt:     time.Now().Add(time.Hour),
		PriceCents:   0,
	}
	if err := ValidateAddShowtimeRequest(req); err != nil {
		t.Fatalf("expected zero-price showtime to be accepted, got error: %v", err)
	}
}

func TestValidateCreateAuditoriumRequestBoundaries(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "minimum length", value: "ABC"},
		{name: "below minimum", value: "AB", wantErr: true},
		{name: "maximum length", value: strings.Repeat("A", 64)},
		{name: "above maximum", value: strings.Repeat("A", 65), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCreateAuditoriumRequest(models.CreateAuditoriumRequest{Name: tt.value})
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateCreateAuditoriumRequest() error = %v, wantErr %t", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCreateSeatsRejectsInvalidSeatFields(t *testing.T) {
	tests := []struct {
		name string
		seat models.CreateSeatRequest
	}{
		{name: "missing row label", seat: models.CreateSeatRequest{SeatNumber: 1}},
		{name: "zero seat number", seat: models.CreateSeatRequest{RowLabel: "A", SeatNumber: 0}},
		{name: "negative seat number", seat: models.CreateSeatRequest{RowLabel: "A", SeatNumber: -1}},
		{name: "row label over maximum", seat: models.CreateSeatRequest{RowLabel: strings.Repeat("A", 129), SeatNumber: 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateCreateSeatsRequest([]models.CreateSeatRequest{tt.seat}); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
