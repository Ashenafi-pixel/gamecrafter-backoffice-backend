package balance

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/response"
	"github.com/tucanbit/internal/handler"
	"github.com/tucanbit/internal/module"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type balance struct {
	balanceModule module.Balance
	log           *zap.Logger
}

func Init(balanceModule module.Balance, log *zap.Logger) handler.Balance {
	return &balance{
		balanceModule: balanceModule,
		log:           log,
	}
}

// Get User Balance.
//
// @Summary		GetUserBalances
// @Description	get user balance
// @Tags			Balance
// @Param			Authorization	header	string	true	"Bearer <token> "
// @Accept			json
// @Produce		json
// @Success		200	{object}	[]dto.Balance
// @Failure		401	{object}	response.ErrorResponse
// @Router			/api/balance [get]
func (b *balance) GetUserBalances(c *gin.Context) {
	userID := c.GetString("user-id")
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	balances, err := b.balanceModule.GetBalanceByUserID(c, userIDParsed)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, balances)
}

// ExchangeBalance Exchange User Balance.
//
//	@Summary		ExchangeBalance
//	@Description	Exchange Balance user balance from one currency to another currency
//	@Tags			Balance
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			exchangeBalanceReq	body		dto.ExchangeBalanceReq	true	"exchange user balance Request"
//	@Success		200					{object}	[]dto.ExchangeBalanceRes
//	@Failure		401					{object}	response.ErrorResponse
//	@Router			/api/balance/exchange [post]
func (b *balance) ExchangeBalance(c *gin.Context) {
	userID := c.GetString("user-id")
	var exchangeBalanceReq dto.ExchangeBalanceReq
	if err := c.ShouldBind(&exchangeBalanceReq); err != nil {
		b.log.Warn(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	exchangeBalanceReq.UserID = userIDParsed
	exchangedBalanceRes, err := b.balanceModule.Exchange(c, exchangeBalanceReq)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, exchangedBalanceRes)
}

// ManualFunding add or remove  fund manually to or from User's Balance.
//
//	@Summary		ManualFunding
//	@Description	ManualFunding add or remove  fund manually to or from User's Balance.
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			fundReq	body		dto.ManualFundReq	true	"add or remove fund Request"
//	@Success		200		{object}	dto.ManualFundRes
//	@Failure		401		{object}	response.ErrorResponse
//	@Router			/api/admin/players/funding [post]
func (b *balance) ManualFunding(c *gin.Context) {

	userID := c.GetString("user-id")
	userIDParsed, err := uuid.Parse(userID)

	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	var fundReq dto.ManualFundReq
	if err := c.ShouldBind(&fundReq); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	if fundReq.Type == constant.ADD_FUND {
		fundReq.AdminID = userIDParsed
		manualFund, err := b.balanceModule.AddManualFunds(c, fundReq)
		if err != nil {
			_ = c.Error(err)
			return
		}
		response.SendSuccessResponse(c, http.StatusCreated, manualFund)
	} else if fundReq.Type == constant.REMOVE_FUND {
		fundReq.AdminID = userIDParsed
		manualFund, err := b.balanceModule.RemoveFundManualy(c, fundReq)

		if err != nil {
			_ = c.Error(err)
			return
		}
		response.SendSuccessResponse(c, http.StatusCreated, manualFund)

	}
	if fundReq.Type != constant.ADD_FUND && fundReq.Type != constant.REMOVE_FUND {
		err = fmt.Errorf("invalid fund type , only add_fund or remove_fund is allowed")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
}

// GetManualFundLogs get manual funds logs.
//
//	@Summary		GetManualFundLogs
//	@Description	Retrieve manual funds for admin based on various query parameters like username, start date, etc.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization				header		string	true	"Bearer <token> "
//	@Param			filter_start_date			query		string	false	"filter_start_date (string, optional)"
//	@Param			filter_end_date				query		string	false	"filter_end_date (string , optional)"
//	@Param			filter_type					query		decimal	false	"filter_start_amount (decimal, optional)"
//	@Param			filter_customer_username	query		string	false	"filter_customer_username (string (deposit,withdrawal), optional)"
//	@Param			filter_customer_email		query		string	false	"filter_customer_email (string (deposit,withdrawal), optional)"
//	@Param			filter_customer_phone		query		string	false	"filter_customer_phone (string (deposit,withdrawal), optional)"
//	@Param			filter_admin_username		query		string	false	"filter_customer_username (string (deposit,withdrawal), optional)"
//	@Param			filter_admin_email			query		string	false	"filter_customer_email (string (deposit,withdrawal), optional)"
//	@Param			filter_admin_phone			query		string	false	"filter_customer_phone  (string (deposit,withdrawal), optional)"
//	@Param			sort_amount					query		string	false	"sort_amount (string , optional)"
//	@Param			sort_date					query		string	false	"filter_start_date (string , optional)"
//	@Param			sort_amount					query		string	false	"sort_amount (string , optional)"
//	@Param			page						query		string	true	"page type (required)"
//	@Param			per_page					query		string	true	"per-page type (required)"
//	@Success		200							{object}	dto.GetManualFundReq
//	@Failure		400							{object}	response.ErrorResponse
//	@Failure		401							{object}	response.ErrorResponse
//	@Router			/api/admin/balance/log/funds [get]
func (b *balance) GetManualFundLogs(c *gin.Context) {

	var req dto.GetManualFundReq

	var err error

	if err := c.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	if req.PerPage == 0 || req.Page == 0 {
		err := fmt.Errorf("invalid page and per_page query can not be empty")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := b.balanceModule.GetManualFundLogs(c, req)

	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusCreated, resp)
}

// CreditWallet credits a user's wallet after payment confirmation.
//
//	@Summary		Credit a user's wallet
//	@Description	Credits a user's wallet after payment confirmation from the payment microservice. Requires a unique payment reference and a valid service secret.
//	@Tags			Wallet, Payment
//	@Accept			json
//	@Produce		json
//	@Param			X-Service-Secret	header		string				true	"Service-to-service secret"
//	@Param			request				body		dto.CreditWalletReq	true	"Credit wallet request"
//	@Success		200					{object}	dto.CreditWalletRes
//	@Failure		400					{object}	response.ErrorResponse
//	@Failure		401					{object}	response.ErrorResponse
//	@Failure		500					{object}	response.ErrorResponse
//	@Router			/api/wallet/credit [patch]
func (b *balance) CreditWallet(c *gin.Context) {
	secret := c.GetHeader("X-Service-Secret")
	validSecret := viper.GetString("paymentMs.secret")
	if secret == "" || secret != validSecret {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized: invalid or missing service secret"})
		return
	}

	var req dto.CreditWalletReq
	if err := c.ShouldBindJSON(&req); err != nil {
		b.log.Warn(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := b.balanceModule.CreditWallet(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}
