package bet

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/db"
	"github.com/tucanbit/platform/utils"
	"go.uber.org/zap"
)

func (b *bet) SaveRounds(ctx context.Context, betroundReq dto.BetRound) (dto.BetRound, error) {
	now := time.Now().In(time.Now().Location()).UTC()
	betroundRes, err := b.db.Queries.SaveBetRound(ctx, db.SaveBetRoundParams{
		Status:     db.BetStatus(betroundReq.Status),
		CrashPoint: betroundReq.CrashPoint,
		CreatedAt:  now,
	})
	if err != nil {
		b.log.Error(err.Error(), zap.Any("betRoundReq", betroundReq))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.BetRound{}, err
	}
	return dto.BetRound{
		ID:         betroundRes.ID,
		Status:     betroundReq.Status,
		CrashPoint: betroundRes.CrashPoint,
		CreatedAt:  &betroundRes.CreatedAt,
	}, nil
}

func (b *bet) GetBetRoundsByStatus(ctx context.Context, status string) ([]dto.BetRound, bool, error) {
	betRounds := []dto.BetRound{}
	betRds, err := b.db.Queries.GetBetRoundsByStatus(ctx, db.BetStatus(status))
	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error(), zap.Any("status", status))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.BetRound{}, false, err
	}
	if err != nil || len(betRds) == 0 {
		return []dto.BetRound{}, false, nil
	}
	for _, bet := range betRds {
		betRounds = append(betRounds, dto.BetRound{
			ID:         bet.ID,
			Status:     string(bet.Status),
			CrashPoint: bet.CrashPoint,
			CreatedAt:  &bet.CreatedAt,
			ClosedAt:   &bet.ClosedAt.Time,
		})
	}

	return betRounds, true, nil
}

func (b *bet) UpdateRoundStatusByID(ctx context.Context, roundID uuid.UUID, status string) (dto.BetRound, error) {
	betRound, err := b.db.Queries.UpdateRoundStatusByID(ctx, db.UpdateRoundStatusByIDParams{
		Status: db.BetStatus(status),
		ID:     roundID,
	})
	if err != nil {
		b.log.Error(err.Error(), zap.Any("roundID", roundID), zap.Any("status", status))
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.BetRound{}, err
	}
	return dto.BetRound{
		ID:         betRound.ID,
		Status:     string(betRound.Status),
		CrashPoint: betRound.CrashPoint,
		CreatedAt:  &betRound.CreatedAt,
		ClosedAt:   &betRound.ClosedAt.Time,
	}, nil
}

func (b *bet) CloseRound(ctx context.Context, roundID uuid.UUID) (dto.BetRound, error) {
	betRound, err := b.db.Queries.CloseRoundByID(ctx, db.CloseRoundByIDParams{
		ClosedAt: sql.NullTime{Time: time.Now(), Valid: true},
		ID:       roundID,
	})
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.BetRound{}, err
	}
	return dto.BetRound{
		ID:         betRound.ID,
		Status:     string(betRound.Status),
		CrashPoint: betRound.CrashPoint,
		CreatedAt:  &betRound.CreatedAt,
		ClosedAt:   &betRound.ClosedAt.Time,
	}, nil
}

func (b *bet) GetRoundByID(ctx context.Context, roundID uuid.UUID) (dto.BetRound, bool, error) {
	betRound, err := b.db.Queries.GetBetRoundByID(ctx, roundID)
	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error(), zap.Any("roundID", roundID))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.BetRound{}, false, err
	}
	if err != nil {
		return dto.BetRound{}, false, nil
	}
	return dto.BetRound{
		ID:         betRound.ID,
		Status:     string(betRound.Status),
		CrashPoint: betRound.CrashPoint,
		CreatedAt:  &betRound.CreatedAt,
		ClosedAt:   &betRound.ClosedAt.Time,
	}, true, nil
}

func (b *bet) GetLeaders(ctx context.Context) (dto.LeadersResp, error) {
	var leadersRes dto.LeadersResp
	leaders, err := b.db.Queries.GetLeaders(ctx)
	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.LeadersResp{}, err
	}
	if err != nil {
		return dto.LeadersResp{}, nil
	}
	for _, leader := range leaders {
		leadersRes.Leaders = append(leadersRes.Leaders, dto.Leader{
			ProfileURL: leader.Profile.String,
			Payout:     leader.TotalCashOut,
		})
		leadersRes.TotalPlayers = int(leader.TotalPlayers)
	}

	return leadersRes, nil
}

func (b *bet) GetFailedRounds(ctx context.Context) ([]dto.BetRound, error) {
	var resp []dto.BetRound
	rounds, err := b.db.Queries.GetFailedRounds(ctx)
	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return []dto.BetRound{}, err
	}
	for _, round := range rounds {
		resp = append(resp, dto.BetRound{
			ID:         round.ID,
			Status:     string(round.Status),
			CrashPoint: round.CrashPoint,
			UserID:     round.UserID,
			Currency:   round.Currency,
			Amount:     round.Amount,
			BetID:      round.BetID,
			CreatedAt:  &round.CreatedAt,
			ClosedAt:   &round.ClosedAt.Time,
		})
	}
	return resp, nil
}

