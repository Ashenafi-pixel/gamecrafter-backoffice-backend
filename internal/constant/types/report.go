package types

import (
	"context"

	"github.com/tucanbit/internal/constant/dto"
)

type Report interface {
	GetDailyReport(ctx context.Context, req dto.DailyReportReq) (dto.DailyReportRes, error)
}
