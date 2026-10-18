package password

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func GeneratePasswordHash(password string) (string, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("internal server error")
	}
	return string(passwordHash), nil
}
