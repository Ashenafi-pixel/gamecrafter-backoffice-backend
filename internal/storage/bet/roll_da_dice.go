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

func (b *bet) CreateRollDaDice(ctx context.Context, req dto.CreateRollDaDiceReq) (dto.RollDaDiceData, error) {
	resp, err := b.db.Queries.CreateRollDaDice(ctx, db.CreateRollDaDiceParams{
		UserID:                req.UserID,
		BetAmount:             req.BetAmount,
		Multiplier:            decimal.NullDecimal{Decimal: req.Multiplier, Valid: true},
		Status:                constant.CLOSED,
		WonAmount:             decimal.NullDecimal{Decimal: req.WonAmount, Valid: true},
		CrashPoint:            req.CrashPoint,
		UserGuessedStartPoint: decimal.NullDecimal{Decimal: req.UserGuessStartPoint, Valid: true},
		UserGuessedEndPoint:   decimal.NullDecimal{Decimal: req.UserGuessEndPoint, Valid: true},
		Timestamp:             time.Now(),
		WonStatus:             sql.NullString{String: req.WonStatus, Valid: true},
	})

	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.RollDaDiceData{}, err
	}

	return dto.RollDaDiceData{
		ID:                  resp.ID,
		UserID:              resp.UserID,
		BetAmount:           resp.BetAmount,
		CrashPoint:          resp.CrashPoint,
		Timestamp:           resp.Timestamp,
		UserGuessStartPoint: resp.UserGuessedStartPoint.Decimal,
		UserGuessEndPoint:   resp.UserGuessedEndPoint.Decimal,
		Status:              resp.Status,
		WonStatus:           resp.WonStatus.String,
		WonAmount:           resp.WonAmount.Decimal,
		Multiplier:          resp.Multiplier.Decimal,
	}, nil
}

func (b *bet) GetUserRollDicBetHistory(ctx context.Context, req dto.GetRequest, userID uuid.UUID) (dto.GetRollDaDiceRespData, bool, error) {
	var data []dto.RollDaDiceData
	resp, err := b.db.Queries.GetRollDaDiceHistory(ctx, db.GetRollDaDiceHistoryParams{
		UserID: userID,
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.GetRollDaDiceRespData{}, false, err
	}

	if err != nil || len(resp) == 0 {
		return dto.GetRollDaDiceRespData{}, false, nil
	}
	totalPages := 1
	for i, r := range resp {
		data = append(data, dto.RollDaDiceData{
			ID:                  r.ID,
			UserID:              r.UserID,
			BetAmount:           r.BetAmount,
			Timestamp:           r.Timestamp,
			UserGuessStartPoint: r.UserGuessedEndPoint.Decimal,
			UserGuessEndPoint:   r.UserGuessedEndPoint.Decimal,
			Multiplier:          r.Multiplier.Decimal,
			Status:              r.Status,
			CrashPoint:          r.CrashPoint,
			WonStatus:           r.WonStatus.String,
			WonAmount:           r.WonAmount.Decimal,
		})
		if i == 0 {
			totalPages = int(int(r.TotalRows) / req.PerPage)
			if int(r.TotalRows)%req.PerPage != 0 {
				totalPages++
			}
		}
	}
	return dto.GetRollDaDiceRespData{
		TotalPages: totalPages,
		UserID:     userID,
		Histories:  data,
	}, true, nil
}
