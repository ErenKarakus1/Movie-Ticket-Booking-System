package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/ticket-service/internal/models"
	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/ticket-service/internal/movieclient"
	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/ticket-service/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

var ErrSeatAlreadyHeld = errors.New("seat already held")
var ErrSeatNotInShowtime = errors.New("seat does not belong to showtime auditorium")
var ErrShowtimeStarted = errors.New("showtime has already started")
var ErrHoldExpiredOrMissing = errors.New("hold is missing or expired")
var ErrHoldDoesntBelongToUser = errors.New("hold belongs to another user")
var ErrInvalidHoldDetails = errors.New("invalid hold details")
var ErrSeatAlreadyBooked = repository.ErrSeatAlreadyBooked

func CreateHold(ctx context.Context, client *redis.Client, movies *movieclient.Client, showtimeID uuid.UUID, seatIDs []uuid.UUID, userID uuid.UUID) (models.HoldDetails, error) {
	showtime, err := movies.GetShowtime(ctx, showtimeID)
	if err != nil {
		return models.HoldDetails{}, err
	}

	showtimeSeats, err := movies.GetShowtimeSeats(ctx, showtimeID)
	if err != nil {
		return models.HoldDetails{}, err
	}
	validSeatIDs := make(map[uuid.UUID]struct{}, len(showtimeSeats))
	for _, seat := range showtimeSeats {
		validSeatIDs[seat.ID] = struct{}{}
	}
	for _, seatID := range seatIDs {
		if _, exists := validSeatIDs[seatID]; !exists {
			return models.HoldDetails{}, ErrSeatNotInShowtime
		}
	}

	now := time.Now()
	if !showtime.StartsAt.After(now) {
		return models.HoldDetails{}, ErrShowtimeStarted
	}

	const holdSeatsScript = `
			-- KEYS[1..N]: hold:<showtimeID>:<seatID>
			-- KEYS[N+1]:  hold:<holdID>
			-- ARGV[1]: holdID
			-- ARGV[2]: userID
			-- ARGV[3]: TTL in milliseconds

			local seatCount = #KEYS - 1
			local holdKey = KEYS[#KEYS]

			for i = 1, seatCount do
				if redis.call("EXISTS", KEYS[i]) == 1 then
					return 0
				end
			end

			for i = 1, seatCount do
				redis.call("SET", KEYS[i], ARGV[1], "PX", ARGV[3])
			end

			redis.call("HSET", holdKey,
				"user_id", ARGV[2],
				"hold_id", ARGV[1]
			)

			for i = 1, seatCount do
				redis.call("HSET", holdKey, "seat:" .. i, KEYS[i])
			end

			redis.call("PEXPIRE", holdKey, ARGV[3])

			return 1
		`
	holdID := uuid.New()
	ttl := 5 * time.Minute
	if untilShowtime := showtime.StartsAt.Sub(now); untilShowtime < ttl {
		ttl = untilShowtime
	}
	expiresAt := now.Add(ttl)
	var keys []string
	for _, seatID := range seatIDs {
		key := fmt.Sprintf("hold:%v:%v", showtimeID.String(), seatID.String())
		keys = append(keys, key)
	}
	keys = append(keys, "hold:"+holdID.String())

	result, err := client.Eval(ctx, holdSeatsScript, keys, holdID.String(), userID.String(), ttl.Milliseconds()).Int()

	if err != nil {
		return models.HoldDetails{}, errors.New("internal server error")
	}

	if result == 0 {
		return models.HoldDetails{}, ErrSeatAlreadyHeld
	}

	return models.HoldDetails{
		HoldID:     holdID,
		ShowtimeID: showtimeID,
		SeatIDs:    seatIDs,
		ExpiresAt:  expiresAt,
	}, nil
}

func DeleteHold(ctx context.Context, client *redis.Client, holdID uuid.UUID, userID uuid.UUID) error {
	const releaseHoldScript = `
		local holdKey = KEYS[1]

		if redis.call("EXISTS", holdKey) == 0 then
			return 0
		end

		if redis.call("HGET", holdKey, "user_id") ~= ARGV[2] then
			return -1
		end

		local fields = redis.call("HKEYS", holdKey)
		for _, field in ipairs(fields) do
			if string.sub(field, 1, 5) == "seat:" then
				local seatKey = redis.call("HGET", holdKey, field)
				if redis.call("GET", seatKey) == ARGV[1] then
					redis.call("DEL", seatKey)
				end
			end
		end

		redis.call("DEL", holdKey)
		return 1
	`
	result, err := client.Eval(
		ctx,
		releaseHoldScript,
		[]string{"hold:" + holdID.String()},
		holdID.String(),
		userID.String(),
	).Int()
	if err != nil {
		return errors.New("internal server error")
	}

	if result == 0 {
		return ErrHoldExpiredOrMissing
	}
	if result == -1 {
		return ErrHoldDoesntBelongToUser
	}
	if result == 1 {
		return nil
	}

	return errors.New("internal server error")
}

