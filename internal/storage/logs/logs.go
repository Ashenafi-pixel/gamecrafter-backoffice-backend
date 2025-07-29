package logs

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/constant/model/db"
	"github.com/joshjones612/egyptkingcrash/internal/constant/persistencedb"
	"github.com/joshjones612/egyptkingcrash/internal/storage"
	"go.uber.org/zap"
)

type logs struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

func Init(db *persistencedb.PersistenceDB, log *zap.Logger) storage.Logs {
	return &logs{
		db:  db,
		log: log,
	}
}

func (l *logs) CreateLoginAttempts(ctx context.Context, loginAttemptReq dto.LoginAttempt) (dto.LoginAttempt, error) {
	lRes, err := l.db.Queries.CreateLoginAttemptsLog(ctx, db.CreateLoginAttemptsLogParams{
		UserID:      uuid.NullUUID{UUID: loginAttemptReq.UserID, Valid: true},
		IpAddress:   loginAttemptReq.IPAddress,
		Success:     loginAttemptReq.Success,
		AttemptTime: sql.NullTime{Time: loginAttemptReq.AttemptTime, Valid: true},
		UserAgent:   sql.NullString{String: loginAttemptReq.UserAgent, Valid: true},
	})
	if err != nil {
		l.log.Error(err.Error(), zap.Any("loginAttemptReq", loginAttemptReq))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.LoginAttempt{}, err
	}
	return dto.LoginAttempt{
		ID:          lRes.ID,
		UserID:      lRes.UserID.UUID,
		IPAddress:   lRes.IpAddress,
		Success:     lRes.Success,
		AttemptTime: lRes.AttemptTime.Time,
		UserAgent:   lRes.UserAgent.String,
	}, nil
}

func (l *logs) CreateLoginSessions(ctx context.Context, userSessionReq dto.UserSessions) (dto.UserSessions, error) {
	userSession, err := l.db.Queries.CreateUserSessions(ctx, db.CreateUserSessionsParams{
		UserID:                uuid.NullUUID{UUID: userSessionReq.UserID, Valid: true},
		Token:                 userSessionReq.Token,
		ExpiresAt:             userSessionReq.ExpiresAt,
		IpAddress:             sql.NullString{String: userSessionReq.IpAddress, Valid: true},
		UserAgent:             sql.NullString{String: userSessionReq.UserAgent, Valid: true},
		CreatedAt:             sql.NullTime{Time: time.Now(), Valid: true},
		RefreshToken:          sql.NullString{String: userSessionReq.RefreshToken, Valid: true},
		RefreshTokenExpiresAt: sql.NullTime{Time: userSessionReq.RefreshTokenExpiresAt, Valid: true},
	})
	if err != nil {
		l.log.Error(err.Error(), zap.Any("userSessionReq", userSessionReq))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.UserSessions{}, err
	}
	return dto.UserSessions{
		ID:                    userSession.ID,
		UserID:                userSession.UserID.UUID,
		Token:                 userSession.Token,
		ExpiresAt:             userSession.ExpiresAt,
		IpAddress:             userSession.IpAddress.String,
		UserAgent:             userSession.UserAgent.String,
		CreatedAt:             userSession.CreatedAt.Time,
		RefreshToken:          userSession.RefreshToken.String,
		RefreshTokenExpiresAt: userSession.RefreshTokenExpiresAt.Time,
	}, nil
}

func (l *logs) CreateSystemLogs(ctx context.Context, systemLogReq dto.SystemLogs) (dto.SystemLogs, error) {
	logData, err := convertToPgJSON(systemLogReq.Detail)
	if err != nil {
		l.log.Error(err.Error(), zap.Any("systemLogReq", systemLogReq))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.SystemLogs{}, err
	}

	systemLog, err := l.db.Queries.CreateSystemLog(ctx, db.CreateSystemLogParams{
		UserID:    systemLogReq.UserID,
		IpAddress: sql.NullString{String: systemLogReq.IPAddress, Valid: true},
		Module:    systemLogReq.Module,
		Detail:    logData,
		Timestamp: sql.NullTime{Time: systemLogReq.Timestamp, Valid: true},
	})

	if err != nil {
		l.log.Error(err.Error(), zap.Any("systemLogReq", systemLogReq))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.SystemLogs{}, err
	}

	logsDetail, err := convertPgJSONToInterface(systemLog.Detail)
	if err != nil {
		l.log.Error(err.Error(), zap.Any("systemLog", systemLog))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.SystemLogs{}, err
	}

	return dto.SystemLogs{
		ID:        systemLog.ID,
		UserID:    systemLog.UserID,
		Module:    systemLog.Module,
		Detail:    logsDetail,
		IPAddress: systemLog.IpAddress.String,
		Timestamp: systemLog.Timestamp.Time,
	}, nil
}

