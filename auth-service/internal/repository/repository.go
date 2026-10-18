package repository

import (
	"context"
	"errors"

	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/auth-service/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEmailAlreadyRegistered = errors.New("email is already registered")
var ErrUserNotFound = errors.New("user not found")

const insertUserSQL = `
	INSERT INTO users (
		id,
		role,
		name,
		email,
		password_hash
	)
	VALUES ($1,$2,$3,$4,$5)
	RETURNING
		id,
		role,
		name,
		email,
		created_at
`

const getUserByEmailSQL = `
	SELECT
		id,
		role,
		name,
		email,
		password_hash,
		created_at
	FROM users
	WHERE email=$1
`

func InsertUser(ctx context.Context, pool *pgxpool.Pool, user models.User) (models.RegisterResponse, error) {
	var resp models.RegisterResponse
	err := pool.QueryRow(
		ctx,
		insertUserSQL,
		user.ID,
		user.Role,
		user.Name,
		user.Email,
		user.PasswordHash,
	).Scan(
		&resp.ID,
		&resp.Role,
		&resp.Name,
		&resp.Email,
		&resp.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.RegisterResponse{}, ErrEmailAlreadyRegistered
		}
		return models.RegisterResponse{}, errors.New("internal server error")
	}

	return resp, nil
}

func GetUserByEmail(ctx context.Context, pool *pgxpool.Pool, email string) (models.User, error) {
	var user models.User
	err := pool.QueryRow(
		ctx,
		getUserByEmailSQL,
		email,
	).Scan(
		&user.ID,
		&user.Role,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, errors.New("internal server error")
	}

	return user, nil
}
