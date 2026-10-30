package main

import (
	"context"
	"log"

	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/ticket-service/internal/config"
	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/ticket-service/internal/db"
	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/ticket-service/internal/handlers"
	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/ticket-service/internal/middlewares"
	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/ticket-service/internal/movieclient"
	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/ticket-service/internal/redis"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()
	client, err := redis.NewClient(cfg.RedisURL)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("redis client created")

	movies, err := movieclient.New(cfg.MovieServiceURL)
	if err != nil {
		log.Fatal(err)
	}
	pool, err := db.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	router.POST("/showtimes/:id/holds", middlewares.UserMiddleware(), handlers.CreateHold(client, movies))
	router.DELETE("/holds/:id", middlewares.UserMiddleware(), handlers.DeleteHoldHandler(client))
	router.POST("/holds/:id/confirm", middlewares.UserMiddleware(), handlers.ConfirmHoldHandler(client, movies, pool))
	router.GET("/bookings/:id", middlewares.UserMiddleware(), handlers.GetBookingHandler(pool))
	router.GET("/users/me/bookings", middlewares.UserMiddleware(), handlers.GetAllBookingsHandler(pool))

	if err := router.Run(":8083"); err != nil {
		log.Fatal(err)
	}
}
