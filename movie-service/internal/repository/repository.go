package repository

import (
	"context"
	"errors"

	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/movie-service/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrMovieNotFound = errors.New("movie not found")

const insertMovieSQL = `
	INSERT INTO movies (
		id,
		title,
		description,
		duration_minutes,
		release_date
	)
	VALUES ($1,$2,$3,$4,$5)
	RETURNING
		id,
		title,
		description,
		duration_minutes,
		release_date,
		created_at
`

const getMovieByIDSQL = `
	SELECT
		id,
		title,
		description,
		duration_minutes,
		release_date,
		created_at
	FROM movies
	WHERE id=$1
`

const getAllMoviesSQL = `
	SELECT
		id,
		title,
		description,
		duration_minutes,
		release_date,
		created_at
	FROM movies
	ORDER BY created_at DESC
`

func InsertMovie(ctx context.Context, pool *pgxpool.Pool, movie models.Movie) (models.Movie, error) {
	var insertedMovie models.Movie
	err := pool.QueryRow(
		ctx,
		insertMovieSQL,
		movie.ID,
		movie.Title,
		movie.Description,
		movie.DurationMinutes,
		movie.ReleaseDate,
	).Scan(
		&insertedMovie.ID,
		&insertedMovie.Title,
		&insertedMovie.Description,
		&insertedMovie.DurationMinutes,
		&insertedMovie.ReleaseDate,
		&insertedMovie.CreatedAt,
	)

	if err != nil {
		return models.Movie{}, errors.New("internal server error")
	}

	return insertedMovie, nil
}

func GetMovieByID(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (models.Movie, error) {
	var movie models.Movie
	err := pool.QueryRow(
		ctx,
		getMovieByIDSQL,
		id,
	).Scan(
		&movie.ID,
		&movie.Title,
		&movie.Description,
		&movie.DurationMinutes,
		&movie.ReleaseDate,
		&movie.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Movie{}, ErrMovieNotFound
		}
		return models.Movie{}, errors.New("internal server error")
	}

	return movie, nil
}

func GetAllMovies(ctx context.Context, pool *pgxpool.Pool) ([]models.Movie, error) {
	rows, err := pool.Query(
		ctx,
		getAllMoviesSQL,
	)
	if err != nil {
		return []models.Movie{}, errors.New("internal server error")
	}
	defer rows.Close()

	var movies []models.Movie
	for rows.Next() {
		var movie models.Movie
		err := rows.Scan(
			&movie.ID,
			&movie.Title,
			&movie.Description,
			&movie.DurationMinutes,
			&movie.ReleaseDate,
			&movie.CreatedAt,
		)
		if err != nil {
			return []models.Movie{}, errors.New("internal server error")
		}
		movies = append(movies, movie)
	}

	if err := rows.Err(); err != nil {
		return []models.Movie{}, errors.New("internal server error")
	}
	return movies, nil
}
