package report

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/constant/model/response"
	"github.com/joshjones612/egyptkingcrash/internal/handler"
	"github.com/joshjones612/egyptkingcrash/internal/module"
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
