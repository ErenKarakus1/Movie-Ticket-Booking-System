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
