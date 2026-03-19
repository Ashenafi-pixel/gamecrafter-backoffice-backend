package operator

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/response"
	"github.com/tucanbit/internal/module"
)

type OperatorHandler struct {
	operatorModule module.Operator
}

func NewOperatorHandler(operatorModule module.Operator) *OperatorHandler {
	return &OperatorHandler{
		operatorModule: operatorModule,
	}
}

// CreateOperatorCredential creates API credentials for an operator; returns api_key and signing_key once.
func (h *OperatorHandler) CreateOperatorCredential(c *gin.Context) {
	operatorID, ok := parseOperatorID(c)
	if !ok {
		return
	}

	cred, err := h.operatorModule.CreateOperatorCredential(c.Request.Context(), operatorID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusCreated, cred)
}

// RotateOperatorCredential rotates the signing key; returns new signing_key once.
func (h *OperatorHandler) RotateOperatorCredential(c *gin.Context) {
	operatorID, ok := parseOperatorID(c)
	if !ok {
		return
	}
	credIDStr := c.Param("credentialId")
	credID, err := strconv.ParseInt(credIDStr, 10, 32)
	if err != nil {
		_ = c.Error(errors.ErrInvalidUserInput.Wrap(err, "invalid credential ID format"))
		return
	}

	res, err := h.operatorModule.RotateOperatorCredential(c.Request.Context(), operatorID, int32(credID))
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, res)
}

// AssignAllGamesToOperator assigns all games in the system to this operator.
func (h *OperatorHandler) AssignAllGamesToOperator(c *gin.Context) {
	operatorID, ok := parseOperatorID(c)
	if !ok {
		return
	}

	if err := h.operatorModule.AssignAllGamesToOperator(c.Request.Context(), operatorID); err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message": "all games assigned to operator",
	})
}

// CreateOperator creates a new operator.
func (h *OperatorHandler) CreateOperator(c *gin.Context) {
	var req dto.CreateOperatorReq
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errors.ErrInvalidUserInput.Wrap(err, "invalid request body"))
		return
	}

	op, err := h.operatorModule.CreateOperator(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusCreated, op)
}

// GetOperatorByID returns a single operator.
func (h *OperatorHandler) GetOperatorByID(c *gin.Context) {
	operatorID, ok := parseOperatorID(c)
	if !ok {
		return
	}

	op, err := h.operatorModule.GetOperatorByID(c.Request.Context(), operatorID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, op)
}

// GetOperators returns a paginated list of operators.
func (h *OperatorHandler) GetOperators(c *gin.Context) {
	var req dto.GetOperatorsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		_ = c.Error(errors.ErrInvalidUserInput.Wrap(err, "invalid query parameters"))
		return
	}

	res, err := h.operatorModule.GetOperators(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, res)
}

// UpdateOperator updates an existing operator.
func (h *OperatorHandler) UpdateOperator(c *gin.Context) {
	operatorID, ok := parseOperatorID(c)
	if !ok {
		return
	}

	var req dto.UpdateOperatorReq
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errors.ErrInvalidUserInput.Wrap(err, "invalid request body"))
		return
	}
	req.OperatorID = operatorID

	op, err := h.operatorModule.UpdateOperator(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, op)
}

// DeleteOperator deletes an operator.
func (h *OperatorHandler) DeleteOperator(c *gin.Context) {
	operatorID, ok := parseOperatorID(c)
	if !ok {
		return
	}

	if err := h.operatorModule.DeleteOperator(c.Request.Context(), operatorID); err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusNoContent, nil)
}

// ChangeOperatorStatus updates the is_active flag for an operator.
func (h *OperatorHandler) ChangeOperatorStatus(c *gin.Context) {
	operatorID, ok := parseOperatorID(c)
	if !ok {
		return
	}

	var req dto.ChangeOperatorStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errors.ErrInvalidUserInput.Wrap(err, "invalid request body"))
		return
	}

	if err := h.operatorModule.ChangeOperatorStatus(c.Request.Context(), operatorID, req); err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, gin.H{"is_active": req.IsActive})
}

// AssignGamesToOperator assigns individual games to an operator.
func (h *OperatorHandler) AssignGamesToOperator(c *gin.Context) {
	operatorID, ok := parseOperatorID(c)
	if !ok {
		return
	}

	var req dto.AssignOperatorGamesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errors.ErrInvalidUserInput.Wrap(err, "invalid request body"))
		return
	}

	if err := h.operatorModule.AssignGamesToOperator(c.Request.Context(), operatorID, req.GameIDs); err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message":  "games assigned to operator",
		"game_ids": req.GameIDs,
	})
}

// RevokeGamesFromOperator revokes individual games from an operator.
func (h *OperatorHandler) RevokeGamesFromOperator(c *gin.Context) {
	operatorID, ok := parseOperatorID(c)
	if !ok {
		return
	}

	var req dto.AssignOperatorGamesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errors.ErrInvalidUserInput.Wrap(err, "invalid request body"))
		return
	}

	if err := h.operatorModule.RevokeGamesFromOperator(c.Request.Context(), operatorID, req.GameIDs); err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message":  "games revoked from operator",
		"game_ids": req.GameIDs,
	})
}

