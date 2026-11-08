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

func TestValidateRegisterRequestAcceptsValidInput(t *testing.T) {
	req := models.RegisterRequest{
		Name:     "Alice Example",
		Email:    "alice@example.com",
		Password: "correct-horse-123",
	}
	if err := ValidateRegisterRequest(req); err != nil {
		t.Fatalf("expected valid registration request, got error: %v", err)
	}
}

func TestValidateRegisterRequestAcceptsMaximumPasswordLength(t *testing.T) {
	req := models.RegisterRequest{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "12345678901234567890123456789012",
	}
	if err := ValidateRegisterRequest(req); err != nil {
		t.Fatalf("expected 32-character password to be accepted, got error: %v", err)
	}
}

func TestValidateRegisterRequestRejectsNameOverMaximum(t *testing.T) {
	req := models.RegisterRequest{
		Name:     "123456789012345678901234567890123456789012345678901",
		Email:    "alice@example.com",
		Password: "password123",
	}
	if err := ValidateRegisterRequest(req); err == nil {
		t.Fatal("expected name longer than 50 characters to be rejected")
	}
}

func TestValidateLoginRequestAcceptsValidInput(t *testing.T) {
	req := models.LoginRequest{Email: "alice@example.com", Password: "password123"}
	if err := ValidateLoginRequest(req); err != nil {
		t.Fatalf("expected valid login request, got error: %v", err)
	}
}
