package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Movie struct {
	ID              uuid.UUID `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	DurationMinutes int       `json:"duration_minutes"`
	ReleaseDate     time.Time `json:"release_date"`
	CreatedAt       time.Time `json:"created_at"`
}

type AddMovieRequest struct {
	Title           string `json:"title"`
	Description     string `json:"description"`
	DurationMinutes int    `json:"duration_minutes"`
	ReleaseDate     string `json:"release_date"`
}

func (r *AddMovieRequest) Normalize() {
	r.Title = strings.TrimSpace(r.Title)
	r.Description = strings.TrimSpace(r.Description)
}

type Auditorium struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateAuditoriumRequest struct {
	Name string `json:"name"`
}

func (r *CreateAuditoriumRequest) Normalize() {
	r.Name = strings.TrimSpace(r.Name)
}

type Showtime struct {
	ID           uuid.UUID `json:"id"`
	MovieID      uuid.UUID `json:"movie_id"`
	AuditoriumID uuid.UUID `json:"auditorium_id"`
	StartsAt     time.Time `json:"starts_at"`
	EndsAt       time.Time `json:"ends_at"`
	PriceCents   int       `json:"price_cents"`
	CreatedAt    time.Time `json:"created_at"`
}

type AddShowtimeRequest struct {
	AuditoriumID uuid.UUID `json:"auditorium_id"`
	StartsAt     time.Time `json:"starts_at"`
	PriceCents   int       `json:"price_cents"`
}
