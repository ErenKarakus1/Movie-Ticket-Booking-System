package jwt

import (
	"errors"
	"time"

	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/auth-service/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(jwtSecret string, user models.User) (string, error) {
	claims := jwt.MapClaims{
		"id":    user.ID.String(),
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(time.Hour * 8).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	ss, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", errors.New("internal server error")
	}

	return ss, nil
}
