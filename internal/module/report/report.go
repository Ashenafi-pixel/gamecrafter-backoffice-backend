package report

import (
	"context"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/module"
	"github.com/tucanbit/internal/storage"
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

func (r *report) GetDuplicateIPAccounts(ctx context.Context) ([]dto.DuplicateIPAccountsReport, error) {
	return r.reportStorage.GetDuplicateIPAccounts(ctx)
}

func (r *report) SuspendAccountsByIP(ctx context.Context, req dto.SuspendAccountsByIPReq, adminID uuid.UUID) ([]uuid.UUID, error) {
	// Get all user IDs from the IP address
	userIDs, err := r.reportStorage.SuspendAccountsByIP(ctx, req.IPAddress)
	if err != nil {
		r.log.Error("Failed to get user IDs from IP", zap.Error(err), zap.String("ip_address", req.IPAddress))
		return nil, err
	}

	return userIDs, nil
}

func (r *report) GetBigWinners(ctx context.Context, req dto.BigWinnersReportReq, userBrandIDs []uuid.UUID) (dto.BigWinnersReportRes, error) {
	return r.reportStorage.GetBigWinners(ctx, req, userBrandIDs)
}

func (r *report) GetPlayerMetrics(ctx context.Context, req dto.PlayerMetricsReportReq, userBrandIDs []uuid.UUID) (dto.PlayerMetricsReportRes, error) {
	return r.reportStorage.GetPlayerMetrics(ctx, req, userBrandIDs)
}

func (r *report) GetPlayerTransactions(ctx context.Context, req dto.PlayerTransactionsReq) (dto.PlayerTransactionsRes, error) {
	return r.reportStorage.GetPlayerTransactions(ctx, req)
}

func (r *report) GetCountryMetrics(ctx context.Context, req dto.CountryReportReq, userBrandIDs []uuid.UUID) (dto.CountryReportRes, error) {
	return r.reportStorage.GetCountryMetrics(ctx, req, userBrandIDs)
}

func (r *report) GetCountryPlayers(ctx context.Context, req dto.CountryPlayersReq, userBrandIDs []uuid.UUID) (dto.CountryPlayersRes, error) {
	return r.reportStorage.GetCountryPlayers(ctx, req, userBrandIDs)
}

func (r *report) GetGamePerformance(ctx context.Context, req dto.GamePerformanceReportReq, userBrandIDs []uuid.UUID) (dto.GamePerformanceReportRes, error) {
	return r.reportStorage.GetGamePerformance(ctx, req, userBrandIDs)
}

func (r *report) GetGamePlayers(ctx context.Context, req dto.GamePlayersReq, userBrandIDs []uuid.UUID) (dto.GamePlayersRes, error) {
	return r.reportStorage.GetGamePlayers(ctx, req, userBrandIDs)
}

func (r *report) GetProviderPerformance(ctx context.Context, req dto.ProviderPerformanceReportReq, userBrandIDs []uuid.UUID) (dto.ProviderPerformanceReportRes, error) {
	return r.reportStorage.GetProviderPerformance(ctx, req, userBrandIDs)
}

func (r *report) GetAffiliateReport(ctx context.Context, req dto.AffiliateReportReq, adminID uuid.UUID) (dto.AffiliateReportRes, error) {
	// TODO: Get allowed referral codes for this admin from RBAC
	// For now, allow all referral codes (empty slice means no restriction)
	allowedReferralCodes := []string{}
	
	return r.reportStorage.GetAffiliateReport(ctx, req, allowedReferralCodes)
}
