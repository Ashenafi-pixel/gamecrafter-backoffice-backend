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

