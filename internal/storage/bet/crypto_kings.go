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
	"github.com/shopspring/decimal"
)

func (b *bet) CreateCryptoKings(ctx context.Context, req dto.CreateCryptoKingData) (dto.CreateCryptoKingData, error) {
	var rangeStartValue decimal.NullDecimal
	var rangeEndValue decimal.NullDecimal

	if req.Type == constant.CRYPTO_KING_RANGE {
		rangeEndValue = decimal.NullDecimal{Decimal: req.SelectedEndValue, Valid: true}
		rangeStartValue = decimal.NullDecimal{Decimal: req.SelectedStartValue, Valid: true}
	}

	resp, err := b.db.Queries.CreateCryptoKings(ctx, db.CreateCryptoKingsParams{
		UserID:             req.UserID,
		Status:             constant.CLOSED,
		BetAmount:          decimal.NewFromInt(req.BetAmount),
		WonAmount:          decimal.NullDecimal{Decimal: req.WonAmount, Valid: true},
		StartCryptoValue:   req.StartCryptoValue,
		EndCryptoValue:     req.EndCryptoValue,
		SelectedEndSecond:  sql.NullInt32{Int32: int32(req.SelectedEndSecond), Valid: true},
		Type:               req.Type,
		SelectedStartValue: rangeStartValue,
		SelectedEndValue:   rangeEndValue,
		WonStatus:          req.WonStatus,
		Timestamp:          time.Now(),
	})

	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.CreateCryptoKingData{}, err
	}

	return dto.CreateCryptoKingData{
		ID:                 resp.ID,
		UserID:             resp.UserID,
		Status:             resp.Status,
		BetAmount:          resp.BetAmount.IntPart(),
		WonAmount:          resp.WonAmount.Decimal,
		StartCryptoValue:   resp.StartCryptoValue,
		EndCryptoValue:     resp.EndCryptoValue,
		SelectedEndSecond:  int(resp.SelectedEndSecond.Int32),
		SelectedStartValue: resp.SelectedStartValue.Decimal,
		SelectedEndValue:   resp.SelectedEndValue.Decimal,
		WonStatus:          resp.WonStatus,
		Type:               resp.Type,
		Timestamp:          resp.Timestamp,
	}, nil
}

func (b *bet) GetCrytoKingsBetHistoryByUserID(ctx context.Context, req dto.GetRequest, userID uuid.UUID) (dto.GetCryptoKingsUserBetHistoryRes, bool, error) {
	var data []dto.CreateCryptoKingData
	resp, err := b.db.Queries.GetCrytoKingsBetHistoryByUserID(ctx, db.GetCrytoKingsBetHistoryByUserIDParams{
		UserID: userID,
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetCryptoKingsUserBetHistoryRes{}, false, err
	}

	if err != nil || len(resp) == 0 {
		return dto.GetCryptoKingsUserBetHistoryRes{}, false, nil
	}
	totalPages := 1
	for i, res := range resp {
		data = append(data, dto.CreateCryptoKingData{
			ID:                 res.ID,
			UserID:             res.UserID,
			Status:             res.Status,
			BetAmount:          res.BetAmount.IntPart(),
			WonAmount:          res.WonAmount.Decimal,
			StartCryptoValue:   res.StartCryptoValue,
			EndCryptoValue:     res.EndCryptoValue,
			SelectedEndSecond:  int(res.SelectedEndSecond.Int32),
			SelectedStartValue: res.SelectedStartValue.Decimal,
			SelectedEndValue:   res.SelectedEndValue.Decimal,
			WonStatus:          res.WonStatus,
			Type:               res.Type,
			Timestamp:          res.Timestamp,
		})
		if i == 0 {
			totalPages = int(int(res.TotalRows) / req.PerPage)
			if int(res.TotalRows)%req.PerPage != 0 {
				totalPages++
			}
		}
	}
	return dto.GetCryptoKingsUserBetHistoryRes{
		Message: constant.SUCCESS,
		Data: dto.GetCryptoKingsUserBetHistoryResData{
			UserID:     userID,
			TotalPages: totalPages,
			Histories:  data,
		},
	}, true, nil
}
