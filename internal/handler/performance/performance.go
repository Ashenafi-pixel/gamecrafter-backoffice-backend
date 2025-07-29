package performance

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joshjones612/egyptkingcrash/internal/constant/model/response"
	"github.com/joshjones612/egyptkingcrash/internal/handler"
	"github.com/joshjones612/egyptkingcrash/internal/module"
	"go.uber.org/zap"
)

type performance struct {
	performanceModule module.Performance
	log               *zap.Logger
}

func Init(performanceModule module.Performance, log *zap.Logger) handler.Performance {
	return &performance{
		performanceModule: performanceModule,
		log:               log,
	}
}

// GetFinancialMetrics Admin Get Financial Metrics.
//	@Summary		GetFinancialMetrics
//	@Description	Retrieve Financial Metrics.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token> "
//	@Success		200				{object}	[]dto.FinancialMatrix
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admin/metrics/financial [get]
func (p *performance) GetFinancialMetrics(c *gin.Context) {

	resp, err := p.performanceModule.GetFinancialMatrix(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GameMatrics Admin Get Game Metrics.
//	@Summary		GameMatrics
//	@Description	Retrieve Game Metrics.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token> "
//	@Success		200				{object}	dto.GameMatricsRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admin/metrics/game [get]
func (p *performance) GameMatrics(c *gin.Context) {

	resp, err := p.performanceModule.GetGameMatrics(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}