func (b *bet) SaveFailedBetsLogaAuto(ctx context.Context, req dto.SaveFailedBetsLog) error {
	adminID := utils.NullUUID(req.AdminID)
	_, err := b.db.Queries.SaveFailedBetsLogAuto(ctx, db.SaveFailedBetsLogAutoParams{
		UserID:        req.UserID,
		RoundID:       req.RoundID,
		BetID:         req.BetID,
		Status:        req.Status,
		Manual:        req.Manual,
		AdminID:       adminID,
		CreatedAt:     time.Now(),
		TransactionID: req.TransactionID,
	})
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return err
	}
	return nil
}

func (b *bet) GetUserFundByUserIDAndRoundID(ctx context.Context, userID, roundID uuid.UUID) (dto.SaveFailedBetsLog, bool, error) {
	usr, err := b.db.Queries.GetUserFundByUserIDAndRoundID(ctx, db.GetUserFundByUserIDAndRoundIDParams{
		UserID:  userID,
		RoundID: roundID,
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.SaveFailedBetsLog{}, false, err
	}

	if err != nil {
		return dto.SaveFailedBetsLog{}, false, nil
	}

	return dto.SaveFailedBetsLog{
		UserID:        usr.UserID,
		RoundID:       usr.RoundID,
		BetID:         usr.BetID,
		AdminID:       usr.AdminID.UUID,
		Manual:        usr.Manual,
		Status:        usr.Status,
		CreatedAt:     usr.CreatedAt,
		TransactionID: usr.TransactionID,
	}, false, nil
}

func (b *bet) GetAllFailedRounds(ctx context.Context, req dto.GetFailedRoundsReq) (dto.GetFailedRoundsRes, error) {
	var resp dto.GetFailedRoundsRes
	res, err := b.db.Queries.GetAllFailedRounds(ctx, db.GetAllFailedRoundsParams{
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())

	}

	for _, rs := range res {
		failedLogs := dto.FailedBetLogs{
			ID:            rs.FailedBetID,
			UserID:        rs.ID,
			RoundID:       rs.RoundID,
			BetID:         rs.BetID,
			IsManual:      rs.IsManual,
			TransactionID: rs.RefundTransactionID,
			Status:        rs.RefundStatus,
			CreatedAt:     rs.RefundAt,
		}
		resp.Data = append(resp.Data, dto.FailedRoundsResData{
			Round: dto.BetRound{
				ID:         rs.RoundID,
				Status:     constant.ROUND_FAILED,
				CrashPoint: rs.CrashPoint,
				UserID:     rs.ID,
				BetID:      rs.BetID,
				Amount:     rs.Amount,
				Currency:   rs.Currency,
				CreatedAt:  &rs.RoundCreatedAt,
			},
			Bet: dto.Bet{
				BetID:               rs.BetID,
				RoundID:             rs.RoundID,
				UserID:              rs.ID,
				Amount:              rs.Amount,
				Currency:            rs.Currency,
				ClientTransactionID: rs.BetTransactionID,
				Timestamp:           rs.BetTimestamp.Time,
			},
			User: dto.User{
				ID:             rs.ID,
				PhoneNumber:    rs.PhoneNumber.String,
				FirstName:      rs.FirstName.String,
				LastName:       rs.LastName.String,
				Email:          rs.Email.String,
				ProfilePicture: rs.Profile.String,
				DateOfBirth:    rs.DateOfBirth.String,
				Source:         rs.Source.String,
			},
			FailedBetLogs: &failedLogs,
		})
	}
	return resp, nil
}

func (b *bet) GetNotRefundedFailedRounds(ctx context.Context, req dto.GetFailedRoundsReq) (dto.GetFailedRoundsRes, error) {
	var resp dto.GetFailedRoundsRes
	res, err := b.db.Queries.GetUnRefundedFaildedRouns(ctx, db.GetUnRefundedFaildedRounsParams{
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())

	}

	for _, rs := range res {
		resp.Data = append(resp.Data, dto.FailedRoundsResData{
			Round: dto.BetRound{
				ID:         rs.RoundID,
				Status:     constant.ROUND_FAILED,
				CrashPoint: rs.CrashPoint,
				UserID:     rs.ID,
				BetID:      rs.BetID,
				Amount:     rs.Amount,
				Currency:   rs.Currency,
				CreatedAt:  &rs.RoundCreatedAt,
			},
			Bet: dto.Bet{
				BetID:               rs.BetID,
				RoundID:             rs.RoundID,
				UserID:              rs.ID,
				Amount:              rs.Amount,
				Currency:            rs.Currency,
				ClientTransactionID: rs.BetTransactionID,
				Timestamp:           rs.BetTimestamp.Time,
			},
			User: dto.User{
				ID:             rs.ID,
				PhoneNumber:    rs.PhoneNumber.String,
				FirstName:      rs.FirstName.String,
				LastName:       rs.LastName.String,
				Email:          rs.Email.String,
				ProfilePicture: rs.Profile.String,
				DateOfBirth:    rs.DateOfBirth.String,
				Source:         rs.Source.String,
			},
			FailedBetLogs: nil,
		})
	}
	resp.Message = constant.SUCCESS
	return resp, nil
}
