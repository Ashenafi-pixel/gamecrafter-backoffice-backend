package cashback

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/module/cashback"
	"go.uber.org/zap"
)

type CashbackHandler struct {
	cashbackService *cashback.CashbackService
	logger          *zap.Logger
}

func NewCashbackHandler(cashbackService *cashback.CashbackService, logger *zap.Logger) *CashbackHandler {
	return &CashbackHandler{
		cashbackService: cashbackService,
		logger:          logger,
	}
}

// GetUserCashbackSummary returns the user's cashback summary
// @Summary Get user cashback summary
// @Description Returns comprehensive cashback information for the authenticated user
// @Tags Cashback
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserCashbackSummary
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /user/cashback [get]
func (h *CashbackHandler) GetUserCashbackSummary(c *gin.Context) {
	userID := c.GetString("user-id")
	if userID == "" {
		h.logger.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "User not authenticated",
		})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		h.logger.Error("Invalid user ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid user ID",
		})
		return
	}

	h.logger.Info("Getting user cashback summary", zap.String("user_id", userUUID.String()))

	summary, err := h.cashbackService.GetUserCashbackSummary(c.Request.Context(), userUUID)
	if err != nil {
		h.logger.Error("Failed to get user cashback summary", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get cashback summary",
		})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// ClaimCashback processes a cashback claim request
// @Summary Claim cashback
// @Description Allows users to claim their available cashback
// @Tags Cashback
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CashbackClaimRequest true "Cashback claim request"
// @Success 200 {object} dto.CashbackClaimResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /user/cashback/claim [post]
func (h *CashbackHandler) ClaimCashback(c *gin.Context) {
	userID := c.GetString("user-id")
	if userID == "" {
		h.logger.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "User not authenticated",
		})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		h.logger.Error("Invalid user ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid user ID",
		})
		return
	}

	var request dto.CashbackClaimRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Failed to bind cashback claim request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request format",
		})
		return
	}

	h.logger.Info("Processing cashback claim",
		zap.String("user_id", userUUID.String()),
		zap.String("amount", request.Amount.String()))

	response, err := h.cashbackService.ClaimCashback(c.Request.Context(), userUUID, request)
	if err != nil {
		h.logger.Error("Failed to process cashback claim", zap.Error(err))

		if err != nil && err.Error() == "invalid user input" {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to process cashback claim",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetCashbackTiers returns all available cashback tiers
// @Summary Get cashback tiers
// @Description Returns all available cashback tiers
// @Tags Cashback
// @Accept json
// @Produce json
// @Success 200 {array} dto.CashbackTier
// @Failure 500 {object} dto.ErrorResponse
// @Router /cashback/tiers [get]
func (h *CashbackHandler) GetCashbackTiers(c *gin.Context) {
	h.logger.Info("Getting cashback tiers")

	tiers, err := h.cashbackService.GetCashbackTiers(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get cashback tiers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get cashback tiers",
		})
		return
	}

	c.JSON(http.StatusOK, tiers)
}

