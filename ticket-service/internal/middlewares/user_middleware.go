package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func UserMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		role := ctx.GetHeader("X-User-Role")
		if role != "user" {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		ctx.Next()
	}
}
