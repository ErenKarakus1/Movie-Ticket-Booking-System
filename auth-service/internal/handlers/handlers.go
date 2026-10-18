package handlers

import (
	"errors"
	"net/http"

	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/auth-service/internal/models"
	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/auth-service/internal/repository"
	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/auth-service/internal/services"
	"github.com/ErenKarakus1/Movie-Ticket-Booking-System/auth-service/internal/validations"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req models.RegisterRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		req.Normalize()
		if err := validations.ValidateRegisterRequest(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		user, err := services.CreateUser(ctx.Request.Context(), pool, req)
		if err != nil {
			if errors.Is(err, repository.ErrEmailAlreadyRegistered) {
				ctx.JSON(http.StatusConflict, gin.H{"error": "email is already registered"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusCreated, user)
	}
}
