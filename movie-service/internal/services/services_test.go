package services

import (
	"testing"
	"time"

	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/movie-service/internal/models"
	"github.com/google/uuid"
)

func TestGenerateMovieIDParsesReleaseDateAndCreatesID(t *testing.T) {
	req := models.AddMovieRequest{
		Title:           "Example",
		Description:     "A movie description",
		DurationMinutes: 105,
		ReleaseDate:     "2024-02-29",
	}
	movie, err := generateMovieID(req)
	if err != nil {
		t.Fatalf("generateMovieID() error = %v", err)
	}
	if movie.ID == uuid.Nil {
		t.Fatal("expected generated movie ID")
	}
	if got := movie.ReleaseDate.Format("2006-01-02"); got != req.ReleaseDate {
		t.Fatalf("release date = %q, want %q", got, req.ReleaseDate)
	}

	if _, err := generateMovieID(models.AddMovieRequest{ReleaseDate: "2024-02-30"}); err == nil {
		t.Fatal("expected invalid calendar date to fail")
	}
}

func TestGenerateShowtimeCalculatesEndTime(t *testing.T) {
	startsAt := time.Date(2030, 3, 4, 18, 30, 0, 0, time.UTC)
	movie := models.Movie{ID: uuid.New(), DurationMinutes: 123}
	auditoriumID := uuid.New()
	req := models.AddShowtimeRequest{AuditoriumID: auditoriumID, StartsAt: startsAt, PriceCents: 1500}

	showtime := generateShowtime(req, movie)
	if showtime.ID == uuid.Nil {
		t.Fatal("expected generated showtime ID")
	}
	if showtime.MovieID != movie.ID || showtime.AuditoriumID != auditoriumID {
		t.Fatalf("unexpected showtime relationships: %#v", showtime)
	}
	if !showtime.StartsAt.Equal(startsAt) || !showtime.EndsAt.Equal(startsAt.Add(123*time.Minute)) {
		t.Fatalf("unexpected showtime interval: starts=%v ends=%v", showtime.StartsAt, showtime.EndsAt)
	}
	if showtime.PriceCents != req.PriceCents {
		t.Fatalf("price cents = %d, want %d", showtime.PriceCents, req.PriceCents)
	}
}

func TestGenerateSeatAssignsIDAndAuditorium(t *testing.T) {
	auditoriumID := uuid.New()
	seat := generateSeatWithID(auditoriumID, models.CreateSeatRequest{RowLabel: "B", SeatNumber: 12})
	if seat.ID == uuid.Nil {
		t.Fatal("expected generated seat ID")
	}
	if seat.AuditoriumID != auditoriumID || seat.RowLabel != "B" || seat.SeatNumber != 12 {
		t.Fatalf("unexpected generated seat: %#v", seat)
	}
}
