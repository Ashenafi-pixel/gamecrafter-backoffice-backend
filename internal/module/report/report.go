package report

import (
	"context"

	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/module"
	"github.com/joshjones612/egyptkingcrash/internal/storage"
	"go.uber.org/zap"
)

type report struct {
	log           *zap.Logger
	reportStorage storage.Report
}

func Init(reportStorage storage.Report, log *zap.Logger) module.Report {
	return &report{
		log:           log,
		reportStorage: reportStorage,
	}
}

func (r *report) DailyReport(ctx context.Context, req dto.DailyReportReq) (dto.DailyReportRes, error) {
	if err := dto.ValidateDailyReportReq(req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.DailyReportRes{}, err
	}

	return r.reportStorage.DailyReport(ctx, req)
}
