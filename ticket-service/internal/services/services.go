package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/ticket-service/internal/models"
	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/ticket-service/internal/movieclient"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var ErrSeatAlreadyHeld = errors.New("seat already held")
var ErrSeatNotInShowtime = errors.New("seat does not belong to showtime auditorium")
var ErrShowtimeStarted = errors.New("showtime has already started")
var ErrHoldExpiredOrMissing = errors.New("hold is missing or expired")
var ErrHoldDoesntBelongToUser = errors.New("hold belongs to another user")

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
