package validations

import (
	"errors"
	"time"
	"unicode/utf8"

	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/movie-service/internal/models"
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
		return errors.New("duration minutes is required")
	}

	if req.ReleaseDate == "" {
		return errors.New("release date is required")
	}

	if _, err := time.Parse("2006-01-02", req.ReleaseDate); err != nil {
		return errors.New("invalid release date format")
	}

	return nil
}
