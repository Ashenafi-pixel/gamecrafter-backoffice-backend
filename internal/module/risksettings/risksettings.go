package risksettings

import (
	"context"

	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/module"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
)

type risksettings struct {
	log                 *zap.Logger
	riskSettingsStorage storage.RiskSettings
}

func Init(riskSettingsStorage storage.RiskSettings, log *zap.Logger) module.RiskSettings {
	return &risksettings{
		log:                 log,
		riskSettingsStorage: riskSettingsStorage,
	}
}

func (r *risksettings) GetRiskSettings(ctx context.Context) (dto.RiskSettings, error) {
	return r.riskSettingsStorage.GetRiskSettings(ctx)
}

func (r *risksettings) SetRiskSettings(ctx context.Context, req dto.RiskSettings) (dto.RiskSettings, error) {
	return r.riskSettingsStorage.SetRiskSettings(ctx, req)
}