func ConfirmHold(ctx context.Context, client *redis.Client, movies *movieclient.Client, pool *pgxpool.Pool, userID uuid.UUID, holdID uuid.UUID) (models.Booking, error) {
	existingBooking, found, err := repository.GetBookingByHoldID(ctx, pool, holdID)
	if err != nil {
		return models.Booking{}, err
	}
	if found {
		if existingBooking.UserID != userID {
			return models.Booking{}, ErrHoldDoesntBelongToUser
		}
		// The booking is already committed. A failed cleanup can leave only a
		// temporary Redis hold; the database uniqueness constraint protects the seats.
		_ = DeleteHold(ctx, client, holdID, userID)
		return existingBooking, nil
	}

	holdKey := "hold:" + holdID.String()
	holdUserID, err := client.HGet(ctx, holdKey, "user_id").Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return models.Booking{}, ErrHoldExpiredOrMissing
		}
		return models.Booking{}, fmt.Errorf("read hold owner: %w", err)
	}
	if holdUserID != userID.String() {
		return models.Booking{}, ErrHoldDoesntBelongToUser
	}

	fields, err := client.HGetAll(ctx, holdKey).Result()
	if err != nil {
		return models.Booking{}, fmt.Errorf("read hold details: %w", err)
	}
	if len(fields) == 0 {
		return models.Booking{}, ErrHoldExpiredOrMissing
	}

	var showtimeID uuid.UUID
	var seatIDs []uuid.UUID
	for field, seatKey := range fields {
		if !strings.HasPrefix(field, "seat:") {
			continue
		}
		heldBy, err := client.Get(ctx, seatKey).Result()
		if errors.Is(err, redis.Nil) || (err == nil && heldBy != holdID.String()) {
			return models.Booking{}, ErrHoldExpiredOrMissing
		}
		if err != nil {
			return models.Booking{}, fmt.Errorf("verify held seat: %w", err)
		}

		parts := strings.SplitN(seatKey, ":", 3)
		if len(parts) != 3 || parts[0] != "hold" {
			return models.Booking{}, ErrInvalidHoldDetails
		}

		parsedShowtimeID, err := uuid.Parse(parts[1])
		if err != nil {
			return models.Booking{}, ErrInvalidHoldDetails
		}

		seatID, err := uuid.Parse(parts[2])
		if err != nil {
			return models.Booking{}, ErrInvalidHoldDetails
		}
		if showtimeID == uuid.Nil {
			showtimeID = parsedShowtimeID
		} else if showtimeID != parsedShowtimeID {
			return models.Booking{}, ErrInvalidHoldDetails
		}
		seatIDs = append(seatIDs, seatID)
	}
	if len(seatIDs) == 0 || showtimeID == uuid.Nil {
		return models.Booking{}, ErrHoldExpiredOrMissing
	}

	showtime, err := movies.GetShowtime(ctx, showtimeID)
	if err != nil {
		return models.Booking{}, err
	}

	booking := models.Booking{
		ID:         uuid.New(),
		HoldID:     holdID,
		UserID:     userID,
		ShowtimeID: showtimeID,
		Status:     "confirmed",
	}
	bookingSeats := make([]models.BookingSeat, 0, len(seatIDs))
	for _, seatID := range seatIDs {
		bookingSeats = append(bookingSeats, models.BookingSeat{
			BookingID:  booking.ID,
			ShowtimeID: showtimeID,
			SeatID:     seatID,
			PriceCents: showtime.PriceCents,
		})
	}
	_, err = repository.InsertBooking(ctx, pool, &booking, bookingSeats)
	if err != nil {
		return models.Booking{}, err
	}
	if booking.UserID != userID {
		return models.Booking{}, ErrHoldDoesntBelongToUser
	}

	if err := DeleteHold(ctx, client, holdID, userID); err != nil {
		if !errors.Is(err, ErrHoldExpiredOrMissing) {
			return booking, errors.New("internal server error")
		}
	}
	return booking, nil
}

func GetBooking(ctx context.Context, pool *pgxpool.Pool, bookingID uuid.UUID, userID uuid.UUID) (models.GetBookingResponse, error) {
	resp, err := repository.GetBooking(ctx, pool, bookingID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrBookingNotFound) {
			return models.GetBookingResponse{}, err
		}
		return models.GetBookingResponse{}, errors.New("internal server error")
	}

	return resp, err
}
