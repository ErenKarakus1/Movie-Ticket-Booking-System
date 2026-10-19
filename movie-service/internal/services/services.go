package services

import (
	"context"
	"errors"
	"time"

	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/movie-service/internal/models"
	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/movie-service/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func generateMovieID(req models.AddMovieRequest) (models.Movie, error) {
	date, err := time.Parse("2006-01-02", req.ReleaseDate)
	if err != nil {
		return models.Movie{}, errors.New("internal server error")
	}
	return models.Movie{
		ID:              uuid.New(),
		Title:           req.Title,
		Description:     req.Description,
		DurationMinutes: req.DurationMinutes,
		ReleaseDate:     date,
	}, nil
}

func CreateMovie(ctx context.Context, pool *pgxpool.Pool, req models.AddMovieRequest) (models.Movie, error) {
	movie, err := generateMovieID(req)
	if err != nil {
		return models.Movie{}, errors.New("internal server error")
	}
	insertedMovie, err := repository.InsertMovie(ctx, pool, movie)
	if err != nil {
		return models.Movie{}, errors.New("internal server error")
	}
	return insertedMovie, nil
}

func GetMovie(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (models.Movie, error) {
	movie, err := repository.GetMovieByID(ctx, pool, id)
	if err != nil {
		if errors.Is(err, repository.ErrMovieNotFound) {
			return models.Movie{}, err
		}
		return models.Movie{}, errors.New("internal server error")
	}

	return movie, nil
}
