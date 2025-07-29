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
)

func (b *bet) CreateSpinningWheel(ctx context.Context, req dto.SpinningWheelData) (dto.SpinningWheelData, error) {

	resp, err := b.db.Queries.CreateSpinningWheel(ctx, db.CreateSpinningWheelParams{
		UserID:    req.UserID,
		Status:    req.Status,
		BetAmount: req.BetAmount,
		Timestamp: time.Now(),
		WonStatus: sql.NullString{String: req.WonStatus, Valid: true},
		WonAmount: sql.NullString{String: req.WonAmount, Valid: true},
		Type:      sql.NullString{String: req.Type, Valid: true},
	})

	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.SpinningWheelData{}, err
	}

	return dto.SpinningWheelData{
		ID:        resp.ID,
		UserID:    resp.UserID,
		Status:    resp.Status,
		BetAmount: resp.BetAmount,
		WonAmount: resp.WonAmount.String,
		Timestamp: resp.Timestamp,
	}, nil
}

func (b *bet) GetSpinningWheelUserBetHistory(ctx context.Context, req dto.GetRequest, userID uuid.UUID) (dto.GetSpinningWheelHistoryResp, bool, error) {
	var data []dto.SpinningWheelData

	resp, err := b.db.Queries.GetSpinningWheelUserHistory(ctx, db.GetSpinningWheelUserHistoryParams{
		UserID: userID,
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.GetSpinningWheelHistoryResp{}, false, err
	}

	if err != nil || len(resp) == 0 {
		return dto.GetSpinningWheelHistoryResp{
			Message: constant.SUCCESS,
			Data: dto.GetSpinningWheelData{
				TotalPages: 1,
				Histories:  []dto.SpinningWheelData{},
			},
		}, false, nil
	}

	totalPages := 1
	for i, r := range resp {
		data = append(data, dto.SpinningWheelData{
			ID:        r.ID,
			UserID:    r.UserID,
			Status:    r.Status,
			BetAmount: r.BetAmount,
			WonStatus: r.WonStatus.String,
			Timestamp: r.Timestamp,
			WonAmount: r.WonAmount.String,
			Type:      r.Type.String,
		})
		if i == 0 {
			totalPages = int(int(r.TotalRows) / req.PerPage)
			if int(r.TotalRows)%req.PerPage != 0 {
				totalPages++
			}
		}

	}

	return dto.GetSpinningWheelHistoryResp{
		Message: constant.SUCCESS,
		Data: dto.GetSpinningWheelData{
			TotalPages: totalPages,
			Histories:  data,
		},
	}, true, nil
}

// create spinning wheel bet mystery

func (b *bet) CreateSpinningWheelMystery(ctx context.Context, req dto.CreateSpinningWheelMysteryReq) (dto.CreateSpinningWheelMysteryRes, error) {

	resp, err := b.db.Queries.CreateSpinningWheelMysteries(ctx, db.CreateSpinningWheelMysteriesParams{
		Name:      req.Name,
		Amount:    req.Amount,
		Type:      db.Spinningwheelmysterytypes(req.Type),
		Status:    req.Status,
		Frequency: int32(req.Frequency),
		CreatedBy: req.CreatedBy,
		CreatedAt: time.Now(),
		Icon:      req.Icon,
	})

	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.CreateSpinningWheelMysteryRes{}, err
	}

	return dto.CreateSpinningWheelMysteryRes{
		Message: constant.SUCCESS,
		Data: dto.SpinningWheelMysteryResData{
			ID:        resp.ID,
			Amount:    resp.Amount,
			Frequency: int(resp.Frequency),
			Type:      dto.SpinningWheelMysteryTypes(resp.Type),
			Status:    resp.Status,
			Icon:      resp.Icon,
		},
	}, nil
}

func (b *bet) GetSpinningWheelMystery(ctx context.Context, req dto.GetRequest) (dto.GetSpinningWheelMysteryRes, error) {
	var data []dto.SpinningWheelMysteryResData
	resp, err := b.db.Queries.GetSpinningWheelMysteries(ctx, db.GetSpinningWheelMysteriesParams{
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.GetSpinningWheelMysteryRes{}, err
	}
	if err != nil {
		return dto.GetSpinningWheelMysteryRes{}, err
	}
	totalPages := 1
	for i, r := range resp {
		data = append(data, dto.SpinningWheelMysteryResData{
			ID:        r.ID,
			Amount:    r.Amount,
			Frequency: int(r.Frequency),
			Type:      dto.SpinningWheelMysteryTypes(r.Type),
			Status:    r.Status,
			Icon:      r.Icon,
		})
		if i == 0 {
			totalPages = int(int(r.TotalRows) / req.PerPage)
			if int(r.TotalRows)%req.PerPage != 0 {
				totalPages++
			}
		}
	}

	return dto.GetSpinningWheelMysteryRes{
		Message:   constant.SUCCESS,
		TotalPage: totalPages,
		Data:      data,
	}, nil
}

func (b *bet) UpdateSpinningWheelMystery(ctx context.Context, req dto.UpdateSpinningWheelMysteryReq) (dto.UpdateSpinningWheelMysteryRes, error) {
	resp, err := b.db.Queries.UpdateSpinningWheelMystery(ctx, db.UpdateSpinningWheelMysteryParams{
		ID:        req.ID,
		Amount:    req.Amount,
		Type:      db.Spinningwheelmysterytypes(req.Type),
		Status:    req.Status,
		Frequency: int32(req.Frequency),
	})

	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.UpdateSpinningWheelMysteryRes{}, err
	}

	return dto.UpdateSpinningWheelMysteryRes{
		Message: constant.SUCCESS,
		Data: dto.SpinningWheelMysteryResData{
			ID:        resp.ID,
			Amount:    resp.Amount,
			Frequency: int(resp.Frequency),
			Type:      dto.SpinningWheelMysteryTypes(resp.Type),
			Status:    resp.Status,
			Icon:      resp.Icon,
		},
	}, nil
}