// GetCashbackStats returns admin statistics for the cashback system
// @Summary Get cashback statistics
// @Description Returns comprehensive statistics for the cashback system (Admin only)
// @Tags Cashback
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.AdminCashbackStats
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/cashback/stats [get]
func (h *CashbackHandler) GetCashbackStats(c *gin.Context) {
	h.logger.Info("Getting cashback statistics")

	stats, err := h.cashbackService.GetCashbackStats(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get cashback stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get cashback statistics",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// CreateCashbackTier creates a new cashback tier (Admin only)
// @Summary Create cashback tier
// @Description Creates a new cashback tier (Admin only)
// @Tags Cashback
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CashbackTierUpdateRequest true "Cashback tier data"
// @Success 201 {object} dto.CashbackTier
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/cashback/tiers [post]
func (h *CashbackHandler) CreateCashbackTier(c *gin.Context) {
	var request dto.CashbackTierUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Failed to bind cashback tier request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request format",
		})
		return
	}

	h.logger.Info("Creating cashback tier", zap.String("tier_name", request.TierName))

	// Convert request to CashbackTier DTO
	tier := dto.CashbackTier{
		TierName:               request.TierName,
		MinExpectedGGRRequired: request.MinGGRRequired,
		CashbackPercentage:     request.CashbackPercentage,
		BonusMultiplier:        request.BonusMultiplier,
		DailyCashbackLimit:     request.DailyCashbackLimit,
		WeeklyCashbackLimit:    request.WeeklyCashbackLimit,
		MonthlyCashbackLimit:   request.MonthlyCashbackLimit,
		SpecialBenefits:        request.SpecialBenefits,
		IsActive:               request.IsActive,
	}

	createdTier, err := h.cashbackService.CreateCashbackTier(c.Request.Context(), tier)
	if err != nil {
		h.logger.Error("Failed to create cashback tier", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create cashback tier",
		})
		return
	}

	h.logger.Info("Created cashback tier successfully",
		zap.String("tier_id", createdTier.ID.String()),
		zap.String("tier_name", createdTier.TierName))

	c.JSON(http.StatusCreated, createdTier)
}

// UpdateCashbackTier updates an existing cashback tier (Admin only)
// @Summary Update cashback tier
// @Description Updates an existing cashback tier (Admin only)
// @Tags Cashback
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Tier ID"
// @Param request body dto.CashbackTierUpdateRequest true "Cashback tier data"
// @Success 200 {object} dto.CashbackTier
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/cashback/tiers/{id} [put]
func (h *CashbackHandler) UpdateCashbackTier(c *gin.Context) {
	tierIDStr := c.Param("id")
	tierID, err := uuid.Parse(tierIDStr)
	if err != nil {
		h.logger.Error("Invalid tier ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid tier ID",
		})
		return
	}

	var request dto.CashbackTierUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Failed to bind cashback tier request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request format",
		})
		return
	}

	h.logger.Info("Updating cashback tier",
		zap.String("tier_id", tierID.String()),
		zap.String("tier_name", request.TierName))

	// Convert request to CashbackTier DTO
	tier := dto.CashbackTier{
		TierName:               request.TierName,
		MinExpectedGGRRequired: request.MinGGRRequired,
		CashbackPercentage:     request.CashbackPercentage,
		BonusMultiplier:        request.BonusMultiplier,
		DailyCashbackLimit:     request.DailyCashbackLimit,
		WeeklyCashbackLimit:    request.WeeklyCashbackLimit,
		MonthlyCashbackLimit:   request.MonthlyCashbackLimit,
		SpecialBenefits:        request.SpecialBenefits,
		IsActive:               request.IsActive,
	}

	updatedTier, err := h.cashbackService.UpdateCashbackTier(c.Request.Context(), tierID, tier)
	if err != nil {
		h.logger.Error("Failed to update cashback tier", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update cashback tier",
		})
		return
	}

	h.logger.Info("Updated cashback tier successfully",
		zap.String("tier_id", tierID.String()),
		zap.String("tier_name", updatedTier.TierName))

	c.JSON(http.StatusOK, updatedTier)
}

// DeleteCashbackTier deletes a cashback tier (Admin only)
// @Summary Delete cashback tier
// @Description Deletes a cashback tier (Admin only)
// @Tags Cashback
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Tier ID"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/cashback/tiers/{id} [delete]
func (h *CashbackHandler) DeleteCashbackTier(c *gin.Context) {
	tierIDStr := c.Param("id")
	tierID, err := uuid.Parse(tierIDStr)
	if err != nil {
		h.logger.Error("Invalid tier ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid tier ID",
		})
		return
	}

	h.logger.Info("Deleting cashback tier", zap.String("tier_id", tierID.String()))

	err = h.cashbackService.DeleteCashbackTier(c.Request.Context(), tierID)
	if err != nil {
		h.logger.Error("Failed to delete cashback tier", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete cashback tier",
		})
		return
	}

	h.logger.Info("Deleted cashback tier successfully", zap.String("tier_id", tierID.String()))
	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Cashback tier deleted successfully",
	})
}

