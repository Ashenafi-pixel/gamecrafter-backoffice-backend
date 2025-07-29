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

func (b *bet) CreateScratchCardsBet(ctx context.Context, req dto.ScratchCardsBetData) (dto.ScratchCardsBetData, error) {
	resp, err := b.db.Queries.CreateScratchCardsBet(ctx, db.CreateScratchCardsBetParams{
		UserID:    req.UserID,
		Status:    constant.CLOSED,
		BetAmount: req.BetAmount,
		WonAmount: decimal.NullDecimal{Decimal: req.WonAmount, Valid: true},
		WonStatus: sql.NullString{String: req.WonStatus, Valid: true},
		Timestamp: time.Now(),
	})

	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.ScratchCardsBetData{}, err
	}

	return dto.ScratchCardsBetData{
		ID:        resp.ID,
		UserID:    resp.UserID,
		Status:    resp.Status,
		BetAmount: resp.BetAmount,
		WonStatus: resp.WonStatus.String,
		Timestamp: resp.Timestamp,
		WonAmount: resp.WonAmount.Decimal,
	}, nil
}

func (b *bet) GetUserScratchCardBetHistories(ctx context.Context, req dto.GetRequest, userID uuid.UUID) (dto.GetScratchBetHistoriesResp, bool, error) {
	var resp []dto.ScratchCardsBetData
	dbResp, err := b.db.Queries.GetUserScratchCardBetHistories(ctx, db.GetUserScratchCardBetHistoriesParams{
		UserID: userID,
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.GetScratchBetHistoriesResp{}, false, err
	}

	if err != nil || len(dbResp) == 0 {
		return dto.GetScratchBetHistoriesResp{
			Message: constant.SUCCESS,
			Data: dto.ScratchCardsBetRespData{
				TotalPages: 1,
				UserID:     userID,
				Histories:  []dto.ScratchCardsBetData{},
			},
		}, false, nil
	}

	totalPages := 1
	for i, r := range dbResp {
		resp = append(resp, dto.ScratchCardsBetData{
			ID:        r.ID,
			UserID:    r.UserID,
			Status:    r.Status,
			BetAmount: r.BetAmount,
			WonStatus: r.WonStatus.String,
			Timestamp: r.Timestamp,
			WonAmount: r.WonAmount.Decimal,
		})
		if i == 0 {
			totalPages = int(int(r.TotalRows) / req.PerPage)
			if int(r.TotalRows)%req.PerPage != 0 {
				totalPages++
			}
		}

	}

	return dto.GetScratchBetHistoriesResp{
		Message: constant.SUCCESS,
		Data: dto.ScratchCardsBetRespData{
			TotalPages: totalPages,
			UserID:     userID,
			Histories:  resp,
		},
	}, true, nil
}
