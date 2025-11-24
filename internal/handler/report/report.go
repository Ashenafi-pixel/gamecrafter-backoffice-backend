package report

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

type report struct {
	log          *zap.Logger
	reportModule module.Report
}

func Init(reportModule module.Report, log *zap.Logger) handler.Report {
	return &report{
		log:          log,
		reportModule: reportModule,
	}
}

// GetDailyReport Get Daily Report.
//
//	@Summary		GetDailyReport
//	@Description	Returns daily aggregated report
//	@Tags			Report
//	@Accept			json
//	@Produce		json
//	@Param			date	query		string	true	"Report Date (YYYY-MM-DD)"
//	@Success		200		{object}	dto.DailyReportRes
//	@Failure		400		{object}	response.ErrorResponse
//	@Router			/api/admin/report/daily [get]
func (r *report) GetDailyReport(ctx *gin.Context) {
	var req dto.DailyReportReq

	if err := ctx.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	reportRes, err := r.reportModule.DailyReport(ctx, req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, reportRes)
}

// GetDuplicateIPAccounts Get Duplicate IP Accounts Report.
//
//	@Summary		GetDuplicateIPAccounts
//	@Description	Returns a report of accounts created from the same IP address (potential bot/spam accounts)
//	@Tags			Report
//	@Accept			json
//	@Produce		json
//	@Success		200		{array}	dto.DuplicateIPAccountsReport
//	@Failure		400		{object}	response.ErrorResponse
//	@Router			/api/admin/report/duplicate-ip-accounts [get]
func (r *report) GetDuplicateIPAccounts(ctx *gin.Context) {
	reportRes, err := r.reportModule.GetDuplicateIPAccounts(ctx.Request.Context())
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, reportRes)
}
