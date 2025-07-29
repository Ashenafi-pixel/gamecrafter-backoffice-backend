package risksettings

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

type riskSettings struct {
	log                *zap.Logger
	riskSettingsModule module.RiskSettings
}

func Init(riskSettingsModule module.RiskSettings, log *zap.Logger) handler.RiskSettings {
	return &riskSettings{
		riskSettingsModule: riskSettingsModule,
		log:                log,
	}
}

// GetRiskSettings
// @Summary		GetRiskSettings
// @Description	GetRiskSettings retrieves the current risk settings
// @Tags			Admin
// @Accept			json
// @Produce		json
// @Param			Authorization	header		string	true	"Bearer <token>"
// @Success		200				{object}	dto.RiskSettings
// @Failure		400				{object}	response.ErrorResponse
// @Failure		500				{object}	response.ErrorResponse
// @Router			/api/admin/risksettings [get]

func (r *riskSettings) GetRiskSettings(c *gin.Context) {

	riskSettings, err := r.riskSettingsModule.GetRiskSettings(c)
	if err != nil {
		r.log.Warn("Failed to get risk settings", zap.Error(err))
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, riskSettings)
}

// SetRiskSettings
// @Summary		SetRiskSettings
// @Description	SetRiskSettings updates the risk settings
// @Tags			Admin
// @Accept			json
// @Produce		json
// @Param			Authorization	header		string	true	"Bearer <token>"
// @Param			riskSettings	body		dto.RiskSettings	true	"Risk Settings"
// @Success		200				{object}	dto.RiskSettings
// @Failure		400				{object}	response.ErrorResponse
// @Failure		500				{object}	response.ErrorResponse
// @Router			/api/admin/risksettings [put]

func (r *riskSettings) SetRiskSettings(c *gin.Context) {
	var riskSettings dto.RiskSettings
	if err := c.ShouldBindJSON(&riskSettings); err != nil {
		r.log.Error("Invalid risk settings input", zap.Error(err))
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid risk settings input")
		_ = c.Error(err)
		return
	}

	updatedRiskSettings, err := r.riskSettingsModule.SetRiskSettings(c, riskSettings)
	if err != nil {
		r.log.Error("Failed to set risk settings", zap.Error(err))
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, updatedRiskSettings)
}
