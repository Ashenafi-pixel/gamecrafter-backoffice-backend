package bet

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/response"
	"go.uber.org/zap"
)

// CreateLootBox creates a new loot box.
//
//	@Summary		CreateLootBox
//	@Description	CreateLootBox allows admin to create a new loot box
//	@Tags			bet
//	@Param			Authorization	header	string	true	"Bearer <token>"
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.CreateLootBoxReq	true	"create loot box Request"
//	@Success		201	{object}	dto.CreateLootBoxRes
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		500	{object}	response.ErrorResponse
//	@Router			/api/admin/lootboxes [post]
func (b *bet) CreateLootBox(c *gin.Context) {
	var req dto.CreateLootBoxReq
	if err := c.ShouldBind(&req); err != nil {
		b.log.Error("failed to bind request", zap.Error(err))
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid request body")
		_ = c.Error(err)
		return
	}

	resp, err := b.betModule.CreateLootBox(c.Request.Context(), req)
	if err != nil {
		b.log.Error("failed to create lootbox", zap.Error(err), zap.Any("req", req))
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusCreated, resp)
}

// UpdateLootBox updates an existing loot box.
//
//	@Summary		UpdateLootBox
//	@Description	UpdateLootBox allows admin to update an existing loot box
//	@Tags			bet
//	@Param			Authorization	header	string	true	"Bearer <token>"
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.UpdateLootBoxReq	true	"update loot box Request"
//	@Success		200	{object}	dto.UpdateLootBoxRes
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		500	{object}	response.ErrorResponse
//	@Router			/api/admin/lootboxes [put]
func (b *bet) UpdateLootBox(c *gin.Context) {
	var req dto.UpdateLootBoxReq
	if err := c.ShouldBind(&req); err != nil {
		b.log.Error("failed to bind request", zap.Error(err))
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid request body")
		_ = c.Error(err)
		return
	}

	resp, err := b.betModule.UpdateLootBox(c.Request.Context(), req)
	if err != nil {
		b.log.Error("failed to update lootbox", zap.Error(err), zap.Any("req", req))
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// DeleteLootBox deletes a loot box by its ID.
//
//	@Summary		DeleteLootBox
//	@Description	DeleteLootBox allows admin to delete a loot box by its ID
//	@Tags			bet
//	@Param			Authorization	header		string	true	"Bearer <token>"
//	@Param			id				path		string	true	"lootbox id"
//	@Success		200				{object}	dto.DeleteLootBoxRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		500				{object}	response.ErrorResponse
//	@Router			/api/admin/lootboxes/{id} [delete]
func (b *bet) DeleteLootBox(c *gin.Context) {
	id, ok := c.Params.Get("id")
	if !ok {
		err := errors.ErrInvalidUserInput.Wrap(nil, "lootbox id is required")
		b.log.Error("failed to get lootbox id from params", zap.Error(err))
		_ = c.Error(err)
		return
	}

	// Convert id to uuid
	lootBoxID, err := uuid.Parse(id)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid lootbox id format")
		b.log.Error("failed to parse lootbox id", zap.Error(err), zap.String("id", id))
		_ = c.Error(err)
		return
	}

	resp, err := b.betModule.DeleteLootBox(c.Request.Context(), lootBoxID)
	if err != nil {
		b.log.Error("failed to delete lootbox", zap.Error(err))
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetLootBox retrieves all loot boxes.
//
//	@Summary		GetLootBox
//	@Description	GetLootBox retrieves all loot boxes
//	@Tags			bet
//	@Param			Authorization	header	string	true	"Bearer <token>"
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	[]dto.LootBox
//	@Failure		500	{object}	response.ErrorResponse
//	@Router			/api/admin/lootboxes [get]
func (b *bet) GetLootBox(c *gin.Context) {
	resp, err := b.betModule.GetLootBox(c.Request.Context())
	if err != nil {
		b.log.Error("failed to get lootbox", zap.Error(err))
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// PlaceLootBoxBet places a bet on a loot box.
//
//	@Summary		PlaceLootBoxBet
//	@Description	PlaceLootBoxBet allows users to place a bet on a loot box
//	@Tags			bet
//	@Param			Authorization	header	string	true	"Bearer <token>"
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.PlaceLootBoxBetReq	true	"place loot box bet Request"
//	@Success		200	{object}	[]dto.PlaceLootBoxResp
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		500	{object}	response.ErrorResponse
//	@Router			/api/bet/lootboxes [get]
func (b *bet) PlaceLootBoxBet(c *gin.Context) {

	userID := c.GetString("user-id")

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := b.betModule.PlaceLootBoxBet(c.Request.Context(), userIDParsed)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// SelectLootBox selects a loot box for the user.
//
//	@Summary		SelectLootBox
//	@Description	SelectLootBox allows users to select a loot box
//	@Tags			bet
//	@Param			Authorization	header	string	true	"Bearer <token>"
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.PlaceLootBoxResp	true	"select loot box Request"
//	@Success		200	{object}	dto.LootBoxBetResp
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		500	{object}	response.ErrorResponse
//	@Router			/api/bet/lootboxes/select [post]
func (b *bet) SelectLootBox(c *gin.Context) {
	var req dto.PlaceLootBoxResp
	if err := c.ShouldBind(&req); err != nil {
		b.log.Error("failed to bind request", zap.Error(err))
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid request body")
		_ = c.Error(err)
		return
	}

	userID := c.GetString("user-id")

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := b.betModule.SelectLootBox(c.Request.Context(), req, userIDParsed)
	if err != nil {
		b.log.Error("failed to select lootbox", zap.Error(err), zap.Any("req", req))
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}
