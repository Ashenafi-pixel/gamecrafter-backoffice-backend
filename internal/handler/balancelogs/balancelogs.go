package balancelogs

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/response"
	"github.com/tucanbit/internal/handler"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

type balance_logs struct {
	balanceLogsModule module.BalanceLogs
	log               *zap.Logger
}

func Init(balanceLogModule module.BalanceLogs, log *zap.Logger) handler.BalanceLogs {
	return &balance_logs{
		balanceLogsModule: balanceLogModule,
		log:               log,
	}
}

// GetBalanceLogs Getbalance logs.
//
//	@Summary		GetBalanceLogs
//	@Description	Retrieve balance logs based on various query parameters like user ID, start date, etc.
//	@Tags			BalanceLogs
//	@Accept			json
//	@Produce		json
//	@Param			Authorization		header		string	true	"Bearer <token>"
//	@Param			user_id				query		string	false	"User ID (UUID, optional)"
//	@Param			per_page			query		int		true	"Per page (required)"
//	@Param			page				query		int		true	"Page (required)"
//	@Param			offset				query		int		false	"Offset (optional)"
//	@Param			component			query		string	false	"Component (optional)"
//	@Param			operation_group_id	query		string	false	"Operation group ID (UUID, optional)"
//	@Param			operation_type_id	query		string	false	"Operation type ID (UUID, optional)"
//	@Param			start_date			query		string	false	"Start date (optional, format: YYYY-MM-DD)"
//	@Param			end_date			query		string	false	"End date (optional, format: YYYY-MM-DD)"
//	@Param			start_amount		query		number	false	"Start amount (optional, decimal)"
//	@Param			end_amount			query		number	false	"End amount (optional, decimal)"
//	@Success		200					{object}	dto.GetBalanceLogRes
//	@Failure		400					{object}	response.ErrorResponse
//	@Failure		401					{object}	response.ErrorResponse
//	@Router			/api/balance/logs [get]
func (bl *balance_logs) GetBalanceLogs(c *gin.Context) {
	var balanceLogReq dto.GetBalanceLogReq

	var err error

	if err := c.ShouldBindQuery(&balanceLogReq); err != nil {
		if strings.Contains(err.Error(), "is not valid value for uuid.UUID") {
			// parse uuid from query
			operationTypeIDStr := c.Query("operation_type_id")
			operationTypeID, err := uuid.Parse(operationTypeIDStr)
			if err != nil {
				err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
				_ = c.Error(err)
				return
			}
			balanceLogReq.OperationTypeID = operationTypeID
		} else {
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			_ = c.Error(err)
			return
		}
	}

	if userID := c.GetString("user-id"); userID != "" {
		balanceLogReq.UserID, err = uuid.Parse(userID)
		if err != nil {
			bl.log.Error(err.Error(), zap.Any("user-id", userID))
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			_ = c.Error(err)
			return
		}
	}

	balanceLogsRes, err := bl.balanceLogsModule.GetBalanceLogs(c, balanceLogReq)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, balanceLogsRes)
}

// GetBalanceLogByID Get a balance log by its ID.
//
//	@Summary		Get a balance log by ID
//	@Description	Retrieve a single balance log entry by its unique ID.
//	@Tags			BalanceLogs
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token>"
//	@Param			id				path		string	true	"Balance Log ID (UUID)"
//	@Success		200				{object}	dto.BalanceLogsRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Failure		404				{object}	response.ErrorResponse
//	@Router			/api/balance/logs/{id} [get]
func (bl *balance_logs) GetBalanceLogByID(c *gin.Context) {

	balanceLogID := c.Param("id")
	if balanceLogID == "" {
		err := errors.ErrInvalidUserInput.New("balance log ID is required")
		_ = c.Error(err)
		return
	}

	id, err := uuid.Parse(balanceLogID)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid balance log ID format")
		_ = c.Error(err)
		return
	}

	balanceLogRes, err := bl.balanceLogsModule.GetBalanceLogByID(c, id)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, balanceLogRes)
}

// GetBalanceLogsForAdmin Getbalance logs.
//
//	@Summary		GetBalanceLogsForAdmin
//	@Description	Retrieve balance logs for admin based on various query parameters like username, start date, etc.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization			header		string	true	"Bearer <token> "
//	@Param			filter_username			query		string	false	"username (string, optional)"
//	@Param			filter_end_amount		query		decimal	false	"filter_end_amount (decimal, optional)"
//	@Param			filter_start_amount		query		decimal	false	"filter_start_amount (decimal, optional)"
//	@Param			filter_transaction_type	query		string	false	"filter_transaction_type (string (deposit,withdrawal), optional)"
//	@Param			filter_end_date			query		string	false	"filter_end_date (string , optional)"
//	@Param			filter_status			query		string	false	"filter_status (string , optional)"
//	@Param			filter_start_date		query		string	false	"filter_start_date (string , optional)"
//	@Param			sort_amount				query		string	false	"sort_amount (string , optional)"
//	@Param			sort_date				query		string	false	"filter_start_date (string , optional)"
//	@Param			sort_username			query		string	false	"sort_username (string , optional)"
//	@Param			page					query		string	true	"page type (required)"
//	@Param			per_page				query		string	true	"per-page type (required)"
//	@Success		200						{object}	dto.AdminGetBalanceLogsReq
//	@Failure		400						{object}	response.ErrorResponse
//	@Failure		401						{object}	response.ErrorResponse
//	@Router			/api/admin/balance/logs [get]
func (bl *balance_logs) GetBalanceLogsForAdmin(c *gin.Context) {
	var balanceLogReq dto.AdminGetBalanceLogsReq

	var err error

	if err := c.ShouldBindQuery(&balanceLogReq); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	balanceLogsRes, err := bl.balanceLogsModule.GetBalanceLogsForAdmin(c, balanceLogReq)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, balanceLogsRes)
}
