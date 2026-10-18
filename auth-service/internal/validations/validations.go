package validations

import (
	"errors"
	"net/mail"
	"unicode/utf8"

	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/auth-service/internal/models"
)

func ValidateRegisterRequest(r models.RegisterRequest) error {
	nameLength := utf8.RuneCountInString(r.Name)

	if r.Name == "" {
		return errors.New("name is required")
	}

	if nameLength < 3 {
		return errors.New("name must be at least 3 characters")
	}

	if nameLength > 50 {
		return errors.New("name must be at most 50 characters")
	}

	if r.Email == "" {
		return errors.New("email is required")
	}

	if len(r.Email) > 128 {
		return errors.New("email must be at most 128 bytes")
	}
	_, err := mail.ParseAddress(r.Email)
	if err != nil {
		return errors.New("invalid email")
	}

	passwordLength := utf8.RuneCountInString(r.Password)

	if r.Password == "" {
		return errors.New("password is required")
	}

	if passwordLength < 8 {
		return errors.New("password must be at least 8 characters")
	}

	if passwordLength > 32 {
		return errors.New("password must be at most 32 characters")
	}

	if len(r.Password) > 72 {
		return errors.New("password must be at most 72 bytes")
	}

	return nil

}
