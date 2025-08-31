package bet

import (
	"context"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/db"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

func (b *bet) CreateStreetKingsGame(ctx context.Context, req dto.CreateStreetKingsReqData) (dto.CreateStreetKingsRespData, error) {
	resp, err := b.db.CreateStreetKingsGame(ctx, db.CreateStreetKingsGameParams{
		UserID:     req.UserID,
		Version:    req.CreateCrashKingsReq.Version,
		BetAmount:  req.CreateCrashKingsReq.BetAmount,
		CrashPoint: req.CrashPoint,
		Timestamp:  req.Timestamp,
	})

	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.CreateStreetKingsRespData{}, err
	}
	return dto.CreateStreetKingsRespData{
		ID:         resp.ID,
		Version:    resp.Version,
		BetAmount:  resp.BetAmount,
		CrashPoint: resp.CrashPoint,
		Status:     resp.Status,
		Timestamp:  resp.Timestamp,
	}, nil
}

func (b *bet) CloseStreetKingsCrash(ctx context.Context, req dto.StreetKingsCashoutReq) {
	var wonOrNull decimal.NullDecimal
	var cashOutPoint decimal.NullDecimal
	if req.WonAmount.GreaterThan(decimal.Zero) {
		wonOrNull = decimal.NullDecimal{Decimal: req.WonAmount, Valid: true}
		cashOutPoint = decimal.NullDecimal{Decimal: req.CashoutPoint, Valid: true}
	}
	b.db.Queries.CloseStreetKingGameByID(ctx, db.CloseStreetKingGameByIDParams{
		Status:       req.Status,
		WonAmount:    wonOrNull,
		CashOutPoint: cashOutPoint,
		ID:           req.ID,
	})
}

func (b *bet) GetStreetKingsCrashByID(ctx context.Context, ID uuid.UUID) (dto.StreetKingsCrashResp, error) {
	resp, err := b.db.Queries.GetStreetKingsGameByID(ctx, ID)
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.StreetKingsCrashResp{}, err
	}
	return dto.StreetKingsCrashResp{
		Message: constant.SUCCESS,
		Data: dto.StreetKingsCrashRespData{
			ID:           resp.ID,
			UserID:       resp.UserID,
			Status:       resp.Status,
			Version:      resp.Version,
			BetAmount:    resp.BetAmount,
			WonAmount:    resp.WonAmount.Decimal,
			CrashPoint:   resp.CrashPoint,
			CashoutPoint: resp.CashOutPoint.Decimal,
			Timestamp:    resp.Timestamp,
		},
	}, err
}

func (b *bet) GetStreetKingsGamesByUserIDAndVersion(ctx context.Context, req dto.GetStreetkingHistoryReq, userID uuid.UUID) (dto.GetStreetkingHistoryRes, bool, error) {
	var respData []dto.StreetKingsCrashRespData
	resps, err := b.db.Queries.GetStreetKingsGamesByUserIDAndVersion(ctx, db.GetStreetKingsGamesByUserIDAndVersionParams{
		UserID:  userID,
		Version: req.Version,
		Limit:   int32(req.PerPage),
		Offset:  int32(req.Page),
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetStreetkingHistoryRes{}, false, err
	}

	if err != nil || len(resps) == 0 {
		return dto.GetStreetkingHistoryRes{}, false, nil
	}
	totalPages := 1
	for i, resp := range resps {
		wonStats := "WON"
		if resp.WonAmount.Decimal.Equal(decimal.Zero) {
			wonStats = "LOSE"
		}
		respData = append(respData, dto.StreetKingsCrashRespData{
			ID:           resp.ID,
			UserID:       resp.UserID,
			Status:       resp.Status,
			Version:      resp.Version,
			BetAmount:    resp.BetAmount,
			WonAmount:    resp.WonAmount.Decimal,
			CrashPoint:   resp.CrashPoint,
			CashoutPoint: resp.CashOutPoint.Decimal,
			Timestamp:    resp.Timestamp,
			WonStatus:    wonStats,
		})
		if i == 0 {
			totalPages = int(int(resp.TotalRows) / req.PerPage)
			if int(resp.TotalRows)%req.PerPage != 0 {
				totalPages++
			}
		}
	}

	return dto.GetStreetkingHistoryRes{
		Message:    constant.SUCCESS,
		TotalPages: totalPages,
		Data:       respData,
	}, true, nil

}
