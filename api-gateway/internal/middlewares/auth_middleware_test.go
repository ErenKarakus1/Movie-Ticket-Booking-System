package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestAuthMiddlewareRejectsMissingMalformedAndWrongSecretTokens(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := []byte("test-secret")
	wrongSecretToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"id": "user-id", "role": "user"})
	wrongSecret, err := wrongSecretToken.SignedString([]byte("different-secret"))
	if err != nil {
		t.Fatal(err)
	}
	validNoIdentityToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"role": "user"})
	validNoIdentity, err := validNoIdentityToken.SignedString(secret)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		header string
	}{
		{"missing header", ""},
		{"wrong scheme", "Basic token"},
		{"malformed token", "Bearer not-a-jwt"},
		{"wrong signing secret", "Bearer " + wrongSecret},
		{"missing identity claim", "Bearer " + validNoIdentity},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/protected", AuthMiddleware(string(secret)), func(c *gin.Context) { c.Status(http.StatusNoContent) })
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			if resp.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d; body=%s", resp.Code, http.StatusUnauthorized, resp.Body.String())
			}
		})
	}
}

func TestAuthMiddlewareOverwritesForgedIdentityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	claims := jwt.MapClaims{"id": "signed-user", "role": "user"}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.GET("/protected", AuthMiddleware("test-secret"), func(c *gin.Context) {
		c.String(http.StatusOK, c.GetHeader("X-User-ID")+":"+c.GetHeader("X-User-Role"))
	})
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-User-ID", "forged-user")
	req.Header.Set("X-User-Role", "admin")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK || resp.Body.String() != "signed-user:user" {
		t.Fatalf("got status %d and body %q", resp.Code, resp.Body.String())
	}
}