// ReorderCashbackTiers reorders cashback tiers (Admin only)
// @Summary Reorder cashback tiers
// @Description Reorders cashback tiers by updating their tier levels (Admin only)
// @Tags Cashback
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.ReorderTiersRequest true "Tier order data"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/cashback/tiers/reorder [post]
func (h *CashbackHandler) ReorderCashbackTiers(c *gin.Context) {
	var request dto.ReorderTiersRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Failed to bind reorder request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request format",
		})
		return
	}

	h.logger.Info("Reordering cashback tiers", zap.Int("tier_count", len(request.TierOrder)))

	err := h.cashbackService.ReorderCashbackTiers(c.Request.Context(), request.TierOrder)
	if err != nil {
		h.logger.Error("Failed to reorder cashback tiers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to reorder cashback tiers",
		})
		return
	}

	h.logger.Info("Reordered cashback tiers successfully")
	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Cashback tiers reordered successfully",
	})
}

// CreateCashbackPromotion creates a new cashback promotion (Admin only)
// @Summary Create cashback promotion
// @Description Creates a new cashback promotion (Admin only)
// @Tags Cashback
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CashbackPromotionRequest true "Cashback promotion data"
// @Success 201 {object} dto.CashbackPromotion
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/cashback/promotions [post]
func (h *CashbackHandler) CreateCashbackPromotion(c *gin.Context) {
	var request dto.CashbackPromotionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Failed to bind cashback promotion request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request format",
		})
		return
	}

	h.logger.Info("Creating cashback promotion", zap.String("promotion_name", request.PromotionName))

	// Implementation would go here
	c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
		Code:    http.StatusNotImplemented,
		Message: "Feature not implemented yet",
	})
}

// GetUserCashbackEarnings returns the user's cashback earnings history
// @Summary Get user cashback earnings
// @Description Returns the user's cashback earnings history
// @Tags Cashback
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {array} dto.CashbackEarning
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /user/cashback/earnings [get]
func (h *CashbackHandler) GetUserCashbackEarnings(c *gin.Context) {
	userID := c.GetString("user-id")
	if userID == "" {
		h.logger.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "User not authenticated",
		})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		h.logger.Error("Invalid user ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid user ID",
		})
		return
	}

	// Parse pagination parameters
	page := 1
	limit := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	h.logger.Info("Getting user cashback earnings",
		zap.String("user_id", userUUID.String()),
		zap.Int("page", page),
		zap.Int("limit", limit))

	// Get cashback earnings from service
	earnings, err := h.cashbackService.GetUserCashbackEarnings(c.Request.Context(), userUUID)
	if err != nil {
		h.logger.Error("Failed to get user cashback earnings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get cashback earnings",
		})
		return
	}

	// Apply pagination
	total := len(earnings)
	start := (page - 1) * limit
	end := start + limit

	if start > total {
		earnings = []dto.CashbackEarning{}
	} else {
		if end > total {
			end = total
		}
		earnings = earnings[start:end]
	}

	h.logger.Info("Retrieved user cashback earnings",
		zap.String("user_id", userUUID.String()),
		zap.Int("total", total),
		zap.Int("returned", len(earnings)),
		zap.Int("page", page),
		zap.Int("limit", limit))

	c.JSON(http.StatusOK, gin.H{
		"earnings": earnings,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

// GetUserCashbackClaims returns the user's cashback claims history
// @Summary Get user cashback claims
// @Description Returns the user's cashback claims history
// @Tags Cashback
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {array} dto.CashbackClaim
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /user/cashback/claims [get]
func (h *CashbackHandler) GetUserCashbackClaims(c *gin.Context) {
	userID := c.GetString("user-id")
	if userID == "" {
		h.logger.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "User not authenticated",
		})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		h.logger.Error("Invalid user ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid user ID",
		})
		return
	}

	// Parse pagination parameters
	page := 1
	limit := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	h.logger.Info("Getting user cashback claims",
		zap.String("user_id", userUUID.String()),
		zap.Int("page", page),
		zap.Int("limit", limit))

	// Implementation would go here
	c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
		Code:    http.StatusNotImplemented,
		Message: "Feature not implemented yet",
	})
}

