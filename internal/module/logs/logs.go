package logs

import (
	"context"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/module"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
)

type SystemLogs struct {
	log        *zap.Logger
	logStorage storage.Logs
}

func Init(log *zap.Logger, logStorage storage.Logs) module.SystemLogs {
	return &SystemLogs{
		log:        log,
		logStorage: logStorage,
	}
}

func (s *SystemLogs) CreateSystemLogs(ctx context.Context, systemLogReq dto.SystemLogs) (dto.SystemLogs, error) {
	return s.logStorage.CreateSystemLogs(ctx, systemLogReq)
}

func (s *SystemLogs) GetSystemLogs(ctx context.Context, req dto.GetSystemLogsReq) (dto.SystemLogsRes, error) {
	empty := dto.SystemLogsRes{
		SystemLogs: []dto.SystemLogs{},
		TotalPage:  0,
	}
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset

	if req.UserID != uuid.Nil {
		resp, err := s.logStorage.GetSystemLogsByUser(ctx, req)
		if err != nil {
			s.log.Error("Error fetching system logs by user", zap.Error(err))
			return empty, err
		}
		if len(resp.SystemLogs) == 0 {
			s.log.Warn("No system logs found for user", zap.String("user_id", req.UserID.String()))
			return empty, nil
		}
		return resp, nil
	}

	if req.Module != "" {
		resp, err := s.logStorage.GetSystemLogsByModule(ctx, req)
		if err != nil {
			s.log.Error("Error fetching system logs by module", zap.Error(err))
			return empty, err
		}
		if len(resp.SystemLogs) == 0 {
			s.log.Warn("No system logs found for module", zap.String("module", req.Module))
			return empty, nil
		}
		return resp, nil
	}

	if req.From != nil && req.To != nil {
		resp, err := s.logStorage.GetSystemLogsByStartAndEndDate(ctx, req)
		if err != nil {
			s.log.Error("Error fetching system logs by start and end date", zap.Error(err))
			return empty, err
		}
		if len(resp.SystemLogs) == 0 {
			s.log.Warn("No system logs found for the specified date range", zap.Time("from", *req.From), zap.Time("to", *req.To))
			return empty, nil
		}
		return resp, nil
	}

	if req.From != nil {
		resp, err := s.logStorage.GetSystemLogsByStartData(ctx, req)
		if err != nil {
			s.log.Error("Error fetching system logs by start date", zap.Error(err))
			return empty, err
		}
		if len(resp.SystemLogs) == 0 {
			s.log.Warn("No system logs found for the specified start date", zap.Time("from", *req.From))
			return empty, nil
		}
		return resp, nil
	}

	if req.To != nil {
		resp, err := s.logStorage.GetSystemLogsByEndDate(ctx, req)
		if err != nil {
			s.log.Error("Error fetching system logs by end date", zap.Error(err))
			return empty, err
		}
		if len(resp.SystemLogs) == 0 {
			s.log.Warn("No system logs found for the specified end date", zap.Time("to", *req.To))
			return empty, nil
		}
		return resp, nil
	}

	resp, err := s.logStorage.GetSystemLogs(ctx, dto.GetRequest{
		Page:    req.Page,
		PerPage: req.PerPage,
	})
	if err != nil {
		s.log.Error("Error fetching system logs", zap.Error(err))
		return empty, err
	}
	if len(resp.SystemLogs) == 0 {
		s.log.Warn("No system logs found")
		return empty, nil
	}
	return resp, nil
}

func (s *SystemLogs) GetAvailableLogsModule(ctx context.Context) ([]string, error) {
	return s.logStorage.GetAvailableModules(ctx)
}
