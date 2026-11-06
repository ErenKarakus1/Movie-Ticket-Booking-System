package validations

import (
	"errors"
	"time"
	"unicode/utf8"

	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/movie-service/internal/models"
	"github.com/google/uuid"
)

func ValidateAddMovieRequest(req models.AddMovieRequest) error {
	lenTitle := utf8.RuneCountInString(req.Title)
	lenDescription := utf8.RuneCountInString(req.Description)

	if req.Title == "" {
		return errors.New("title is required")
	}
	if lenTitle > 128 {
		return errors.New("title must be at most 128 characters")
	}

	if req.Description == "" {
		return errors.New("description is required")
	}
	if lenDescription < 3 {
		return errors.New("description must be at least 3 characters")
	}
	if lenDescription > 512 {
		return errors.New("description must be at most 512 characters")
	}

	if req.DurationMinutes <= 0 {
		return errors.New("duration must be greater than zero minutes")
	}

	if req.ReleaseDate == "" {
		return errors.New("release date is required")
	}

	if _, err := time.Parse("2006-01-02", req.ReleaseDate); err != nil {
		return errors.New("release date must use YYYY-MM-DD format")
	}

	return nil
}

func ValidateCreateAuditoriumRequest(r models.CreateAuditoriumRequest) error {
	if r.Name == "" {
		return errors.New("auditorium name is required")
	}
	nameLength := utf8.RuneCountInString(r.Name)
	if nameLength < 3 {
		return errors.New("auditorium name must be at least 3 characters")
	}
	if nameLength > 64 {
		return errors.New("auditorium name must be at most 64 characters")
	}
	return nil
}

func ValidateAddShowtimeRequest(req models.AddShowtimeRequest) error {
	if req.PriceCents < 0 {
		return errors.New("price must be zero or greater, in cents")
	}
	if req.AuditoriumID == uuid.Nil {
		return errors.New("auditorium id is required")
	}
	if req.StartsAt.IsZero() {
		return errors.New("start time is required")
	}
	if req.StartsAt.Before(time.Now()) {
		return errors.New("showtime must start in the future")
	}
	return nil
}

type seatKey struct {
	rowLabel   string
	seatNumber int
}

func ValidateCreateSeatsRequest(req []models.CreateSeatRequest) error {
	if len(req) == 0 {
		return errors.New("at least one seat is required")
	}
	if len(req) > 500 {
		return errors.New("you can create at most 500 seats per request")
	}

	seen := make(map[seatKey]struct{}, len(req))
	for _, seat := range req {
		if seat.RowLabel == "" {
			return errors.New("all seats must have a row label")
		}
		if seat.SeatNumber <= 0 {
			return errors.New("each seat must have a seat number greater than zero")
		}
		charLen := utf8.RuneCountInString(seat.RowLabel)
		if charLen > 128 {
			return errors.New("row labels must be 128 characters or fewer")
		}
		key := seatKey{
			rowLabel:   seat.RowLabel,
			seatNumber: seat.SeatNumber,
		}
		if _, exists := seen[key]; exists {
			return errors.New("the request contains the same seat more than once")
		}
		seen[key] = struct{}{}
	}
	return nil
}