func (l *logs) GetSystemLogs(ctx context.Context, req dto.GetRequest) (dto.SystemLogsRes, error) {
	var systemLogs []dto.SystemLogs
	sysLogs, err := l.db.Queries.GetSystemLogs(ctx, db.GetSystemLogsParams{
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		l.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.SystemLogsRes{}, err
	}

	if err != nil {
		return dto.SystemLogsRes{}, nil
	}

	totalPage := 1
	for i, log := range sysLogs {
		logsDetail, err := convertPgJSONToInterface(log.Detail)
		if err != nil {
			l.log.Error(err.Error(), zap.Any("log", log))
			err = errors.ErrUnableTocreate.Wrap(err, err.Error())
			return dto.SystemLogsRes{}, err
		}

		systemLogs = append(systemLogs, dto.SystemLogs{
			ID:        log.ID,
			UserID:    log.UserID,
			Module:    log.Module,
			Detail:    logsDetail,
			IPAddress: log.IpAddress.String,
			Timestamp: log.Timestamp.Time,
			Roles:     log.Roles,
		})
		if i == 0 {
			ps := int(log.TotalRows) / req.PerPage
			if int(log.TotalRows)%req.PerPage != 0 {
				ps += 1
			}
			totalPage = ps
		}
	}

	return dto.SystemLogsRes{
		SystemLogs: systemLogs,
		TotalPage:  totalPage,
	}, nil
}

func (l *logs) GetSystemLogsByModule(ctx context.Context, req dto.GetSystemLogsReq) (dto.SystemLogsRes, error) {
	var systemLogs []dto.SystemLogs
	sysLogs, err := l.db.Queries.GetSystemLogsByModule(ctx, db.GetSystemLogsByModuleParams{
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
		Module: req.Module,
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		l.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.SystemLogsRes{}, err
	}

	if err != nil {
		return dto.SystemLogsRes{}, nil
	}

	totalPage := 1

	for i, log := range sysLogs {
		logsDetail, err := convertPgJSONToInterface(log.Detail)
		if err != nil {
			l.log.Error(err.Error(), zap.Any("log", log))
			err = errors.ErrUnableTocreate.Wrap(err, err.Error())
			return dto.SystemLogsRes{}, err
		}

		systemLogs = append(systemLogs, dto.SystemLogs{
			ID:        log.ID,
			UserID:    log.UserID,
			Module:    log.Module,
			Detail:    logsDetail,
			IPAddress: log.IpAddress.String,
			Timestamp: log.Timestamp.Time,
			Roles:     log.Roles,
		})
		if i == 0 {
			ps := int(log.TotalRows) / req.PerPage
			if int(log.TotalRows)%req.PerPage != 0 {
				ps += 1
			}
			totalPage = ps
		}
	}

	return dto.SystemLogsRes{SystemLogs: systemLogs, TotalPage: totalPage}, nil
}

func (l *logs) GetSystemLogsByUser(ctx context.Context, req dto.GetSystemLogsReq) (dto.SystemLogsRes, error) {
	var systemLogs []dto.SystemLogs
	sysLogs, err := l.db.Queries.GetSystemLogsByUserID(ctx, db.GetSystemLogsByUserIDParams{
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
		UserID: req.UserID,
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		l.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.SystemLogsRes{}, err
	}

	if err != nil {
		return dto.SystemLogsRes{}, nil
	}
	totalPage := 1

	for i, log := range sysLogs {
		logsDetail, err := convertPgJSONToInterface(log.Detail)
		if err != nil {
			l.log.Error(err.Error(), zap.Any("log", log))
			err = errors.ErrUnableTocreate.Wrap(err, err.Error())
			return dto.SystemLogsRes{}, err
		}

		systemLogs = append(systemLogs, dto.SystemLogs{
			ID:        log.ID,
			UserID:    log.UserID,
			Module:    log.Module,
			Detail:    logsDetail,
			IPAddress: log.IpAddress.String,
			Timestamp: log.Timestamp.Time,
			Roles:     log.Roles,
		})
		if i == 0 {
			ps := int(log.TotalRows) / req.PerPage
			if int(log.TotalRows)%req.PerPage != 0 {
				ps += 1
			}
			totalPage = ps
		}
	}

	return dto.SystemLogsRes{SystemLogs: systemLogs, TotalPage: totalPage}, nil
}

func (l *logs) GetSystemLogsByStartData(ctx context.Context, req dto.GetSystemLogsReq) (dto.SystemLogsRes, error) {
	var systemLogs []dto.SystemLogs
	sysLogs, err := l.db.Queries.GetSystemLogsByStartData(ctx, db.GetSystemLogsByStartDataParams{
		Limit:     int32(req.PerPage),
		Offset:    int32(req.Page),
		Timestamp: sql.NullTime{Time: *req.From, Valid: true},
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		l.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.SystemLogsRes{}, err
	}

	if err != nil {
		return dto.SystemLogsRes{}, nil
	}
	totalPage := 1

	for i, log := range sysLogs {
		logsDetail, err := convertPgJSONToInterface(log.Detail)
		if err != nil {
			l.log.Error(err.Error(), zap.Any("log", log))
			err = errors.ErrUnableTocreate.Wrap(err, err.Error())
			return dto.SystemLogsRes{}, err
		}

		systemLogs = append(systemLogs, dto.SystemLogs{
			ID:        log.ID,
			UserID:    log.UserID,
			Module:    log.Module,
			Detail:    logsDetail,
			IPAddress: log.IpAddress.String,
			Timestamp: log.Timestamp.Time,
			Roles:     log.Roles,
		})
		if i == 0 {
			ps := int(log.TotalRows) / req.PerPage
			if int(log.TotalRows)%req.PerPage != 0 {
				ps += 1
			}
			totalPage = ps
		}
	}

	return dto.SystemLogsRes{SystemLogs: systemLogs, TotalPage: totalPage}, nil
}

func (l *logs) GetSystemLogsByEndDate(ctx context.Context, req dto.GetSystemLogsReq) (dto.SystemLogsRes, error) {
	var systemLogs []dto.SystemLogs
	sysLogs, err := l.db.Queries.GetSystemLogsByEndData(ctx, db.GetSystemLogsByEndDataParams{
		Limit:     int32(req.PerPage),
		Offset:    int32(req.Page),
		Timestamp: sql.NullTime{Time: *req.To, Valid: true},
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		l.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.SystemLogsRes{}, err
	}

	if err != nil {
		return dto.SystemLogsRes{}, nil
	}
	totalPage := 1
	for i, log := range sysLogs {
		logsDetail, err := convertPgJSONToInterface(log.Detail)
		if err != nil {
			l.log.Error(err.Error(), zap.Any("log", log))
			err = errors.ErrUnableTocreate.Wrap(err, err.Error())
			return dto.SystemLogsRes{}, err
		}

		systemLogs = append(systemLogs, dto.SystemLogs{
			ID:        log.ID,
			UserID:    log.UserID,
			Module:    log.Module,
			Detail:    logsDetail,
			IPAddress: log.IpAddress.String,
			Timestamp: log.Timestamp.Time,
			Roles:     log.Roles,
		})
		if i == 0 {
			ps := int(log.TotalRows) / req.PerPage
			if int(log.TotalRows)%req.PerPage != 0 {
				ps += 1
			}
			totalPage = ps

		}
	}

	return dto.SystemLogsRes{SystemLogs: systemLogs, TotalPage: totalPage}, nil
}

func (l *logs) GetSystemLogsByStartAndEndDate(ctx context.Context, req dto.GetSystemLogsReq) (dto.SystemLogsRes, error) {
	var systemLogs []dto.SystemLogs
	sysLogs, err := l.db.Queries.GetSystemLogsByStartAndEndData(ctx, db.GetSystemLogsByStartAndEndDataParams{
		Limit:       int32(req.PerPage),
		Offset:      int32(req.Page),
		Timestamp:   sql.NullTime{Time: *req.From, Valid: true},
		Timestamp_2: sql.NullTime{Time: *req.To, Valid: true},
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		l.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.SystemLogsRes{}, err
	}

	if err != nil {
		return dto.SystemLogsRes{}, nil
	}

	totalPage := 1

	for i, log := range sysLogs {
		logsDetail, err := convertPgJSONToInterface(log.Detail)
		if err != nil {
			l.log.Error(err.Error(), zap.Any("log", log))
			err = errors.ErrUnableTocreate.Wrap(err, err.Error())
			return dto.SystemLogsRes{}, err
		}

		systemLogs = append(systemLogs, dto.SystemLogs{
			ID:        log.ID,
			UserID:    log.UserID,
			Module:    log.Module,
			Detail:    logsDetail,
			IPAddress: log.IpAddress.String,
			Timestamp: log.Timestamp.Time,
			Roles:     log.Roles,
		})
		if i == 0 {
			ps := int(log.TotalRows) / req.PerPage
			if int(log.TotalRows)%req.PerPage != 0 {
				ps += 1
			}
			totalPage = ps

		}
	}
	systemLogsRes := dto.SystemLogsRes{
		SystemLogs: systemLogs,
		TotalPage:  totalPage,
	}

	return systemLogsRes, nil
}

func (l *logs) GetAvailableModules(ctx context.Context) ([]string, error) {
	modules, err := l.db.Queries.GetAvailableModule(ctx)
	if err != nil {
		l.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []string{}, err
	}

	return modules, nil
}

func (l *logs) GetUserSessionByRefreshToken(ctx context.Context, refreshToken string) (dto.UserSessions, error) {
	session, err := l.db.Queries.GetUserSessionByRefreshToken(ctx, sql.NullString{String: refreshToken, Valid: true})
	if err != nil {
		return dto.UserSessions{}, err
	}
	return dto.UserSessions{
		ID:                    session.ID,
		UserID:                session.UserID.UUID,
		Token:                 session.Token,
		ExpiresAt:             session.ExpiresAt,
		RefreshToken:          session.RefreshToken.String,
		RefreshTokenExpiresAt: session.RefreshTokenExpiresAt.Time,
		IpAddress:             session.IpAddress.String,
		UserAgent:             session.UserAgent.String,
		CreatedAt:             session.CreatedAt.Time,
	}, nil
}

func (l *logs) UpdateUserSessionRefreshToken(ctx context.Context, sessionID uuid.UUID, newToken string, newExpiry time.Time) error {
	return l.db.Queries.UpdateUserSessionRefreshToken(ctx, db.UpdateUserSessionRefreshTokenParams{
		ID:                    sessionID,
		RefreshToken:          sql.NullString{String: newToken, Valid: true},
		RefreshTokenExpiresAt: sql.NullTime{Time: newExpiry, Valid: true},
	})
}

func (l *logs) InvalidateOldUserSessions(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error {
	return l.db.Queries.InvalidateOldUserSessions(ctx, db.InvalidateOldUserSessionsParams{
		UserID: uuid.NullUUID{UUID: userID, Valid: true},
		ID:     sessionID,
	})
}

func (l *logs) GetSessionsExpiringSoon(ctx context.Context, duration time.Duration) ([]dto.UserSession, error) {
	expiryTime := time.Now().Add(duration)

	sessions, err := l.db.Queries.GetSessionsExpiringSoon(ctx, sql.NullTime{
		Time:  expiryTime,
		Valid: true,
	})
	if err != nil {
		l.log.Error(err.Error(), zap.Duration("duration", duration))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return nil, err
	}

	var userSessions []dto.UserSession
	for _, session := range sessions {
		userSessions = append(userSessions, dto.UserSession{
			ID:                 session.ID,
			UserID:             session.UserID.UUID,
			Token:              session.Token,
			ExpiresAt:          session.ExpiresAt,
			IpAddress:          session.IpAddress.String,
			UserAgent:          session.UserAgent.String,
			CreatedAt:          session.CreatedAt.Time,
			RefreshToken:       session.RefreshToken.String,
			RefreshTokenExpiry: session.RefreshTokenExpiresAt.Time,
		})
	}

	return userSessions, nil
}

func (l *logs) InvalidateAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	err := l.db.Queries.InvalidateAllUserSessions(ctx, uuid.NullUUID{
		UUID:  userID,
		Valid: true,
	})
	if err != nil {
		l.log.Error(err.Error(), zap.String("userID", userID.String()))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return err
	}
	return nil
}
