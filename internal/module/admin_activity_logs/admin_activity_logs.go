package admin_activity_logs

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/storage/admin_activity_logs"
	"go.uber.org/zap"
)

type AdminActivityLogsModule interface {
	LogAdminActivity(ctx context.Context, req dto.CreateAdminActivityLogReq) error
	GetAdminActivityLogs(ctx context.Context, req dto.GetAdminActivityLogsReq) (dto.AdminActivityLogsRes, error)
	GetAdminActivityLogByID(ctx context.Context, id uuid.UUID) (dto.AdminActivityLog, error)
	GetAdminActivityStats(ctx context.Context, from, to *time.Time) (dto.AdminActivityStats, error)
	GetAdminActivityCategories(ctx context.Context) ([]dto.AdminActivityCategory, error)
	GetAdminActivityActions(ctx context.Context) ([]dto.AdminActivityAction, error)
	GetAdminActivityActionsByCategory(ctx context.Context, category string) ([]dto.AdminActivityAction, error)
	DeleteAdminActivityLog(ctx context.Context, id uuid.UUID) error
	DeleteAdminActivityLogsByAdmin(ctx context.Context, adminUserID uuid.UUID) error
	DeleteOldAdminActivityLogs(ctx context.Context, before time.Time) error
}

type adminActivityLogsModule struct {
	storage admin_activity_logs.AdminActivityLogsStorage
	log     *zap.Logger
}

func NewAdminActivityLogsModule(storage admin_activity_logs.AdminActivityLogsStorage, log *zap.Logger) AdminActivityLogsModule {
	return &adminActivityLogsModule{
		storage: storage,
		log:     log,
	}
}

func (a *adminActivityLogsModule) LogAdminActivity(ctx context.Context, req dto.CreateAdminActivityLogReq) error {
	_, err := a.storage.CreateAdminActivityLog(ctx, req)
	if err != nil {
		a.log.Error("Failed to log admin activity", zap.Error(err))
		return err
	}
	return nil
}

func (a *adminActivityLogsModule) GetAdminActivityLogs(ctx context.Context, req dto.GetAdminActivityLogsReq) (dto.AdminActivityLogsRes, error) {
	return a.storage.GetAdminActivityLogs(ctx, req)
}

func (a *adminActivityLogsModule) GetAdminActivityLogByID(ctx context.Context, id uuid.UUID) (dto.AdminActivityLog, error) {
	return a.storage.GetAdminActivityLogByID(ctx, id)
}

func (a *adminActivityLogsModule) GetAdminActivityStats(ctx context.Context, from, to *time.Time) (dto.AdminActivityStats, error) {
	return a.storage.GetAdminActivityStats(ctx, from, to)
}

func (a *adminActivityLogsModule) GetAdminActivityCategories(ctx context.Context) ([]dto.AdminActivityCategory, error) {
	return a.storage.GetAdminActivityCategories(ctx)
}

func (a *adminActivityLogsModule) GetAdminActivityActions(ctx context.Context) ([]dto.AdminActivityAction, error) {
	return a.storage.GetAdminActivityActions(ctx)
}

func (a *adminActivityLogsModule) GetAdminActivityActionsByCategory(ctx context.Context, category string) ([]dto.AdminActivityAction, error) {
	return a.storage.GetAdminActivityActionsByCategory(ctx, category)
}

func (a *adminActivityLogsModule) DeleteAdminActivityLog(ctx context.Context, id uuid.UUID) error {
	return a.storage.DeleteAdminActivityLog(ctx, id)
}

func (a *adminActivityLogsModule) DeleteAdminActivityLogsByAdmin(ctx context.Context, adminUserID uuid.UUID) error {
	return a.storage.DeleteAdminActivityLogsByAdmin(ctx, adminUserID)
}

func (a *adminActivityLogsModule) DeleteOldAdminActivityLogs(ctx context.Context, before time.Time) error {
	return a.storage.DeleteOldAdminActivityLogs(ctx, before)
}
