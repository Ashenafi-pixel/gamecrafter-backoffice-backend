package bet

import (
	"context"
	"math"

	"github.com/google/uuid"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/constant/model/db"
	"github.com/joshjones612/egyptkingcrash/platform/utils"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

func (b *bet) SavePlinkoBet(ctx context.Context, req dto.PlacePlinkoGame) (dto.PlacePlinkoGame, error) {
	path := utils.ConvertIntMapToString(req.DropPath)
	resp, err := b.db.Queries.SavePlinkoBet(ctx, db.SavePlinkoBetParams{
		UserID:        req.UserID,
		BetAmount:     req.BetAmount,
		DropPath:      path,
		Multiplier:    decimal.NullDecimal{Decimal: req.Multiplier, Valid: true},
		WinAmount:     decimal.NullDecimal{Decimal: req.WinAmount, Valid: true},
		Finalposition: decimal.NullDecimal{Decimal: decimal.NewFromInt(int64(req.FinalPosition)), Valid: true},
		Timestamp:     req.Timestamp,
	})

	if err != nil {
		b.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.PlacePlinkoGame{}, err
	}

	return dto.PlacePlinkoGame{
		ID:            resp.ID,
		UserID:        resp.UserID,
		Timestamp:     resp.Timestamp,
		BetAmount:     resp.BetAmount,
		DropPath:      req.DropPath,
		Multiplier:    resp.Multiplier.Decimal,
		WinAmount:     resp.WinAmount.Decimal,
		FinalPosition: int(resp.Finalposition.Decimal.IntPart()),
	}, nil
}

func (b *bet) GetUserPlinkoBetHistoryByID(ctx context.Context, req dto.PlinkoBetHistoryReq) (dto.PlinkoBetHistoryRes, bool, error) {
	games := []dto.PlacePlinkoGame{}

	resps, err := b.db.Queries.GetPlinkoBetHistoryByUserID(ctx, db.GetPlinkoBetHistoryByUserIDParams{
		UserID: req.UserID,
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.PlinkoBetHistoryRes{}, false, err
	}

	if err != nil {
		return dto.PlinkoBetHistoryRes{}, false, nil
	}

	for _, resp := range resps {
		dropPath, err := utils.ConvertFromStringToIntMapMap(resp.DropPath)
		if err != nil {
			b.log.Error(err.Error(), zap.Any("resp", resp))
			continue
		}

		games = append(games, dto.PlacePlinkoGame{
			ID:            resp.ID,
			UserID:        resp.UserID,
			Timestamp:     resp.Timestamp,
			BetAmount:     resp.BetAmount,
			Multiplier:    resp.Multiplier.Decimal,
			WinAmount:     resp.WinAmount.Decimal,
			FinalPosition: int(resp.Finalposition.Decimal.IntPart()),
			DropPath:      dropPath,
		})

	}

	total, err := b.db.Queries.CountPlinkoBetHistoryByUserID(ctx, req.UserID)
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.PlinkoBetHistoryRes{}, false, err
	}

	ps := float64(total / int64(req.PerPage))
	totalPages := int(math.Ceil(ps))
	return dto.PlinkoBetHistoryRes{
		TotalPages: totalPages,
		Games:      games,
	}, true, nil
}

func (b *bet) GetPlinkoGameStats(ctx context.Context, userID uuid.UUID) (dto.PlinkoGameStatRes, bool, error) {

	resp, err := b.db.Queries.PlinkoGameState(ctx, userID)
	if err != nil && err.Error() == dto.ErrNoRows {
		b.log.Error(err.Error(), zap.Any("user_id", userID))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.PlinkoGameStatRes{}, false, err
	}

	if err != nil {
		return dto.PlinkoGameStatRes{}, false, nil
	}
	highestWin, err := b.db.Queries.GetUserHighestPlinkoBet(ctx, userID)
	if err != nil && err.Error() == dto.ErrNoRows {
		b.log.Error(err.Error(), zap.Any("user_id", userID))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.PlinkoGameStatRes{}, false, err
	}

	if err != nil {
		return dto.PlinkoGameStatRes{}, false, nil
	}

	return dto.PlinkoGameStatRes{
		TotalGames:        int(resp.TotalGames),
		TotalWon:          resp.TotalWin,
		TotalWagered:      resp.TotalWagered,
		NetProfit:         resp.NetProfit,
		AverageMultiplier: resp.AverageMultiplier,
		HighestWin: dto.HighestWin{
			ID:         highestWin.ID,
			Timestamp:  highestWin.Timestamp,
			BetAmount:  highestWin.BetAmount,
			Multiplier: highestWin.Multiplier.Decimal,
			WinAmount:  highestWin.WinAmount.Decimal,
			PinCount:   decimal.NewFromInt(10),
		},
	}, true, nil
}
