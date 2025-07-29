package performance

import (
	"context"

	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/module"
	"github.com/joshjones612/egyptkingcrash/internal/storage"
	"go.uber.org/zap"
)

type performance struct {
	performanceStorage storage.Performance
	log                *zap.Logger
}

func Init(performanceStorage storage.Performance, log *zap.Logger) module.Performance {
	return &performance{
		performanceStorage: performanceStorage,
		log:                log,
	}
}

func (p *performance) GetFinancialMatrix(ctx context.Context) ([]dto.FinancialMatrix, error) {
	return p.performanceStorage.GetFinancialMatrix(ctx)
}

func (p *performance) GetGameMatrics(ctx context.Context) (dto.GameMatricsRes, error) {
	return p.performanceStorage.GetGameMatrics(ctx)
}