func (b *bet) DeleteSpinningWheelMystery(ctx context.Context, id uuid.UUID) error {
	err := b.db.Queries.DeleteSpinningWheelMystery(ctx, id)

	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return err
	}
	return nil
}

func (b *bet) CreateSpinningWheelConfig(ctx context.Context, req dto.CreateSpinningWheelConfigReq) (dto.CreateSpinningWheelConfigRes, error) {
	resp, err := b.db.Queries.CreateSpinningWheelConfig(ctx, db.CreateSpinningWheelConfigParams{
		CreatedAt: time.Now(),
		CreatedBy: req.CreatedBy,
		Name:      req.Name,
		Amount:    req.Amount,
		Type:      db.Spinningwheeltypes(req.Type),
		Frequency: int32(req.Frequency),
		Icon:      req.Icon,
		Color:     req.Color,
	})

	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.CreateSpinningWheelConfigRes{}, err
	}

	return dto.CreateSpinningWheelConfigRes{
		Message: constant.SUCCESS,
		Data: dto.SpinningWheelConfigData{
			ID:        resp.ID,
			Name:      resp.Name,
			Amount:    resp.Amount,
			Type:      dto.SpinningWheelTypes(resp.Type),
			Frequency: int(resp.Frequency),
			Icon:      resp.Icon,
			Color:     resp.Color,
		},
	}, nil
}

func (b *bet) GetSpinningWheelConfig(ctx context.Context, req dto.GetRequest) (dto.GetSpinningWheelConfigRes, error) {
	var data []dto.SpinningWheelConfigData
	resp, err := b.db.Queries.GetSpinningWheelConfigs(ctx, db.GetSpinningWheelConfigsParams{
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.GetSpinningWheelConfigRes{}, err
	}
	if err != nil {
		return dto.GetSpinningWheelConfigRes{}, err
	}
	totalPages := 1
	for i, r := range resp {
		data = append(data, dto.SpinningWheelConfigData{
			ID:        r.ID,
			Name:      r.Name,
			Amount:    r.Amount,
			Type:      dto.SpinningWheelTypes(r.Type),
			Frequency: int(r.Frequency),
			Icon:      r.Icon,
			Color:     r.Color,
		})
		if i == 0 {
			totalPages = int(int(r.TotalRows) / req.PerPage)
			if int(r.TotalRows)%req.PerPage != 0 {
				totalPages++
			}
		}
	}

	return dto.GetSpinningWheelConfigRes{
		Message:                 constant.SUCCESS,
		TotalPage:               totalPages,
		SpinningWheelConfigData: data,
	}, nil
}

func (b *bet) GetAllSpinningWheelConfigs(ctx context.Context) ([]dto.SpinningWheelConfigData, error) {
	resp, err := b.db.Queries.GetAllSpinningWheelConfigs(ctx)
	if err != nil && err.Error() != dto.ErrNoRows {
		return nil, err
	}
	var data []dto.SpinningWheelConfigData
	for _, r := range resp {
		data = append(data, dto.SpinningWheelConfigData{
			ID:        r.ID,
			Name:      r.Name,
			Amount:    r.Amount,
			Status:    r.Status,
			Type:      dto.SpinningWheelTypes(r.Type),
			Frequency: int(r.Frequency),
			Icon:      r.Icon,
		})
	}
	return data, nil
}

func (b *bet) GetAllSpinningWheelMysteries(ctx context.Context) ([]dto.SpinningWheelMysteryResData, error) {
	resp, err := b.db.Queries.GetAllSpinningWheelMysteries(ctx)
	if err != nil && err.Error() != dto.ErrNoRows {
		return nil, err
	}
	var data []dto.SpinningWheelMysteryResData
	for _, r := range resp {
		data = append(data, dto.SpinningWheelMysteryResData{
			ID:        r.ID,
			Amount:    r.Amount,
			Frequency: int(r.Frequency),
			Type:      dto.SpinningWheelMysteryTypes(r.Type),
			Status:    r.Status,
			Icon:      r.Icon,
		})
	}
	return data, nil
}

func (b *bet) DeleteSpinningWheelConfig(ctx context.Context, id uuid.UUID) error {
	err := b.db.Queries.DeleteSpinningWheelConfig(ctx, id)
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return err
	}
	return nil
}

func (b *bet) UpdateSpinningWheelConfig(ctx context.Context, req dto.UpdateSpinningWheelConfigReq) (dto.UpdateSpinningWheelConfigRes, error) {
	resp, err := b.db.Queries.UpdateSpinningWheelConfig(ctx, db.UpdateSpinningWheelConfigParams{
		ID:        req.ID,
		Name:      req.Name,
		Amount:    req.Amount,
		Type:      db.Spinningwheeltypes(req.Type),
		Frequency: int32(req.Frequency),
	})
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.UpdateSpinningWheelConfigRes{}, err
	}
	return dto.UpdateSpinningWheelConfigRes{
		Message: constant.SUCCESS,
		Data: dto.SpinningWheelConfigData{
			ID:        resp.ID,
			Name:      resp.Name,
			Amount:    resp.Amount,
			Type:      dto.SpinningWheelTypes(resp.Type),
			Frequency: int(resp.Frequency),
			Icon:      resp.Icon,
			Status:    resp.Status,
		},
	}, nil
}