// AssignProviderToOperator assigns a whole provider to an operator.
func (h *OperatorHandler) AssignProviderToOperator(c *gin.Context) {
	operatorID, ok := parseOperatorID(c)
	if !ok {
		return
	}

	var req dto.AssignOperatorProviderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errors.ErrInvalidUserInput.Wrap(err, "invalid request body"))
		return
	}

	if err := h.operatorModule.AssignProviderToOperator(c.Request.Context(), operatorID, req.ProviderID); err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, gin.H{"message": "provider assigned"})
}

// RevokeProviderFromOperator revokes a provider from an operator.
func (h *OperatorHandler) RevokeProviderFromOperator(c *gin.Context) {
	operatorID, ok := parseOperatorID(c)
	if !ok {
		return
	}

	providerID := c.Param("providerId")
	if providerID == "" {
		_ = c.Error(errors.ErrInvalidUserInput.New("provider_id is required"))
		return
	}

	if err := h.operatorModule.RevokeProviderFromOperator(c.Request.Context(), operatorID, providerID); err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusNoContent, nil)
}

// AddOperatorAllowedOrigin adds an allowed embed origin for an operator.
func (h *OperatorHandler) AddOperatorAllowedOrigin(c *gin.Context) {
	operatorID, ok := parseOperatorID(c)
	if !ok {
		return
	}

	var req dto.AddOperatorAllowedOriginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errors.ErrInvalidUserInput.Wrap(err, "invalid request body"))
		return
	}

	res, err := h.operatorModule.AddOperatorAllowedOrigin(c.Request.Context(), operatorID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusCreated, res)
}

// RemoveOperatorAllowedOrigin removes an allowed origin for an operator by origin ID.
func (h *OperatorHandler) RemoveOperatorAllowedOrigin(c *gin.Context) {
	operatorID, ok := parseOperatorID(c)
	if !ok {
		return
	}

	originIDStr := c.Param("originId")
	originID, err := strconv.ParseInt(originIDStr, 10, 32)
	if err != nil || originID <= 0 {
		_ = c.Error(errors.ErrInvalidUserInput.Wrap(err, "invalid origin ID format"))
		return
	}

	if err := h.operatorModule.RemoveOperatorAllowedOrigin(c.Request.Context(), operatorID, int32(originID)); err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusNoContent, nil)
}

// ListOperatorAllowedOrigins lists allowed origins for an operator.
func (h *OperatorHandler) ListOperatorAllowedOrigins(c *gin.Context) {
	operatorID, ok := parseOperatorID(c)
	if !ok {
		return
	}

	res, err := h.operatorModule.ListOperatorAllowedOrigins(c.Request.Context(), operatorID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, res)
}

// GetOperatorFeatureFlags returns feature flags for an operator.
func (h *OperatorHandler) GetOperatorFeatureFlags(c *gin.Context) {
	operatorID, ok := parseOperatorID(c)
	if !ok {
		return
	}

	res, err := h.operatorModule.GetOperatorFeatureFlags(c.Request.Context(), operatorID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, res)
}

// UpdateOperatorFeatureFlags updates feature flags for an operator.
func (h *OperatorHandler) UpdateOperatorFeatureFlags(c *gin.Context) {
	operatorID, ok := parseOperatorID(c)
	if !ok {
		return
	}

	var req dto.UpdateOperatorFeatureFlagsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errors.ErrInvalidUserInput.Wrap(err, "invalid request body"))
		return
	}

	if err := h.operatorModule.UpdateOperatorFeatureFlags(c.Request.Context(), operatorID, req); err != nil {
		_ = c.Error(err)
		return
	}

	flags, err := h.operatorModule.GetOperatorFeatureFlags(c.Request.Context(), operatorID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, flags)
}

// GetOperatorGames lists effective game IDs assigned to an operator (direct + via providers).
func (h *OperatorHandler) GetOperatorGames(c *gin.Context) {
	operatorID, ok := parseOperatorID(c)
	if !ok {
		return
	}

	games, err := h.operatorModule.GetOperatorGames(c.Request.Context(), operatorID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, gin.H{
		"operator_id": operatorID,
		"games":       games,
	})
}

func parseOperatorID(c *gin.Context) (int32, bool) {
	// Route params historically used `:id`, but some newer routes use `:operator_id`.
	// Accept both to avoid handler/route mismatches.
	idStr := c.Param("operator_id")
	if idStr == "" {
		idStr = c.Param("id")
	}
	if idStr == "" {
		_ = c.Error(errors.ErrInvalidUserInput.New("operator_id is required"))
		return 0, false
	}

	id64, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil || id64 <= 0 {
		_ = c.Error(errors.ErrInvalidUserInput.New("invalid operator id"))
		return 0, false
	}

	return int32(id64), true
}

