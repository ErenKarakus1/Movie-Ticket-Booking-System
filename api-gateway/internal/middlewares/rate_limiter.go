package middlewares

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RateLimiter(client *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.GetHeader("X-User-ID")
		if userID == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user id is required"})
			return
		}
		key := "rate_limit:" + userID
		count, err := client.Incr(context.Background(), key).Result()
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		if count == 1 {
			if err := client.Expire(context.Background(), key, window).Err(); err != nil {
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				return
			}
		}
		if count > int64(limit) {
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		ctx.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		ctx.Header("X-RateLimit-Remaining", strconv.Itoa(limit-int(count)))
		ctx.Next()
	}
}
