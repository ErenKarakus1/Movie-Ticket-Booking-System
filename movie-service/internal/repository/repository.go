package repository

import (
	"context"
	"errors"

	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/movie-service/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrMovieNotFound = errors.New("movie not found")
var ErrAuditoriumNotFound = errors.New("auditorium not found")
var ErrOverlap = errors.New("showtime overlaps with another screening at the auditorium")
var ErrSeatExists = errors.New("seat already exists in auditorium")

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

const updateMovieSQL = `
	UPDATE movies
	SET 
		title=$1,
		description=$2,
		duration_minutes=$3,
		release_date=$4
	WHERE id=$5
	RETURNING
		id,
		title,
		description,
		duration_minutes,
		release_date,
		created_at
`

const insertAuditoriumSQL = `
	INSERT INTO auditoriums (
		id,
		name
	)
	VALUES ($1,$2)
	RETURNING
		id,
		name,
		created_at
`

const getAuditoriumByIDSQL = `
	SELECT
		id,
		name,
		created_at
	FROM auditoriums
	WHERE id=$1
`

const checkOverlapSQL = `
	SELECT EXISTS (
		SELECT 1
		FROM showtimes
		WHERE auditorium_id=$1
		AND starts_at < $2  
		AND ends_at > $3
	)
`

const insertShowtimeSQL = `
	INSERT INTO showtimes (
		id,
		movie_id,
		auditorium_id,
		starts_at,
		ends_at,
		price_cents
	)
	VALUES ($1,$2,$3,$4,$5,$6)
	RETURNING
		id,
		movie_id,
		auditorium_id,
		starts_at,
		ends_at,
		price_cents,
		created_at
`

const lockAuditoriumRowSQL = `
	SELECT id
	FROM auditoriums
	WHERE id=$1
	FOR UPDATE
`

const getMovieShowtimesSQL = `
	SELECT
		id,
		movie_id,
		auditorium_id,
		starts_at,
		ends_at,
		price_cents,
		created_at
	FROM showtimes
	WHERE movie_id=$1
	AND starts_at > NOW()
	ORDER BY starts_at ASC
`

const insertSeatSQL = `
	INSERT INTO seats (
		id,
		auditorium_id,
		row_label,
		seat_number
	)
	VALUES ($1,$2,$3,$4)
	RETURNING
		id,
		auditorium_id,
		row_label,
		seat_number,
		created_at
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

func UpdateMovie(ctx context.Context, pool *pgxpool.Pool, req models.AddMovieRequest, id uuid.UUID) (models.Movie, error) {
	var movie models.Movie
	err := pool.QueryRow(
		ctx,
		updateMovieSQL,
		req.Title,
		req.Description,
		req.DurationMinutes,
		req.ReleaseDate,
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

func InsertAuditorium(ctx context.Context, pool *pgxpool.Pool, auditorium models.Auditorium) (models.Auditorium, error) {
	var insertedAuditorium models.Auditorium
	err := pool.QueryRow(
		ctx,
		insertAuditoriumSQL,
		auditorium.ID,
		auditorium.Name,
	).Scan(
		&insertedAuditorium.ID,
		&insertedAuditorium.Name,
		&insertedAuditorium.CreatedAt,
	)
	if err != nil {
		return models.Auditorium{}, errors.New("internal server error")
	}
	return insertedAuditorium, nil
}

func GetAuditoriumByID(ctx context.Context, pool *pgxpool.Pool, auditoriumID uuid.UUID) (models.Auditorium, error) {
	var auditorium models.Auditorium
	err := pool.QueryRow(
		ctx,
		getAuditoriumByIDSQL,
		auditoriumID,
	).Scan(
		&auditorium.ID,
		&auditorium.Name,
		&auditorium.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Auditorium{}, ErrAuditoriumNotFound
		}
		return models.Auditorium{}, errors.New("internal server error")
	}
	return auditorium, nil
}

func InsertShowtime(ctx context.Context, pool *pgxpool.Pool, showtime models.Showtime) (models.Showtime, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return models.Showtime{}, errors.New("internal server error")
	}
	defer tx.Rollback(ctx)

	var auditoriumID uuid.UUID
	err = tx.QueryRow(
		ctx,
		lockAuditoriumRowSQL,
		showtime.AuditoriumID,
	).Scan(&auditoriumID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Showtime{}, ErrAuditoriumNotFound
		}
		return models.Showtime{}, errors.New("internal server error")
	}

	var overlaps bool
	err = tx.QueryRow(
		ctx,
		checkOverlapSQL,
		showtime.AuditoriumID,
		showtime.EndsAt,
		showtime.StartsAt,
	).Scan(&overlaps)
	if err != nil {
		return models.Showtime{}, errors.New("internal server error")
	}
	if overlaps {
		return models.Showtime{}, ErrOverlap
	}
	var insertedShowtime models.Showtime
	err = tx.QueryRow(
		ctx,
		insertShowtimeSQL,
		showtime.ID,
		showtime.MovieID,
		showtime.AuditoriumID,
		showtime.StartsAt,
		showtime.EndsAt,
		showtime.PriceCents,
	).Scan(
		&insertedShowtime.ID,
		&insertedShowtime.MovieID,
		&insertedShowtime.AuditoriumID,
		&insertedShowtime.StartsAt,
		&insertedShowtime.EndsAt,
		&insertedShowtime.PriceCents,
		&insertedShowtime.CreatedAt,
	)
	if err != nil {
		return models.Showtime{}, errors.New("internal server error")
	}
	err = tx.Commit(ctx)
	if err != nil {
		return models.Showtime{}, errors.New("internal server error")
	}
	return insertedShowtime, nil
}

func GetMovieShowtimes(ctx context.Context, pool *pgxpool.Pool, movieID uuid.UUID) ([]models.Showtime, error) {
	rows, err := pool.Query(
		ctx,
		getMovieShowtimesSQL,
		movieID,
	)
	if err != nil {
		return []models.Showtime{}, errors.New("internal server error")
	}
	defer rows.Close()

	var showtimes []models.Showtime
	for rows.Next() {
		var showtime models.Showtime
		err := rows.Scan(
			&showtime.ID,
			&showtime.MovieID,
			&showtime.AuditoriumID,
			&showtime.StartsAt,
			&showtime.EndsAt,
			&showtime.PriceCents,
			&showtime.CreatedAt,
		)
		if err != nil {
			return []models.Showtime{}, errors.New("internal server error")
		}

		showtimes = append(showtimes, showtime)
	}

	if err := rows.Err(); err != nil {
		return []models.Showtime{}, errors.New("internal server error")
	}

	return showtimes, nil
}

func InsertSeats(ctx context.Context, pool *pgxpool.Pool, seats []models.Seat) ([]models.Seat, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return []models.Seat{}, errors.New("internal server error")
	}
	defer tx.Rollback(ctx)

	var createdSeats []models.Seat

	for _, seat := range seats {
		var createdSeat models.Seat
		err := tx.QueryRow(
			ctx,
			insertSeatSQL,
			seat.ID,
			seat.AuditoriumID,
			seat.RowLabel,
			seat.SeatNumber,
		).Scan(
			&createdSeat.ID,
			&createdSeat.AuditoriumID,
			&createdSeat.RowLabel,
			&createdSeat.SeatNumber,
			&createdSeat.CreatedAt,
		)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return []models.Seat{}, ErrSeatExists
			}
			return []models.Seat{}, errors.New("internal server error")
		}
		createdSeats = append(createdSeats, createdSeat)
	}

	if err := tx.Commit(ctx); err != nil {
		return []models.Seat{}, errors.New("internal server error")
	}

	return createdSeats, nil
}
