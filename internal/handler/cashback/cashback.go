package cashback

import (
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
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "User not authenticated",
		})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		h.logger.Error("Invalid user ID type")
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
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "User not authenticated",
		})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		h.logger.Error("Invalid user ID type")
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

	// Implementation would go here
	c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
		Code:    http.StatusNotImplemented,
		Message: "Feature not implemented yet",
	})
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

	// Implementation would go here
	c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
		Code:    http.StatusNotImplemented,
		Message: "Feature not implemented yet",
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
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "User not authenticated",
		})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		h.logger.Error("Invalid user ID type")
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

	// Implementation would go here
	c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
		Code:    http.StatusNotImplemented,
		Message: "Feature not implemented yet",
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
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "User not authenticated",
		})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		h.logger.Error("Invalid user ID type")
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
