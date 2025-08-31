package exchange

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/response"
	"github.com/tucanbit/internal/handler"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

type exchange struct {
	exchangeModule module.Exchange
	log            *zap.Logger
}

func Init(exchangeModule module.Exchange, log *zap.Logger) handler.Exchange {
	return &exchange{
		exchangeModule: exchangeModule,
		log:            log,
	}
}

// GetExcahnge Get Exchange Rate.
//
//	@Summary		GetExchange
//	@Description	Exchange get exchange rate from one currency to other currency
//	@Tags			Exchange
//	@Accept			json
//	@Produce		json
//	@Param			exchangeReq	body		dto.ExchangeReq	true	"get Exchange Rate Request"
//	@Success		200			{object}	[]dto.ExchangeRes
//	@Failure		400			{object}	response.ErrorResponse
//	@Router			/api/balance/exchange [post]
func (e *exchange) GetExcahnge(c *gin.Context) {
	var exchangeReq dto.ExchangeReq
	if err := c.ShouldBind(&exchangeReq); err != nil {
		e.log.Warn(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	exR, err := e.exchangeModule.GetExchange(c, exchangeReq)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, exR)
}

// GetAvailableCurrencies  get get available currencies.
//
//	@Summary		GetAvailableCurrencies
//	@Description	GetAvailableCurrencies allow user  to get get available currencies.
//	@Tags			Exchange
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Param			page			query	string	true	"page type (required)"
//	@Param			per_page		query	string	true	"per_page type (required)"
//	@Produce		json
//	@Success		200	{object}	dto.GetCurrencyReq
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/currencies [get]
func (e *exchange) GetAvailableCurrencies(c *gin.Context) {

	var req dto.GetRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := e.exchangeModule.GetAvailableCurrencies(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)

}
