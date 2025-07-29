package bet

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/joshjones612/egyptkingcrash/internal/constant"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/constant/model/db"
	"github.com/shopspring/decimal"
)

func (b *bet) CreateQuickHustleBet(ctx context.Context, req dto.CreateQuickHustleBetReq) (dto.CreateQuickHustelBetResData, error) {
	resp, err := b.db.Queries.CreateQuickHustle(ctx, db.CreateQuickHustleParams{
		UserID:    req.UserID,
		BetAmount: req.BetAmount,
		FirstCard: req.FirstCard,
		Timestamp: time.Now(),
	})

	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.CreateQuickHustelBetResData{}, err
	}

	return dto.CreateQuickHustelBetResData{
		ID:        resp.ID,
		UserID:    resp.UserID,
		BetAmount: resp.BetAmount,
		FirstCard: resp.FirstCard,
		Status:    resp.SecondCard.String,
		Timestamp: resp.Timestamp,
	}, nil
}

func (b *bet) CloseQuickHustelBet(ctx context.Context, req dto.CloseQuickHustleBetData) (dto.QuickHustelBetResData, error) {
	resp, err := b.db.Queries.CloseQuickHustleGame(ctx, db.CloseQuickHustleGameParams{
		UserGuessed: sql.NullString{String: req.UserGuess, Valid: true},
		WonStatus:   sql.NullString{String: req.WonStatus, Valid: true},
		SecondCard:  sql.NullString{String: req.SecondCard, Valid: true},
		WonAmount:   decimal.NullDecimal{Decimal: req.WonAmount, Valid: true},
		Status:      constant.CLOSED,
		ID:          req.ID,
	})

	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.QuickHustelBetResData{}, err
	}
	return dto.QuickHustelBetResData{
		ID:          resp.ID,
		UserID:      resp.UserID,
		Status:      resp.Status,
		BetAmount:   resp.BetAmount,
		WonStatus:   resp.WonStatus.String,
		UserGuessed: resp.UserGuessed.String,
		FirstCard:   resp.FirstCard,
		SecondCard:  resp.SecondCard.String,
		Timestamp:   resp.Timestamp,
		WonAmount:   resp.WonAmount.Decimal,
	}, nil
}

func (b *bet) GetQuickHustleByID(ctx context.Context, ID uuid.UUID) (dto.QuickHustelBetResData, bool, error) {
	resp, err := b.db.Queries.GetQuickHustelByID(ctx, ID)
	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.QuickHustelBetResData{}, false, err
	}

	if err != nil {
		return dto.QuickHustelBetResData{}, false, nil
	}

	return dto.QuickHustelBetResData{
		ID:          resp.ID,
		UserID:      resp.UserID,
		Status:      resp.Status,
		BetAmount:   resp.BetAmount,
		WonStatus:   resp.WonStatus.String,
		UserGuessed: resp.UserGuessed.String,
		FirstCard:   resp.FirstCard,
		SecondCard:  resp.SecondCard.String,
		Timestamp:   resp.Timestamp,
		WonAmount:   resp.WonAmount.Decimal,
	}, true, nil
}

func (b *bet) GetQuickHustleBetHistoryByUserID(ctx context.Context, req dto.GetRequest, userID uuid.UUID) (dto.GetQuickHustleResp, bool, error) {
	var data []dto.QuickHustelBetResData
	resp, err := b.db.Queries.GetQuickHustleBetHistoy(ctx, db.GetQuickHustleBetHistoyParams{
		UserID: userID,
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetQuickHustleResp{}, false, err
	}

	if err != nil || len(resp) == 0 {
		return dto.GetQuickHustleResp{}, false, nil
	}
	totalPages := 1
	for i, res := range resp {
		data = append(data, dto.QuickHustelBetResData{
			ID:          res.ID,
			UserID:      res.UserID,
			Status:      res.Status,
			Timestamp:   res.Timestamp,
			BetAmount:   res.BetAmount,
			WonStatus:   res.WonStatus.String,
			UserGuessed: res.UserGuessed.String,
			FirstCard:   res.FirstCard,
			SecondCard:  res.SecondCard.String,
			WonAmount:   res.WonAmount.Decimal,
		})
		if i == 0 {
			totalPages = int(int(res.TotalRows) / req.PerPage)
			if int(res.TotalRows)%req.PerPage != 0 {
				totalPages++
			}
		}
	}
	return dto.GetQuickHustleResp{
		Message: constant.SUCCESS,
		Data: dto.QuickHustleGetHistoryData{
			UserID:     userID,
			TotalPages: totalPages,
			Histories:  data,
		},
	}, true, nil
}