// Admin Dashboard Methods

// GetDashboardStats returns comprehensive dashboard statistics
func (h *CashbackHandler) GetDashboardStats(c *gin.Context) {
	adminHandler := NewAdminDashboardHandler(h.cashbackService, h.logger)
	adminHandler.GetDashboardStats(c)
}

// GetCashbackAnalytics returns detailed analytics for the cashback system
func (h *CashbackHandler) GetCashbackAnalytics(c *gin.Context) {
	adminHandler := NewAdminDashboardHandler(h.cashbackService, h.logger)
	adminHandler.GetCashbackAnalytics(c)
}

// GetSystemHealth returns the health status of the cashback system
func (h *CashbackHandler) GetSystemHealth(c *gin.Context) {
	adminHandler := NewAdminDashboardHandler(h.cashbackService, h.logger)
	adminHandler.GetSystemHealth(c)
}

// GetUserCashbackDetails returns detailed cashback information for a specific user
func (h *CashbackHandler) GetUserCashbackDetails(c *gin.Context) {
	adminHandler := NewAdminDashboardHandler(h.cashbackService, h.logger)
	adminHandler.GetUserCashbackDetails(c)
}

// ProcessManualCashback manually processes cashback for a user
func (h *CashbackHandler) ProcessManualCashback(c *gin.Context) {
	adminHandler := NewAdminDashboardHandler(h.cashbackService, h.logger)
	adminHandler.ProcessManualCashback(c)
}

// ValidateBalanceSync validates balance synchronization for the authenticated user
// @Summary Validate balance synchronization
// @Description Validates if user balances are synchronized between main and GrooveTech systems
// @Tags Balance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /user/balance/validate-sync [get]
func (h *CashbackHandler) ValidateBalanceSync(c *gin.Context) {
	userID := c.GetString("user-id")
	if userID == "" {
		h.logger.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "User not authenticated",
		})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		h.logger.Error("Invalid user ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid user ID",
		})
		return
	}

	// Validate balance synchronization
	status, err := h.cashbackService.ValidateBalanceSync(c.Request.Context(), userUUID)
	if err != nil {
		h.logger.Error("Failed to validate balance sync", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to validate balance synchronization",
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// ReconcileBalances reconciles user balances between main and GrooveTech systems
// @Summary Reconcile user balances
// @Description Synchronizes GrooveTech account balance with main balance
// @Tags Balance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /user/balance/reconcile [post]
func (h *CashbackHandler) ReconcileBalances(c *gin.Context) {
	userID := c.GetString("user-id")
	if userID == "" {
		h.logger.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "User not authenticated",
		})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		h.logger.Error("Invalid user ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid user ID",
		})
		return
	}

	// Reconcile balances
	err = h.cashbackService.ReconcileBalances(c.Request.Context(), userUUID)
	if err != nil {
		h.logger.Error("Failed to reconcile balances", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to reconcile balances",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Balances reconciled successfully",
		"user_id": userUUID.String(),
	})
}

// GetGameHouseEdge returns the house edge configuration for a specific game type
// @Summary Get game house edge
// @Description Returns the house edge configuration for a specific game type
// @Tags House Edge
// @Accept json
// @Produce json
// @Param game_type query string true "Game type (e.g., groovetech, plinko, crash)"
// @Param game_variant query string false "Game variant (optional)"
// @Success 200 {object} dto.GameHouseEdge
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /cashback/house-edge [get]
func (h *CashbackHandler) GetGameHouseEdge(c *gin.Context) {
	gameType := c.Query("game_type")
	if gameType == "" {
		h.logger.Error("Game type is required")
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Game type is required",
		})
		return
	}

	gameVariant := c.Query("game_variant")

	houseEdge, err := h.cashbackService.GetGameHouseEdge(c.Request.Context(), gameType, gameVariant)
	if err != nil {
		h.logger.Error("Failed to get game house edge", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get game house edge",
		})
		return
	}

	c.JSON(http.StatusOK, houseEdge)
}

