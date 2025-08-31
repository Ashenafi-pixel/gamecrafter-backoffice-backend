package bet

import (
	"context"
	"database/sql"
	"encoding/json"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/db"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/storage"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type bet struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

func Init(db *persistencedb.PersistenceDB, log *zap.Logger) storage.Bet {
	return &bet{
		db:  db,
		log: log,
	}
}

func (b *bet) SaveUserBet(ctx context.Context, betReq dto.Bet) (dto.Bet, error) {
	betRes, err := b.db.Queries.PlaceBet(ctx, db.PlaceBetParams{
		UserID:              betReq.UserID,
		RoundID:             betReq.RoundID,
		Amount:              betReq.Amount,
		Currency:            betReq.Currency,
		ClientTransactionID: betReq.ClientTransactionID,
		Timestamp:           sql.NullTime{Time: time.Now().In(time.Now().Location()).UTC(), Valid: true},
	})
	if err != nil {
		b.log.Error(err.Error(), zap.Any("betReq", betReq))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.Bet{}, err
	}
	return dto.Bet{
		BetID:               betRes.ID,
		RoundID:             betRes.RoundID,
		UserID:              betRes.UserID,
		Amount:              betRes.Amount,
		Currency:            betRes.Currency,
		ClientTransactionID: betReq.ClientTransactionID,
		Timestamp:           betRes.Timestamp.Time,
	}, nil
}

func (b *bet) GetUserBetByUserIDAndRoundID(ctx context.Context, betReq dto.Bet) ([]dto.Bet, bool, error) {
	var betRes []dto.Bet
	userBets, err := b.db.Queries.GetUserBetByUserIDAndRoundID(ctx, db.GetUserBetByUserIDAndRoundIDParams{
		UserID:  betReq.UserID,
		RoundID: betReq.RoundID,
		Status:  sql.NullString{String: betReq.Status, Valid: true},
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error(), zap.Any("getBetReq", betReq))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.Bet{}, false, err
	}
	if err != nil {
		return []dto.Bet{}, false, nil
	}
	for _, userBet := range userBets {

		betRes = append(betRes, dto.Bet{
			BetID:               userBet.ID,
			RoundID:             userBet.RoundID,
			UserID:              userBet.UserID,
			Amount:              userBet.Amount,
			Currency:            userBet.Currency,
			ClientTransactionID: userBet.ClientTransactionID,
			Timestamp:           userBet.Timestamp.Time,
			Payout:              userBet.Payout.Decimal,
			CashOutMultiplier:   userBet.CashOutMultiplier.Decimal,
		})
	}
	return betRes, true, nil
}

func (b *bet) CashOut(ctx context.Context, cashOut dto.SaveCashoutReq) (dto.Bet, error) {
	cashoutRes, err := b.db.Queries.CashOut(ctx, db.CashOutParams{
		CashOutMultiplier: decimal.NullDecimal{Decimal: cashOut.Multiplier, Valid: true},
		Payout:            decimal.NullDecimal{Decimal: cashOut.Payout, Valid: true},
		Timestamp:         sql.NullTime{Time: time.Now().In(time.Now().Location()).UTC(), Valid: true},
		ID:                cashOut.ID,
	})
	if err != nil {
		b.log.Error(err.Error(), zap.Any("cashOutReq", cashOut))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.Bet{}, err
	}
	return dto.Bet{
		BetID:               cashoutRes.ID,
		RoundID:             cashoutRes.RoundID,
		UserID:              cashoutRes.UserID,
		Amount:              cashoutRes.Amount,
		Currency:            cashoutRes.Currency,
		ClientTransactionID: cashoutRes.ClientTransactionID,
		Timestamp:           cashoutRes.Timestamp.Time,
		Payout:              cashoutRes.Payout.Decimal,
		CashOutMultiplier:   cashoutRes.CashOutMultiplier.Decimal,
	}, err
}

func (b *bet) ReverseCashOut(ctx context.Context, betID uuid.UUID) error {
	_, err := b.db.Queries.ReverseCashOut(ctx, db.ReverseCashOutParams{
		Timestamp: sql.NullTime{Time: time.Now().In(time.Now().Location()).UTC(), Valid: true},
		ID:        uuid.UUID(betID),
	})
	return err
}

func (b *bet) GetBetHistory(ctx context.Context, getBetHistoryReq dto.GetBetHistoryReq) (dto.BetHistoryResp, bool, error) {
	if getBetHistoryReq.UserID == uuid.Nil {
		return b.GetAllBetHisotry(ctx, getBetHistoryReq)
	} else {
		return b.GetBetHistoryByUserID(ctx, getBetHistoryReq)
	}
}

func (b *bet) GetAllBetHisotry(ctx context.Context, getBetHistoryReq dto.GetBetHistoryReq) (dto.BetHistoryResp, bool, error) {
	var betRes []db.GetBetHistoryRow
	var totalPage int64
	var bestResp []dto.BetRes
	var err error

	betRes, err = b.db.Queries.GetBetHistory(ctx, db.GetBetHistoryParams{
		Offset: int32(getBetHistoryReq.Offset),
		Limit:  int32(getBetHistoryReq.PerPage),
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error(), zap.Any("getBetReq", getBetHistoryReq))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.BetHistoryResp{}, false, err
	}
	if err != nil {
		return dto.BetHistoryResp{}, false, nil
	}
	for _, game := range betRes {
		var bets []dto.BetRes
		if err := json.Unmarshal(game.Bets.Bytes, &bets); err != nil {
			b.log.Error(err.Error(), zap.Any("game", game))
			err = errors.ErrInternalServerError.Wrap(err, err.Error())
			return dto.BetHistoryResp{}, false, err
		}
		bestResp = append(bestResp, bets...)
		totalPage = game.Total
	}
	ps := float64(totalPage / int64(getBetHistoryReq.PerPage))
	total := int(math.Ceil(ps))
	return dto.BetHistoryResp{
		Status: constant.SUCCESS,
		Data: dto.BetHisotryData{
			Page:       getBetHistoryReq.Page,
			TotalPages: total,
			Bets:       bestResp,
		},
	}, true, err
}

func (b *bet) GetBetHistoryByUserID(ctx context.Context, getBetHistoryReq dto.GetBetHistoryReq) (dto.BetHistoryResp, bool, error) {
	var betRes []db.GetBetHistoryByUserIDRow
	var totalPage int64
	var betsResp []dto.BetRes
	var err error

	betRes, err = b.db.Queries.GetBetHistoryByUserID(ctx, db.GetBetHistoryByUserIDParams{
		Offset: int32(getBetHistoryReq.Offset),
		Limit:  int32(getBetHistoryReq.PerPage),
		UserID: getBetHistoryReq.UserID,
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error(), zap.Any("getBetReq", getBetHistoryReq))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.BetHistoryResp{}, false, err
	}
	if err != nil {
		return dto.BetHistoryResp{}, false, nil
	}
	for _, game := range betRes {
		var bets []dto.BetRes
		if err := json.Unmarshal(game.Bets.Bytes, &bets); err != nil {
			b.log.Error(err.Error(), zap.Any("game", game))
			err = errors.ErrInternalServerError.Wrap(err, err.Error())
			return dto.BetHistoryResp{}, false, err
		}
		betsResp = append(betsResp, bets...)
		totalPage = game.Total
	}
	ps := float64(totalPage / int64(getBetHistoryReq.PerPage))
	total := int(math.Ceil(ps))
	return dto.BetHistoryResp{
		Status: constant.SUCCESS,
		Data: dto.BetHisotryData{
			Page:       getBetHistoryReq.Page,
			TotalPages: total,
			Bets:       betsResp,
		},
	}, true, err
}

func (b *bet) UpdateBetStatus(ctx context.Context, betID uuid.UUID, status string) (dto.Bet, error) {
	betRes, err := b.db.Queries.UpdateBetStatus(ctx, db.UpdateBetStatusParams{
		Status: sql.NullString{String: status, Valid: true},
		ID:     betID,
	})
	if err != nil {
		return dto.Bet{}, err
	}
	return dto.Bet{
		BetID:               betRes.ID,
		RoundID:             betRes.RoundID,
		UserID:              betRes.UserID,
		Amount:              betRes.Amount,
		Currency:            betRes.Currency,
		ClientTransactionID: betRes.ClientTransactionID,
		Timestamp:           betRes.Timestamp.Time,
		Payout:              betRes.Payout.Decimal,
		CashOutMultiplier:   betRes.CashOutMultiplier.Decimal,
		Status:              betRes.Status.String,
	}, nil
}

func (b *bet) GetUserActiveBetWithRound(ctx context.Context, userID, roundID uuid.UUID) (dto.BetRound, bool, error) {
	activeBet, err := b.db.Queries.GetUserActiveBetWithRound(ctx, db.GetUserActiveBetWithRoundParams{
		UserID:  userID,
		RoundID: roundID,
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.BetRound{}, false, err
	}

	if err != nil {
		return dto.BetRound{}, false, nil
	}

	return dto.BetRound{
		ID:        activeBet.ID,
		Status:    activeBet.Status.String,
		Currency:  activeBet.Currency,
		CreatedAt: &activeBet.Timestamp.Time,
		UserID:    activeBet.UserID,
		Amount:    activeBet.Amount,
		BetID:     activeBet.RoundID,
	}, true, nil
}
