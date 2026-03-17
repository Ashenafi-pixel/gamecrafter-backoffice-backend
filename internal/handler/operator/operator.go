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

// CreateOperator creates a new operator.
func (h *OperatorHandler) CreateOperator(c *gin.Context) {
	var req dto.CreateOperatorReq
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errors.ErrInvalidUserInput.Wrap(err, err.Error()))
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
		_ = c.Error(errors.ErrInvalidUserInput.Wrap(err, err.Error()))
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
		_ = c.Error(errors.ErrInvalidUserInput.Wrap(err, err.Error()))
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
		_ = c.Error(errors.ErrInvalidUserInput.Wrap(err, err.Error()))
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
		_ = c.Error(errors.ErrInvalidUserInput.Wrap(err, err.Error()))
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

func parseOperatorID(c *gin.Context) (int32, bool) {
	idStr := c.Param("id")
	if idStr == "" {
		_ = c.Error(errors.ErrInvalidUserInput.New("operator id is required"))
		return 0, false
	}

	id64, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil || id64 <= 0 {
		_ = c.Error(errors.ErrInvalidUserInput.New("invalid operator id"))
		return 0, false
	}

	return int32(id64), true
}

