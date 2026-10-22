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

func GetAllMovies(ctx context.Context, pool *pgxpool.Pool) ([]models.Movie, error) {
	movies, err := repository.GetAllMovies(ctx, pool)
	if err != nil {
		return []models.Movie{}, errors.New("internal server error")
	}
	return movies, nil
}

func UpdateMovie(ctx context.Context, pool *pgxpool.Pool, req models.AddMovieRequest, id uuid.UUID) (models.Movie, error) {
	updatedMovie, err := repository.UpdateMovie(ctx, pool, req, id)
	if err != nil {
		if errors.Is(err, repository.ErrMovieNotFound) {
			return models.Movie{}, err
		}
		return models.Movie{}, errors.New("internal server error")
	}

	return updatedMovie, nil
}

func generateAuditoriumID(req models.CreateAuditoriumRequest) models.Auditorium {
	return models.Auditorium{
		ID:   uuid.New(),
		Name: req.Name,
	}
}

func CreateAuditorium(ctx context.Context, pool *pgxpool.Pool, req models.CreateAuditoriumRequest) (models.Auditorium, error) {
	auditorium := generateAuditoriumID(req)
	insertedAuditorium, err := repository.InsertAuditorium(ctx, pool, auditorium)
	if err != nil {
		return models.Auditorium{}, errors.New("internal server error")
	}
	return insertedAuditorium, nil
}

func generateShowtime(req models.AddShowtimeRequest, movie models.Movie) models.Showtime {
	return models.Showtime{
		ID:           uuid.New(),
		MovieID:      movie.ID,
		AuditoriumID: req.AuditoriumID,
		StartsAt:     req.StartsAt,
		EndsAt:       req.StartsAt.Add(time.Duration(movie.DurationMinutes) * time.Minute),
		PriceCents:   req.PriceCents,
	}
}

func AddShowtime(ctx context.Context, pool *pgxpool.Pool, req models.AddShowtimeRequest, movieID uuid.UUID) (models.Showtime, error) {
	movie, err := repository.GetMovieByID(ctx, pool, movieID)
	if err != nil {
		if errors.Is(err, repository.ErrMovieNotFound) {
			return models.Showtime{}, err
		}
		return models.Showtime{}, errors.New("internal server error")
	}
	_, err = repository.GetAuditoriumByID(ctx, pool, req.AuditoriumID)
	if err != nil {
		if errors.Is(err, repository.ErrAuditoriumNotFound) {
			return models.Showtime{}, err
		}
		return models.Showtime{}, errors.New("internal server error")
	}

	showtime := generateShowtime(req, movie)
	insertedShowtime, err := repository.InsertShowtime(ctx, pool, showtime)
	if err != nil {
		if errors.Is(err, repository.ErrOverlap) || errors.Is(err, repository.ErrAuditoriumNotFound) {
			return models.Showtime{}, err
		}
		return models.Showtime{}, errors.New("internal server error")
	}

	return insertedShowtime, nil
}
