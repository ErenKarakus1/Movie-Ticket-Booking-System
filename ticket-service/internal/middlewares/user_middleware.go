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

func AdminMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.GetHeader("X-User-Role") != "admin" {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		ctx.Next()
	}
}
