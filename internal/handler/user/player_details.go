package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"go.uber.org/zap"
)

// GetPlayerDetails - GET /api/admin/users/:user_id/details
// Get comprehensive player details including suspension history and balance logs
func (u *user) GetPlayerDetails(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		u.log.Error("Invalid user ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid user ID",
		})
		return
	}

	// Get player details
	player, exists, err := u.userModule.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		u.log.Error("Failed to get player details", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get player details",
		})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Code:    http.StatusNotFound,
			Message: "Player not found",
		})
		return
	}

	// Get suspension history
	suspensionHistory, err := u.userModule.GetPlayerSuspensionHistory(c.Request.Context(), userID)
	if err != nil {
		u.log.Error("Failed to get suspension history", zap.Error(err))
		// Don't fail the entire request, just log the error
		suspensionHistory = []dto.SuspensionHistory{}
	}

	// Get balance logs
	balanceLogs, err := u.userModule.GetPlayerBalanceLogs(c.Request.Context(), userID)
	if err != nil {
		u.log.Error("Failed to get balance logs", zap.Error(err))
		// Don't fail the entire request, just log the error
		balanceLogs = []dto.BalanceLog{}
	}

	// Get current balances
	balances, err := u.userModule.GetPlayerBalances(c.Request.Context(), userID)
	if err != nil {
		u.log.Error("Failed to get player balances", zap.Error(err))
		// Don't fail the entire request, just log the error
		balances = []dto.Balance{}
	}

	// Get game activity (if available)
	gameActivity, err := u.userModule.GetPlayerGameActivity(c.Request.Context(), userID)
	if err != nil {
		u.log.Error("Failed to get game activity", zap.Error(err))
		// Don't fail the entire request, just log the error
		gameActivity = []dto.GameActivity{}
	}

	// Get player statistics from database
	playerStats, err := u.userModule.GetPlayerStatistics(c.Request.Context(), userID)
	if err != nil {
		u.log.Error("Failed to get player statistics", zap.Error(err))
		// Don't fail the entire request, just log the error
		playerStats = dto.PlayerStatistics{}
	}

	playerDetails := dto.PlayerDetailsResponse{
		Player:            player,
		SuspensionHistory: suspensionHistory,
		BalanceLogs:       balanceLogs,
		Balances:          balances,
		GameActivity:      gameActivity,
		Statistics:        playerStats,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Player details retrieved successfully",
		"data":    playerDetails,
	})
}