// CreateGameHouseEdge creates a new game house edge configuration
// @Summary Create game house edge
// @Description Creates a new game house edge configuration
// @Tags House Edge
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param houseEdge body dto.GameHouseEdge true "House edge configuration"
// @Success 201 {object} dto.GameHouseEdge
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/cashback/house-edge [post]
func (h *CashbackHandler) CreateGameHouseEdge(c *gin.Context) {
	var houseEdge dto.GameHouseEdge
	if err := c.ShouldBindJSON(&houseEdge); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
		})
		return
	}

	createdHouseEdge, err := h.cashbackService.CreateGameHouseEdge(c.Request.Context(), houseEdge)
	if err != nil {
		h.logger.Error("Failed to create game house edge", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create game house edge",
		})
		return
	}

	c.JSON(http.StatusCreated, createdHouseEdge)
}

// GetRetryableOperations returns retryable operations for a user
func (h *CashbackHandler) GetRetryableOperations(c *gin.Context) {
	userID := c.GetString("user-id")
	if userID == "" {
		h.logger.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "User not authenticated",
		})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		h.logger.Error("Invalid user ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid user ID",
		})
		return
	}

	operations, err := h.cashbackService.GetRetryableOperations(c.Request.Context(), userUUID)
	if err != nil {
		h.logger.Error("Failed to get retryable operations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get retryable operations",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"operations": operations,
		"count":      len(operations),
	})
}

// ManualRetryOperation manually retries a specific operation
func (h *CashbackHandler) ManualRetryOperation(c *gin.Context) {
	operationIDStr := c.Param("operation_id")
	if operationIDStr == "" {
		h.logger.Error("Operation ID is required")
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Operation ID is required",
		})
		return
	}

	operationID, err := uuid.Parse(operationIDStr)
	if err != nil {
		h.logger.Error("Invalid operation ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid operation ID",
		})
		return
	}

	err = h.cashbackService.ManualRetryOperation(c.Request.Context(), operationID)
	if err != nil {
		h.logger.Error("Failed to manually retry operation", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to manually retry operation",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Operation retry initiated successfully",
		"operation_id": operationID.String(),
	})
}

// RetryFailedOperations retries all failed operations (admin only)
func (h *CashbackHandler) RetryFailedOperations(c *gin.Context) {
	err := h.cashbackService.RetryFailedOperations(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to retry failed operations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to retry failed operations",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Failed operations retry completed successfully",
	})
}

// GetLevelProgressionInfo returns detailed level progression information for a user
// @Summary Get user level progression info
// @Description Returns detailed level progression information including current tier, next tier, and progress
// @Tags Cashback
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.LevelProgressionInfo
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /user/cashback/level-progression [get]
func (h *CashbackHandler) GetLevelProgressionInfo(c *gin.Context) {
	userID := c.GetString("user-id")
	if userID == "" {
		h.logger.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "User not authenticated",
		})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		h.logger.Error("Invalid user ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid user ID",
		})
		return
	}

	h.logger.Info("Getting level progression info",
		zap.String("user_id", userUUID.String()))

	progressionInfo, err := h.cashbackService.GetLevelProgressionInfo(c.Request.Context(), userUUID)
	if err != nil {
		h.logger.Error("Failed to get level progression info", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get level progression info",
		})
		return
	}

	h.logger.Info("Retrieved level progression info",
		zap.String("user_id", userUUID.String()),
		zap.Int("current_level", progressionInfo.CurrentLevel),
		zap.String("current_tier", progressionInfo.CurrentTier.TierName))

	c.JSON(http.StatusOK, progressionInfo)
}

