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
var ErrBookingNotFound = errors.New("booking not found")
var ErrBookingAlreadyChecked = errors.New("booking already checked")

const getBookedSeatIDsByShowtimeSQL = `
	SELECT seat_id
	FROM booking_seats
	WHERE showtime_id = $1
`

const insertBookingSQL = `
	INSERT INTO bookings (
		id,
		hold_id,
		user_id,
		showtime_id
	)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (hold_id) DO NOTHING
	RETURNING status, created_at
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

const getBookingByIDSQL = `
	SELECT 
		id, 
		showtime_id, 
		status,
		created_at
	FROM bookings
	WHERE id = $1
	AND user_id=$2
`

const getBookingSeatsByIDSQL = `
	SELECT
		seat_id,
		price_cents
	FROM booking_seats
	WHERE booking_id=$1
`

const getAllBookingsByUserIDSQL = `
	SELECT 
		id, 
		showtime_id, 
		status,
		created_at
	FROM bookings
	WHERE user_id=$1
	ORDER BY created_at DESC
`

const redeemBookingSQL = `
	UPDATE bookings
	SET status = 'checked'
	WHERE id = $1 AND status = 'not_checked'
	RETURNING id
`

const getBookingStatusSQL = `
	SELECT status
	FROM bookings
	WHERE id = $1
`

func GetBookedSeatIDsByShowtime(ctx context.Context, pool *pgxpool.Pool, showtimeID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := pool.Query(ctx, getBookedSeatIDsByShowtimeSQL, showtimeID)
	if err != nil {
		return nil, errors.New("internal server error")
	}
	defer rows.Close()

	var seatIDs []uuid.UUID
	for rows.Next() {
		var seatID uuid.UUID
		if err := rows.Scan(&seatID); err != nil {
			return nil, errors.New("internal server error")
		}
		seatIDs = append(seatIDs, seatID)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.New("internal server error")
	}
	return seatIDs, nil
}

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
	).Scan(&booking.Status, &booking.CreatedAt)
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
			if errors.As(err, &pgErr) && pgErr.ConstraintName == "booking_seats_showtime_id_seat_id_key" {
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

func GetBooking(ctx context.Context, pool *pgxpool.Pool, bookingID uuid.UUID, userID uuid.UUID) (models.GetBookingResponse, error) {
	var resp models.GetBookingResponse
	err := pool.QueryRow(
		ctx,
		getBookingByIDSQL,
		bookingID,
		userID,
	).Scan(
		&resp.ID,
		&resp.ShowtimeID,
		&resp.Status,
		&resp.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.GetBookingResponse{}, ErrBookingNotFound
		}
		return models.GetBookingResponse{}, errors.New("internal server error")
	}

	var seats []models.GetBookingSeatResponse
	rows, err := pool.Query(
		ctx,
		getBookingSeatsByIDSQL,
		bookingID,
	)
	if err != nil {
		return models.GetBookingResponse{}, errors.New("internal server error")
	}
	defer rows.Close()

	for rows.Next() {
		var seat models.GetBookingSeatResponse
		err := rows.Scan(
			&seat.SeatID,
			&seat.PriceCents,
		)
		if err != nil {
			return models.GetBookingResponse{}, errors.New("internal server error")
		}
		seats = append(seats, seat)
	}

	if err := rows.Err(); err != nil {
		return models.GetBookingResponse{}, errors.New("internal server error")
	}

	resp.Seats = seats

	return resp, nil
}

func GetAllBookings(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) ([]models.GetBookingResponse, error) {
	bookingRows, err := pool.Query(
		ctx,
		getAllBookingsByUserIDSQL,
		userID,
	)
	if err != nil {
		return []models.GetBookingResponse{}, errors.New("internal server error")
	}
	defer bookingRows.Close()

	var bookings []models.GetBookingResponse
	for bookingRows.Next() {
		var booking models.GetBookingResponse
		err := bookingRows.Scan(
			&booking.ID,
			&booking.ShowtimeID,
			&booking.Status,
			&booking.CreatedAt,
		)
		if err != nil {
			return []models.GetBookingResponse{}, errors.New("internal server error")
		}
		bookings = append(bookings, booking)
	}
	if err := bookingRows.Err(); err != nil {
		return []models.GetBookingResponse{}, errors.New("internal server error")
	}

	for idx, booking := range bookings {
		seatRows, err := pool.Query(
			ctx,
			getBookingSeatsByIDSQL,
			booking.ID,
		)
		if err != nil {
			return []models.GetBookingResponse{}, errors.New("internal server error")
		}

		var seats []models.GetBookingSeatResponse
		for seatRows.Next() {
			var seat models.GetBookingSeatResponse
			err := seatRows.Scan(
				&seat.SeatID,
				&seat.PriceCents,
			)
			if err != nil {
				seatRows.Close()
				return []models.GetBookingResponse{}, errors.New("internal server error")
			}
			seats = append(seats, seat)
		}
		if err := seatRows.Err(); err != nil {
			seatRows.Close()
			return []models.GetBookingResponse{}, errors.New("internal server error")
		}
		bookings[idx].Seats = seats
		seatRows.Close()
	}

	return bookings, nil
}

func RedeemBooking(ctx context.Context, pool *pgxpool.Pool, bookingID uuid.UUID) error {
	var updatedID uuid.UUID
	err := pool.QueryRow(ctx, redeemBookingSQL, bookingID).Scan(&updatedID)
	if err == nil {
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return errors.New("internal server error")
	}

	var status string
	err = pool.QueryRow(ctx, getBookingStatusSQL, bookingID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrBookingNotFound
	}
	if err != nil {
		return errors.New("internal server error")
	}
	if status == "checked" {
		return ErrBookingAlreadyChecked
	}
	return errors.New("internal server error")
}
