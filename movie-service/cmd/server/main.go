package main

import (
	"log"

	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/movie-service/internal/config"
	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/movie-service/internal/db"
	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/movie-service/internal/handlers"
	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/movie-service/internal/middlewares"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	pool, err := db.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("connected to postgres")

	router := gin.Default()

	router.POST("/movies", middlewares.AdminMiddleware(), handlers.CreateMovieHandler(pool))
	router.GET("/movies", handlers.GetAllMoviesHandler(pool))
	router.GET("/movies/:id", handlers.GetMovieHandler(pool))
	router.PUT("/movies/:id", middlewares.AdminMiddleware(), handlers.UpdateMovieHandler(pool))

	if err := router.Run(":8082"); err != nil {
		log.Fatal(err)
	}
}