// ProcessBulkLevelProgression processes level progression for multiple users (admin only)
// @Summary Process bulk level progression
// @Description Processes level progression for multiple users
// @Tags Cashback
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_ids body []string true "Array of user IDs"
// @Success 200 {array} dto.LevelProgressionResult
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/cashback/bulk-level-progression [post]
func (h *CashbackHandler) ProcessBulkLevelProgression(c *gin.Context) {
	var request struct {
		UserIDs []string `json:"user_ids" validate:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
		})
		return
	}

	// Parse user IDs
	var userUUIDs []uuid.UUID
	for _, userIDStr := range request.UserIDs {
		userUUID, err := uuid.Parse(userIDStr)
		if err != nil {
			h.logger.Error("Invalid user ID format", zap.String("user_id", userIDStr), zap.Error(err))
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("Invalid user ID: %s", userIDStr),
			})
			return
		}
		userUUIDs = append(userUUIDs, userUUID)
	}

	h.logger.Info("Processing bulk level progression",
		zap.Int("user_count", len(userUUIDs)))

	results, err := h.cashbackService.ProcessBulkLevelProgression(c.Request.Context(), userUUIDs)
	if err != nil {
		h.logger.Error("Failed to process bulk level progression", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to process bulk level progression",
		})
		return
	}

	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	h.logger.Info("Bulk level progression completed",
		zap.Int("total_users", len(userUUIDs)),
		zap.Int("successful", successCount),
		zap.Int("failed", len(userUUIDs)-successCount))

	c.JSON(http.StatusOK, gin.H{
		"results":     results,
		"total_users": len(userUUIDs),
		"successful":  successCount,
		"failed":      len(userUUIDs) - successCount,
	})
}

// ProcessSingleLevelProgression processes level progression for a single user (admin only)
// @Summary Process single level progression
// @Description Processes level progression for a single user
// @Tags Cashback
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id body string true "User ID"
// @Success 200 {object} dto.LevelProgressionResult
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/cashback/level-progression [post]
func (h *CashbackHandler) ProcessSingleLevelProgression(c *gin.Context) {
	var request struct {
		UserID string `json:"user_id" validate:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
		})
		return
	}

	// Parse user ID
	userUUID, err := uuid.Parse(request.UserID)
	if err != nil {
		h.logger.Error("Invalid user ID format", zap.String("user_id", request.UserID), zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Invalid user ID: %s", request.UserID),
		})
		return
	}

	h.logger.Info("Processing single level progression",
		zap.String("user_id", userUUID.String()))

	result, err := h.cashbackService.ProcessSingleLevelProgression(c.Request.Context(), userUUID)
	if err != nil {
		h.logger.Error("Failed to process single level progression", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to process single level progression",
		})
		return
	}

	h.logger.Info("Single level progression completed",
		zap.String("user_id", userUUID.String()),
		zap.Bool("success", result.Success),
		zap.Int("new_level", result.NewLevel))

	c.JSON(http.StatusOK, result)
}

// GetGlobalRakebackOverride returns the current global rakeback override configuration
// @Summary Get global rakeback override
// @Description Returns the current global rakeback override configuration (Happy Hour mode)
// @Tags Cashback
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.GlobalRakebackOverride
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/cashback/global-override [get]
func (h *CashbackHandler) GetGlobalRakebackOverride(c *gin.Context) {
	h.logger.Info("Getting global rakeback override configuration")

	override, err := h.cashbackService.GetGlobalRakebackOverride(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get global rakeback override", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get global rakeback override",
		})
		return
	}

	c.JSON(http.StatusOK, override)
}

