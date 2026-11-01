package validations

import (
	"testing"

	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/auth-service/internal/models"
)

func TestValidateRegisterRequestRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name string
		req  models.RegisterRequest
	}{
		{"missing name", models.RegisterRequest{Email: "user@example.com", Password: "password123"}},
		{"short name", models.RegisterRequest{Name: "Al", Email: "user@example.com", Password: "password123"}},
		{"invalid email", models.RegisterRequest{Name: "Alice", Email: "invalid", Password: "password123"}},
		{"missing password", models.RegisterRequest{Name: "Alice", Email: "user@example.com"}},
		{"short password", models.RegisterRequest{Name: "Alice", Email: "user@example.com", Password: "short"}},
		{"password over bcrypt byte limit", models.RegisterRequest{Name: "Alice", Email: "user@example.com", Password: "1234567890123456789012345678901234567890123456789012345678901234567890123"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateRegisterRequest(tt.req); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestValidateLoginRequestRejectsInvalidInput(t *testing.T) {
	for _, req := range []models.LoginRequest{
		{Email: "", Password: "password123"},
		{Email: "not-an-email", Password: "password123"},
		{Email: "user@example.com", Password: ""},
	} {
		if err := ValidateLoginRequest(req); err == nil {
			t.Fatalf("expected validation error for %#v", req)
		}
	}
}
