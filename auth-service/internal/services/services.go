package services

import (
	"context"
	"errors"
	"net/mail"

	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/auth-service/internal/models"
	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/auth-service/internal/password"
	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/auth-service/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func generateUser(req models.RegisterRequest) (models.User, error) {
	passwordHash, err := password.GeneratePasswordHash(req.Password)
	if err != nil {
		return models.User{}, errors.New("internal server error")
	}
	mail, err := mail.ParseAddress(req.Email)
	if err != nil {
		return models.User{}, errors.New("internal server error")
	}
	return models.User{
		ID:           uuid.New(),
		Role:         "user",
		Name:         req.Name,
		Email:        mail.Address,
		PasswordHash: passwordHash,
	}, nil
}

func CreateUser(ctx context.Context, pool *pgxpool.Pool, req models.RegisterRequest) (models.RegisterResponse, error) {
	user, err := generateUser(req)
	if err != nil {
		return models.RegisterResponse{}, errors.New("internal server error")
	}
	insertedUser, err := repository.InsertUser(ctx, pool, user)
	if err != nil {
		return models.RegisterResponse{}, err
	}

	return insertedUser, nil
}