// UpdateGlobalRakebackOverride updates the global rakeback override configuration
// @Summary Update global rakeback override
// @Description Updates the global rakeback override configuration (Happy Hour mode) - Admin only
// @Tags Cashback
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.GlobalRakebackOverrideRequest true "Global rakeback override configuration"
// @Success 200 {object} dto.GlobalRakebackOverrideResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/cashback/global-override [put]
func (h *CashbackHandler) UpdateGlobalRakebackOverride(c *gin.Context) {
	// Get admin user ID from context
	adminUserIDStr := c.GetString("user-id")
	if adminUserIDStr == "" {
		h.logger.Error("Admin user ID not found in context")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "Admin user not authenticated",
		})
		return
	}

	adminUserID, err := uuid.Parse(adminUserIDStr)
	if err != nil {
		h.logger.Error("Invalid admin user ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid admin user ID",
		})
		return
	}

	var request dto.GlobalRakebackOverrideRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Failed to bind global rakeback override request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request format",
		})
		return
	}

	h.logger.Info("Updating global rakeback override",
		zap.String("admin_user_id", adminUserID.String()),
		zap.Bool("is_enabled", request.IsEnabled),
		zap.String("override_percentage", request.OverridePercentage.String()))

	updated, err := h.cashbackService.UpdateGlobalRakebackOverride(c.Request.Context(), adminUserID, request)
	if err != nil {
		h.logger.Error("Failed to update global rakeback override", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update global rakeback override",
		})
		return
	}

	// Prepare response
	var message string
	if updated.IsEnabled {
		message = fmt.Sprintf("Global rakeback override enabled! All users now receive %s%% rakeback (Happy Hour activated)", updated.OverridePercentage.String())
	} else {
		message = "Global rakeback override disabled. Users will receive VIP tier-based rakeback."
	}

	response := dto.GlobalRakebackOverrideResponse{
		IsEnabled:          updated.IsEnabled,
		OverridePercentage: updated.OverridePercentage,
		EnabledBy:          updated.EnabledBy,
		EnabledAt:          updated.EnabledAt,
		DisabledBy:         updated.DisabledBy,
		DisabledAt:         updated.DisabledAt,
		Message:            message,
	}

	h.logger.Info("Global rakeback override updated successfully",
		zap.Bool("is_enabled", updated.IsEnabled),
		zap.String("override_percentage", updated.OverridePercentage.String()),
		zap.String("admin_user_id", adminUserID.String()))

	c.JSON(http.StatusOK, response)
}

// CreateRakebackSchedule creates a new scheduled rakeback event
// @Summary Create rakeback schedule
// @Description Creates a new scheduled rakeback event (Happy Hour, Weekend Boost, etc.) - Admin only
// @Tags Cashback
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateRakebackScheduleRequest true "Rakeback schedule configuration"
// @Success 201 {object} dto.RakebackScheduleResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/cashback/schedules [post]
func (h *CashbackHandler) CreateRakebackSchedule(c *gin.Context) {
	adminUserIDStr := c.GetString("user-id")
	if adminUserIDStr == "" {
		h.logger.Error("Admin user ID not found in context")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "Admin user not authenticated",
		})
		return
	}

	adminUserID, err := uuid.Parse(adminUserIDStr)
	if err != nil {
		h.logger.Error("Invalid admin user ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid admin user ID",
		})
		return
	}

	var request dto.CreateRakebackScheduleRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Failed to bind rakeback schedule request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request format",
		})
		return
	}

	h.logger.Info("Creating rakeback schedule",
		zap.String("admin_user_id", adminUserID.String()),
		zap.String("name", request.Name))

	schedule, err := h.cashbackService.CreateRakebackSchedule(c.Request.Context(), adminUserID, request)
	if err != nil {
		h.logger.Error("Failed to create rakeback schedule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	h.logger.Info("Rakeback schedule created successfully",
		zap.String("schedule_id", schedule.ID.String()),
		zap.String("name", schedule.Name))

	c.JSON(http.StatusCreated, schedule)
}

