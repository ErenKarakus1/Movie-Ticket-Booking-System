package handlers

import (
	"errors"
	"net/http"

	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/movie-service/internal/models"
	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/movie-service/internal/repository"
	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/movie-service/internal/services"
	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/movie-service/internal/validations"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateMovieHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req models.AddMovieRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		req.Normalize()
		if err := validations.ValidateAddMovieRequest(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		movie, err := services.CreateMovie(ctx.Request.Context(), pool, req)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusCreated, movie)
	}
}

func GetMovieHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		if id == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "movie id is required"})
			return
		}
		parsedID, err := uuid.Parse(id)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
			return
		}

		movie, err := services.GetMovie(ctx.Request.Context(), pool, parsedID)
		if err != nil {
			if errors.Is(err, repository.ErrMovieNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		ctx.JSON(http.StatusOK, movie)
	}
}

func GetAllMoviesHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		movies, err := services.GetAllMovies(ctx.Request.Context(), pool)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		if movies == nil {
			movies = []models.Movie{}
		}
		ctx.JSON(http.StatusOK, movies)
	}
}
