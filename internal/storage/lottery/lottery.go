package lottery

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/db"
	"github.com/tucanbit/internal/constant/persistencedb"
	"go.uber.org/zap"
)

type lottery struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

func Init(db *persistencedb.PersistenceDB, log *zap.Logger) *lottery {
	return &lottery{
		db:  db,
		log: log,
	}
}

func (l *lottery) CreateLotteryService(ctx context.Context, req dto.CreateLotteryServiceReq) (dto.CreateLotteryServiceRes, error) {
	resp, err := l.db.Queries.CreateLotteryService(ctx, db.CreateLotteryServiceParams{
		Name:         req.ServiceName,
		Description:  sql.NullString{String: req.Description, Valid: req.Description != ""},
		ClientID:     req.ServiceClientID,
		ClientSecret: req.ServiceSecret,
		CallbackUrl:  sql.NullString{String: req.CallbackURL, Valid: req.CallbackURL != ""},
	})

	if err != nil {
		l.log.Error("failed to create lottery service", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.CreateLotteryServiceRes{}, err
	}

	return dto.CreateLotteryServiceRes{
		ServiceID: resp.ID,
	}, nil
}

func (l *lottery) GetLotteryServiceByID(ctx context.Context, serviceID uuid.UUID) (dto.CreateLotteryServiceReq, error) {
	resp, err := l.db.Queries.GetLotteryServiceByID(ctx, serviceID)
	if err != nil {
		l.log.Error("failed to get lottery service by ID", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.CreateLotteryServiceReq{}, err
	}

	return dto.CreateLotteryServiceReq{
		ServiceName:     resp.Name,
		ServiceSecret:   resp.ClientSecret,
		ServiceClientID: resp.ClientID,
		Description:     resp.Description.String,
		CallbackURL:     resp.CallbackUrl.String,
	}, nil
}

func (l *lottery) CreateLotteryWinnersLogs(ctx context.Context, req dto.LotteryLog) (dto.LotteryLog, error) {

	resp, err := l.db.Queries.CreateLotteryWinnersLog(ctx, db.CreateLotteryWinnersLogParams{
		LotteryID:       req.LotteryID,
		UserID:          req.UserID,
		RewardID:        req.RewardID,
		WonAmount:       req.WonAmount,
		Currency:        req.Currency,
		TicketNumber:    req.TicketNumber,
		NumberOfTickets: int32(req.NumberOfTickets),
	})
	if err != nil {
		l.log.Error("failed to create lottery log", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.LotteryLog{}, err
	}
	return dto.LotteryLog{
		ID:           resp.ID,
		LotteryID:    resp.LotteryID,
		UserID:       resp.UserID,
		RewardID:     resp.RewardID,
		WonAmount:    resp.WonAmount,
		Currency:     resp.Currency,
		TicketNumber: resp.TicketNumber,
		Status:       resp.Status,
		CreatedAt:    resp.CreatedAt,
		UpdatedAt:    resp.UpdatedAt,
	}, nil
}
func (l *lottery) CreateLotteryLog(ctx context.Context, req dto.LotteryKafkaLog) (dto.LotteryKafkaLog, error) {
	resp, err := l.db.Queries.CreateLotteryLog(ctx, db.CreateLotteryLogParams{
		LotteryID:       req.LotteryID,
		LotteryRewardID: req.LotteryRewardID,
		DrawNumbers:     req.DrawNumbers,
		Prize:           req.Prize,
	})
	if err != nil {
		l.log.Error("failed to create lottery log", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.LotteryKafkaLog{}, err
	}
	return dto.LotteryKafkaLog{
		ID:              resp.ID,
		LotteryID:       resp.LotteryID,
		LotteryRewardID: resp.LotteryRewardID,
		DrawNumbers:     resp.DrawNumbers,
		Prize:           resp.Prize,
		CreatedAt:       resp.CreatedAt,
		UpdatedAt:       resp.UpdatedAt,
	}, nil
}

func (l *lottery) GetAvailableLotteryService(ctx context.Context) (dto.CreateLotteryServiceReq, error) {
	resp, err := l.db.Queries.GetAvailableLotteryService(ctx)
	if err != nil {
		l.log.Error("failed to get available lottery service", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.CreateLotteryServiceReq{}, err
	}

	return dto.CreateLotteryServiceReq{
		ServiceName:     resp.Name,
		ServiceSecret:   resp.ClientSecret,
		ServiceClientID: resp.ClientID,
		Description:     resp.Description.String,
		CallbackURL:     resp.CallbackUrl.String,
	}, nil
}

func (l *lottery) GetLotteryLogsByUniqIdentifier(ctx context.Context, uniqIdentifier uuid.UUID) ([]dto.LotteryKafkaLog, error) {
	resp, err := l.db.Queries.GetLotteryLogsByUniqIdentifier(ctx, uniqIdentifier)
	if err != nil {
		l.log.Error("failed to get lottery logs by uniq identifier", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return nil, err
	}

	var logs []dto.LotteryKafkaLog
	for _, log := range resp {
		logs = append(logs, dto.LotteryKafkaLog{
			ID:              log.ID,
			LotteryID:       log.LotteryID,
			LotteryRewardID: log.LotteryRewardID,
			DrawNumbers:     log.DrawNumbers,
			Prize:           log.Prize,
			CreatedAt:       log.CreatedAt,
			UpdatedAt:       log.UpdatedAt,
			UniqIdentifier:  log.UniqIdentifier,
		})
	}
	return logs, nil
}
