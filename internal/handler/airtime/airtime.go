package airtime

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/response"
	"github.com/tucanbit/internal/handler"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

type airtime struct {
	log           *zap.Logger
	airtimeModule module.AirtimeProvider
}

func Init(log *zap.Logger, airtimeModule module.AirtimeProvider) handler.AirtimeProvider {
	return &airtime{
		log:           log,
		airtimeModule: airtimeModule,
	}
}

// RefereshAirtimeUtilities get manual funds logs.
//
//	@Summary		RefereshAirtimeUtilities
//	@Description	Refresh available airtime utilities from airtime providers (fetch latest utilities)
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token> "
//	@Success		200				{object}	[]dto.AirtimeUtility
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admin/airtime/refresh [get]
func (a *airtime) RefereshAirtimeUtilities(c *gin.Context) {
	resp, err := a.airtimeModule.RefereshUtilies(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// RefereshAirtimeUtilities get manual funds logs.
//
//	@Summary		RefereshAirtimeUtilities
//	@Description	Fetch available airtime utilities from local database
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			page			query		string	true	"page type (required)"
//	@Param			per_page		query		string	true	"per_page type (required)"
//	@Param			Authorization	header		string	true	"Bearer <token> "
//	@Success		200				{object}	dto.GetAirtimeUtilitiesResp
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admin/airtime/utilities [get]
func (a *airtime) GetAvailableAirtime(c *gin.Context) {
	var req dto.GetRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := a.airtimeModule.GetAvailableAirtimeUtilities(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// UpdateAirtimeStatus .
//
//	@Summary		UpdateAirtimeStatus
//	@Description	UpdateAirtimeStatus update availabilities of airtime
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			addFundReq	body		dto.UpdateAirtimeStatusReq	true	"update airtime availability Request"
//	@Success		200			{object}	dto.UpdateAirtimeStatusResp
//	@Failure		401			{object}	response.ErrorResponse
//	@Router			/api/admin/airtime [put]
func (a *airtime) UpdateAirtimeStatus(c *gin.Context) {
	var req dto.UpdateAirtimeStatusReq

	if err := c.ShouldBind(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
	}

	resp, err := a.airtimeModule.UpdateAirtimeStatus(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// UpdateAirtimeUtilityPrice .
//
//	@Summary		UpdateAirtimeUtilityPrice
//	@Description	UpdateAirtimeUtilityPrice update price of airtime
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			addFundReq	body		dto.UpdateAirtimeUtilityPriceReq	true	"update airtime utilities price Request"
//	@Success		200			{object}	dto.UpdateAirtimeUtilityPriceRes
//	@Failure		401			{object}	response.ErrorResponse
//	@Router			/api/admin/airtime/price [put]
func (a *airtime) UpdateAirtimeUtilityPrice(c *gin.Context) {
	var req dto.UpdateAirtimeUtilityPriceReq

	if err := c.ShouldBind(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := a.airtimeModule.UpdateUtilityPrice(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// ClaimPoints .
//
//	@Summary		ClaimPoints
//	@Description	ClaimPoints points to airtime
//	@Tags			Airtime
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			addFundReq	body		dto.ClaimPointsReq	true	"claim point to airtime"
//	@Success		200			{object}	dto.ClaimPointsResp
//	@Failure		401			{object}	response.ErrorResponse
//	@Router			/api/airtime/claim [post]
func (a *airtime) ClaimPoints(c *gin.Context) {
	var req dto.ClaimPointsReq
	if err := c.ShouldBind(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	userID := c.GetString("user-id")

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		a.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	req.UserID = userIDParsed
	resp, err := a.airtimeModule.ClaimPoints(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)

}

// GetActiveAvailableAirtime get available airtime utilities.
//
//	@Summary		GetActiveAvailableAirtime
//	@Description	Retrieve active airtime utilities.
//	@Tags			Airtime
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token> "
//	@Param			page			query		string	true	"page type (required)"
//	@Param			per_page		query		string	true	"per-page type (required)"
//	@Success		200				{object}	dto.GetAirtimeUtilitiesResp
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/airtime/active/utilities [get]
func (a *airtime) GetActiveAvailableAirtime(c *gin.Context) {
	var req dto.GetRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		_ = c.Error(err)
		return
	}

	resp, err := a.airtimeModule.GetActiveAvailableAirtime(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// RefereshAirtimeUtilities airtime transactions
//
//	@Summary		RefereshAirtimeUtilities
//	@Description	Fetch  airtime transactions
//	@Tags			Airtime
//	@Accept			json
//	@Produce		json
//	@Param			page			query		string	true	"page type (required)"
//	@Param			per_page		query		string	true	"per_page type (required)"
//	@Param			Authorization	header		string	true	"Bearer <token> "
//	@Success		200				{object}	dto.GetAirtimeTransactionsResp
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/airtime/transactions [get]
func (a *airtime) GetUserAirtimeTransactions(c *gin.Context) {
	var req dto.GetRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	userID := c.GetString("user-id")

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		a.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := a.airtimeModule.GetUserAirtimeTransactions(c, req, userIDParsed)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetAllAirtimeUtilitiesTransactions get manual funds logs.
//
//	@Summary		GetAllAirtimeUtilitiesTransactions
//	@Description	Fetch airtime transactions
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			page			query		string	true	"page type (required)"
//	@Param			per_page		query		string	true	"per_page type (required)"
//	@Param			Authorization	header		string	true	"Bearer <token> "
//	@Success		200				{object}	dto.GetAirtimeTransactionsResp
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admin/airtime/transactions [get]
func (a *airtime) GetAllAirtimeUtilitiesTransactions(c *gin.Context) {
	var req dto.GetRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	userID := c.GetString("user-id")

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		a.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := a.airtimeModule.GetUserAirtimeTransactions(c, req, userIDParsed)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// UpdateAirtimeAmount .
//
//	@Summary		UpdateAirtimeAmount
//	@Description	UpdateAirtimeAmount update airtime amount
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.UpdateAirtimeUtilityPriceReq	true	"update airtime amount req"
//	@Success		200	{object}	dto.UpdateAirtimeUtilityPriceRes
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/airtime/amount [put]
func (a *airtime) UpdateAirtimeAmount(c *gin.Context) {
	var req dto.UpdateAirtimeUtilityPriceReq

	if err := c.ShouldBind(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := a.airtimeModule.UpdateUtilityPrice(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetAirtimeUtilitiesStats get airtime utilities stats.
//
//	@Summary		GetAirtimeUtilitiesStats
//	@Description	Fetch airtime utilities stats
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token> "
//	@Success		200				{object}	dto.AirtimeUtilitiesStats
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admin/airtime/stats [get]
func (a *airtime) GetAirtimeUtilitiesStats(c *gin.Context) {
	resp, err := a.airtimeModule.GetAirtimeUtilitiesStats(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}
