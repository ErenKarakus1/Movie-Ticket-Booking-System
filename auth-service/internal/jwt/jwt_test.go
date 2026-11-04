package jwt

import (
	"testing"
	"time"

	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/auth-service/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestGenerateTokenContainsUserClaimsAndExpires(t *testing.T) {
	user := models.User{ID: uuid.New(), Role: "user", Email: "reader@example.com"}
	tokenString, err := GenerateToken("test-secret", user)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	parsed, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			t.Fatalf("unexpected signing method: %v", token.Method)
		}
		return []byte("test-secret"), nil
	})
	if err != nil {
		t.Fatalf("jwt.Parse() error = %v", err)
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		t.Fatal("generated token should have valid map claims")
	}
	if claims["id"] != user.ID.String() || claims["role"] != user.Role || claims["email"] != user.Email {
		t.Fatalf("unexpected token claims: %#v", claims)
	}
	expiresAt, err := claims.GetExpirationTime()
	if err != nil {
		t.Fatalf("GetExpirationTime() error = %v", err)
	}
	if expiresAt == nil || !expiresAt.Time.After(time.Now()) {
		t.Fatalf("expected future expiration, got %v", expiresAt)
	}

	_, err = jwt.Parse(tokenString, func(token *jwt.Token) (any, error) { return []byte("wrong-secret"), nil })
	if err == nil {
		t.Fatal("token should not verify with a different secret")
	}
}
