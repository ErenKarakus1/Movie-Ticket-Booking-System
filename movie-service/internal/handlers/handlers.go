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

func UpdateMovieHandler(pool *pgxpool.Pool) gin.HandlerFunc {
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
		updatedMovie, err := services.UpdateMovie(ctx.Request.Context(), pool, req, parsedID)
		if err != nil {
			if errors.Is(err, repository.ErrMovieNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		ctx.JSON(http.StatusOK, updatedMovie)
	}
}

func CreateAuditoriumHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req models.CreateAuditoriumRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		req.Normalize()
		if err := validations.ValidateCreateAuditoriumRequest(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		auditorium, err := services.CreateAuditorium(ctx.Request.Context(), pool, req)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusCreated, auditorium)
	}
}

func AddShowtimeHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		movieID := ctx.Param("id")
		if movieID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "movie id is required"})
			return
		}
		parsedMovieID, err := uuid.Parse(movieID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
			return
		}
		var req models.AddShowtimeRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		if err := validations.ValidateAddShowtimeRequest(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		showtime, err := services.AddShowtime(ctx.Request.Context(), pool, req, parsedMovieID)
		if err != nil {
			if errors.Is(err, repository.ErrAuditoriumNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "auditorium not found"})
				return
			}
			if errors.Is(err, repository.ErrMovieNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
				return
			}
			if errors.Is(err, repository.ErrOverlap) {
				ctx.JSON(http.StatusConflict, gin.H{"error": "showtime overlaps with another screening at the auditorium"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusCreated, showtime)
	}
}

func GetMovieShowtimesHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		movieID := ctx.Param("id")
		if movieID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "movie id is required"})
			return
		}
		parsedMovieID, err := uuid.Parse(movieID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
			return
		}

		showtimes, err := services.GetMovieShowtimes(ctx.Request.Context(), pool, parsedMovieID)
		if err != nil {
			if errors.Is(err, repository.ErrMovieNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		if showtimes == nil {
			showtimes = []models.Showtime{}
		}

		ctx.JSON(http.StatusOK, showtimes)
	}
}

func CreateSeatsHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		auditoriumID := ctx.Param("id")
		if auditoriumID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "auditorium id is required"})
			return
		}
		parsedAuditoriumID, err := uuid.Parse(auditoriumID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid auditorium id"})
			return
		}

		var req []models.CreateSeatRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		req = models.NormalizeCreateSeatsRequest(req)
		if err := validations.ValidateCreateSeatsRequest(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		seats, err := services.CreateSeats(ctx.Request.Context(), pool, parsedAuditoriumID, req)
		if err != nil {
			if errors.Is(err, repository.ErrSeatExists) {
				ctx.JSON(http.StatusConflict, gin.H{"error": "one or more seats are already exists in the auditorium"})
				return
			}
			if errors.Is(err, repository.ErrAuditoriumNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "auditorium not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusCreated, seats)
	}
}

func GetSeatsHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		showtimeID := ctx.Param("id")
		if showtimeID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "showtime id is required"})
			return
		}
		parsedShowtimeID, err := uuid.Parse(showtimeID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid showtime id"})
			return
		}

		seats, err := services.GetSeats(ctx, pool, parsedShowtimeID)
		if err != nil {
			if errors.Is(err, repository.ErrShowtimeNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "showtime not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		if seats == nil {
			seats = []models.Seat{}
		}
		ctx.JSON(http.StatusOK, seats)
	}
}
