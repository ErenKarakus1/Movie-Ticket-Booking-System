package main

import (
	"log"
	"time"

	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/api-gateway/internal/config"
	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/api-gateway/internal/middlewares"
	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/api-gateway/internal/proxy"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {

	cfg := config.LoadConfig()

	authProxy := proxy.NewProxy("http://localhost:8081")
	movieProxy := proxy.NewProxy("http://localhost:8082")
	ticketProxy := proxy.NewProxy("http://localhost:8083")

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer redisClient.Close()

	router := gin.Default()

	router.POST("/auth/register", authProxy)
	router.POST("/auth/login", authProxy)

	router.GET("/movies", movieProxy)
	router.GET("/movies/:id", movieProxy)
	router.GET("/movies/:id/showtimes", movieProxy)
	router.GET("/showtimes/:id", movieProxy)
	router.GET("/showtimes/:id/seats", ticketProxy)

	protected := router.Group("/")
	protected.Use(middlewares.AuthMiddleware(cfg.JWTSecret), middlewares.RateLimiter(redisClient, 100, time.Minute))

	protected.POST("/movies", movieProxy)
	protected.PUT("/movies/:id", movieProxy)
	protected.POST("/auditoriums", movieProxy)
	protected.POST("/movies/:id/showtimes", movieProxy)
	protected.POST("/auditoriums/:id/seats", movieProxy)

	protected.POST("/showtimes/:id/holds", ticketProxy)
	protected.DELETE("/holds/:id", ticketProxy)
	protected.POST("/holds/:id/confirm", ticketProxy)
	protected.GET("/bookings/:id", ticketProxy)
	protected.POST("/bookings/:id/redeem", ticketProxy)
	protected.GET("/users/me/bookings", ticketProxy)

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
