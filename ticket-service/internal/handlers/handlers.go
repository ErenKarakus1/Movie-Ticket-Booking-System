package handlers

import (
	"errors"
	"net/http"

	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/ticket-service/internal/models"
	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/ticket-service/internal/movieclient"
	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/ticket-service/internal/services"
	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/ticket-service/internal/validations"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func CreateHold(client *redis.Client, movies *movieclient.Client) gin.HandlerFunc {
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

		userID := ctx.GetHeader("X-User-ID")
		if userID == "" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		parsedUserID, err := uuid.Parse(userID)
		if err != nil {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		var req models.CreateHoldRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		if err := validations.ValidateCreateHoldRequest(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		holdDetails, err := services.CreateHold(ctx.Request.Context(), client, movies, parsedShowtimeID, req.SeatIDs, parsedUserID)
		if err != nil {
			if errors.Is(err, services.ErrSeatAlreadyHeld) {
				ctx.JSON(http.StatusConflict, gin.H{"error": "seat is already held"})
				return
			}
			if errors.Is(err, movieclient.ErrShowtimeNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "showtime not found"})
				return
			}
			if errors.Is(err, services.ErrSeatNotInShowtime) {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "one or more seats do not belong to this showtime"})
				return
			}
			if errors.Is(err, services.ErrShowtimeStarted) {
				ctx.JSON(http.StatusConflict, gin.H{"error": "showtime has already started"})
				return
			}
			if errors.Is(err, movieclient.ErrUnavailable) {
				ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "movie service unavailable"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		ctx.JSON(http.StatusCreated, holdDetails)
	}
}

func DeleteHoldHandler(client *redis.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.GetHeader("X-User-ID")
		if userID == "" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		parsedUserID, err := uuid.Parse(userID)
		if err != nil {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		holdID := ctx.Param("id")
		if holdID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "hold id is required"})
			return
		}
		parsedHoldID, err := uuid.Parse(holdID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid hold id"})
			return
		}

		err = services.DeleteHold(ctx.Request.Context(), client, parsedHoldID, parsedUserID)
		if err != nil {
			if errors.Is(err, services.ErrHoldExpiredOrMissing) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "hold expired or missing"})
				return
			}
			if errors.Is(err, services.ErrHoldDoesntBelongToUser) {
				ctx.JSON(http.StatusForbidden, gin.H{"error": "hold belongs to another user"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		ctx.Status(http.StatusOK)
	}
}
