package repository

import (
	"context"
	"errors"

	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/ticket-service/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrSeatAlreadyBooked = errors.New("seat already booked")

const insertBookingSQL = `
	INSERT INTO bookings (
		id,
		hold_id,
		user_id,
		showtime_id,
		status
	)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (hold_id) DO NOTHING
	RETURNING created_at
`

const insertBookingSeatSQL = `
	INSERT INTO booking_seats (
		booking_id,
		showtime_id,
		seat_id,
		price_cents
	)
	VALUES ($1, $2, $3, $4)
`

const getBookingByHoldIDSQL = `
	SELECT 
		id, 
		hold_id, 
		user_id, 
		showtime_id, 
		status, 
		created_at
	FROM bookings
	WHERE hold_id = $1
`

func GetBookingByHoldID(ctx context.Context, pool *pgxpool.Pool, holdID uuid.UUID) (models.Booking, bool, error) {
	var booking models.Booking
	err := pool.QueryRow(ctx, getBookingByHoldIDSQL, holdID).Scan(
		&booking.ID,
		&booking.HoldID,
		&booking.UserID,
		&booking.ShowtimeID,
		&booking.Status,
		&booking.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Booking{}, false, nil
	}
	if err != nil {
		return models.Booking{}, false, errors.New("internal server error")
	}
	return booking, true, nil
}

func InsertBooking(ctx context.Context, pool *pgxpool.Pool, booking *models.Booking, seats []models.BookingSeat) (bool, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return false, errors.New("internal server error")
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, insertBookingSQL,
		booking.ID,
		booking.HoldID,
		booking.UserID,
		booking.ShowtimeID,
		booking.Status,
	).Scan(&booking.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		err = tx.QueryRow(ctx, getBookingByHoldIDSQL, booking.HoldID).Scan(
			&booking.ID,
			&booking.HoldID,
			&booking.UserID,
			&booking.ShowtimeID,
			&booking.Status,
			&booking.CreatedAt,
		)
		if err != nil {
			return false, errors.New("internal server error")
		}
		if err := tx.Commit(ctx); err != nil {
			return false, errors.New("internal server error")
		}
		return false, nil
	}
	if err != nil {
		return false, errors.New("internal server error")
	}

	for _, seat := range seats {
		_, err = tx.Exec(ctx, insertBookingSeatSQL,
			seat.BookingID,
			seat.ShowtimeID,
			seat.SeatID,
			seat.PriceCents,
		)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.ConstraintName == "booking_seats_active_showtime_seat_unique" {
				return false, ErrSeatAlreadyBooked
			}
			return false, errors.New("internal server error")
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return false, errors.New("internal server error")
	}
	return true, nil
}
