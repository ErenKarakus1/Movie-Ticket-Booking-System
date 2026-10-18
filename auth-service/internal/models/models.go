package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Role         string    `json:"role"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	ID        uuid.UUID `json:"id"`
	Role      string    `json:"role"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func (r *RegisterRequest) Normalize() {
	r.Name = strings.TrimSpace(r.Name)
	r.Email = strings.ToLower(strings.TrimSpace(r.Email))
}