// ListRakebackSchedules lists all rakeback schedules
// @Summary List rakeback schedules
// @Description Lists all rakeback schedules with optional status filter and pagination - Admin only
// @Tags Cashback
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status (scheduled, active, completed, cancelled)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.ListRakebackSchedulesResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/cashback/schedules [get]
func (h *CashbackHandler) ListRakebackSchedules(c *gin.Context) {
	status := c.DefaultQuery("status", "")
	page := 1
	pageSize := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	h.logger.Info("Listing rakeback schedules",
		zap.String("status", status),
		zap.Int("page", page),
		zap.Int("page_size", pageSize))

	schedules, err := h.cashbackService.ListRakebackSchedules(c.Request.Context(), status, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list rakeback schedules", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list rakeback schedules",
		})
		return
	}

	c.JSON(http.StatusOK, schedules)
}

// GetRakebackSchedule retrieves a single rakeback schedule
// @Summary Get rakeback schedule
// @Description Retrieves details of a specific rakeback schedule - Admin only
// @Tags Cashback
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Schedule ID"
// @Success 200 {object} dto.RakebackScheduleResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/cashback/schedules/{id} [get]
func (h *CashbackHandler) GetRakebackSchedule(c *gin.Context) {
	scheduleIDStr := c.Param("id")
	scheduleID, err := uuid.Parse(scheduleIDStr)
	if err != nil {
		h.logger.Error("Invalid schedule ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid schedule ID",
		})
		return
	}

	h.logger.Info("Getting rakeback schedule", zap.String("schedule_id", scheduleID.String()))

	schedule, err := h.cashbackService.GetRakebackSchedule(c.Request.Context(), scheduleID)
	if err != nil {
		h.logger.Error("Failed to get rakeback schedule", zap.Error(err))
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Code:    http.StatusNotFound,
			Message: "Rakeback schedule not found",
		})
		return
	}

	c.JSON(http.StatusOK, schedule)
}

// UpdateRakebackSchedule updates an existing rakeback schedule
// @Summary Update rakeback schedule
// @Description Updates an existing scheduled rakeback event (only if not yet active) - Admin only
// @Tags Cashback
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Schedule ID"
// @Param request body dto.UpdateRakebackScheduleRequest true "Updated schedule configuration"
// @Success 200 {object} dto.RakebackScheduleResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/cashback/schedules/{id} [put]
func (h *CashbackHandler) UpdateRakebackSchedule(c *gin.Context) {
	scheduleIDStr := c.Param("id")
	scheduleID, err := uuid.Parse(scheduleIDStr)
	if err != nil {
		h.logger.Error("Invalid schedule ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid schedule ID",
		})
		return
	}

	var request dto.UpdateRakebackScheduleRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Failed to bind update request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request format",
		})
		return
	}

	h.logger.Info("Updating rakeback schedule", zap.String("schedule_id", scheduleID.String()))

	schedule, err := h.cashbackService.UpdateRakebackSchedule(c.Request.Context(), scheduleID, request)
	if err != nil {
		h.logger.Error("Failed to update rakeback schedule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, schedule)
}

// DeleteRakebackSchedule cancels a rakeback schedule
// @Summary Delete rakeback schedule
// @Description Cancels a scheduled rakeback event (only if not yet active) - Admin only
// @Tags Cashback
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Schedule ID"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/cashback/schedules/{id} [delete]
func (h *CashbackHandler) DeleteRakebackSchedule(c *gin.Context) {
	scheduleIDStr := c.Param("id")
	scheduleID, err := uuid.Parse(scheduleIDStr)
	if err != nil {
		h.logger.Error("Invalid schedule ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid schedule ID",
		})
		return
	}

	h.logger.Info("Deleting rakeback schedule", zap.String("schedule_id", scheduleID.String()))

	err = h.cashbackService.DeleteRakebackSchedule(c.Request.Context(), scheduleID)
	if err != nil {
		h.logger.Error("Failed to delete rakeback schedule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Rakeback schedule cancelled successfully",
	})
}
