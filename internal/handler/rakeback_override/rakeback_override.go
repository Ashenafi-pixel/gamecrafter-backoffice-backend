package rakeback_override

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/module/rakeback_override"
	"go.uber.org/zap"
)

type RakebackOverrideHandler struct {
	rakebackOverrideModule rakeback_override.RakebackOverride
	log                    *zap.Logger
}

func NewRakebackOverrideHandler(rakebackOverrideModule rakeback_override.RakebackOverride, log *zap.Logger) *RakebackOverrideHandler {
	return &RakebackOverrideHandler{
		rakebackOverrideModule: rakebackOverrideModule,
		log:                    log,
	}
}

// GetActiveOverride retrieves the currently active global rakeback override
// @Summary Get active global rakeback override
// @Description Retrieves the currently active global rakeback override configuration
// @Tags rakeback-override
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Active override or null"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/admin/rakeback-override/active [get]
func (h *RakebackOverrideHandler) GetActiveOverride(ctx *gin.Context) {
	override, err := h.rakebackOverrideModule.GetActiveOverride(ctx.Request.Context())
	if err != nil {
		h.log.Error("Failed to get active rakeback override", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get active rakeback override",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    override,
	})
}

// GetOverride retrieves the most recent global rakeback override (active or not)
// @Summary Get global rakeback override
// @Description Retrieves the most recent global rakeback override configuration
// @Tags rakeback-override
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Override configuration or null"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/admin/rakeback-override [get]
func (h *RakebackOverrideHandler) GetOverride(ctx *gin.Context) {
	override, err := h.rakebackOverrideModule.GetOverride(ctx.Request.Context())
	if err != nil {
		h.log.Error("Failed to get rakeback override", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get rakeback override",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    override,
	})
}

// CreateOrUpdateOverride creates or updates the global rakeback override
// @Summary Create or update global rakeback override
// @Description Creates a new global rakeback override or updates an existing one
// @Tags rakeback-override
// @Accept json
// @Produce json
// @Param request body dto.CreateOrUpdateRakebackOverrideReq true "Rakeback override configuration"
// @Success 200 {object} map[string]interface{} "Created/updated override"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/admin/rakeback-override [post]
func (h *RakebackOverrideHandler) CreateOrUpdateOverride(ctx *gin.Context) {
	var req dto.CreateOrUpdateRakebackOverrideReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	// Get admin ID from context
	adminUserID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Unauthorized",
		})
		return
	}

	adminUUID, ok := adminUserID.(uuid.UUID)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	override, err := h.rakebackOverrideModule.CreateOrUpdateOverride(ctx.Request.Context(), req, adminUUID)
	if err != nil {
		h.log.Error("Failed to create/update rakeback override", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create/update rakeback override",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Rakeback override created/updated successfully",
		"data":    override,
	})
}

// ToggleOverride enables or disables the global rakeback override
// @Summary Toggle global rakeback override
// @Description Enables or disables the global rakeback override
// @Tags rakeback-override
// @Accept json
// @Produce json
// @Param request body map[string]bool true "Toggle request with is_active field"
// @Success 200 {object} map[string]interface{} "Success response"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/admin/rakeback-override/toggle [patch]
func (h *RakebackOverrideHandler) ToggleOverride(ctx *gin.Context) {
	var req struct {
		IsActive bool `json:"is_active"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	// Get admin ID from context
	adminUserID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Unauthorized",
		})
		return
	}

	adminUUID, ok := adminUserID.(uuid.UUID)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	err := h.rakebackOverrideModule.ToggleOverride(ctx.Request.Context(), req.IsActive, adminUUID)
	if err != nil {
		h.log.Error("Failed to toggle rakeback override", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to toggle rakeback override",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Rakeback override toggled successfully",
	})
}

