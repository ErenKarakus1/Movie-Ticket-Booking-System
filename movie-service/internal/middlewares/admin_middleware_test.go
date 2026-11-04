package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAdminMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name string
		role string
		want int
	}{
		{name: "admin allowed", role: "admin", want: http.StatusNoContent},
		{name: "user denied", role: "user", want: http.StatusForbidden},
		{name: "missing role denied", role: "", want: http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/admin", AdminMiddleware(), func(c *gin.Context) { c.Status(http.StatusNoContent) })
			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			if tc.role != "" {
				req.Header.Set("X-User-Role", tc.role)
			}
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			if resp.Code != tc.want {
				t.Fatalf("status = %d, want %d", resp.Code, tc.want)
			}
		})
	}
}
