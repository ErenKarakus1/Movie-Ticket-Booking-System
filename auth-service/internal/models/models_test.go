package models

import "testing"

func TestRegisterRequestNormalize(t *testing.T) {
	req := RegisterRequest{
		Name:  "  Alice Example  ",
		Email: "  ALICE@Example.COM  ",
	}

	req.Normalize()

	if req.Name != "Alice Example" {
		t.Errorf("Name = %q, want %q", req.Name, "Alice Example")
	}
	if req.Email != "alice@example.com" {
		t.Errorf("Email = %q, want %q", req.Email, "alice@example.com")
	}
}

func TestLoginRequestNormalize(t *testing.T) {
	req := LoginRequest{Email: "  ALICE@Example.COM  ", Password: " keep as entered "}

	req.Normalize()

	if req.Email != "alice@example.com" {
		t.Errorf("Email = %q, want %q", req.Email, "alice@example.com")
	}
	if req.Password != " keep as entered " {
		t.Errorf("Password was modified: %q", req.Password)
	}
}
