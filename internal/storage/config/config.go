package config

import (
	"context"

	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/db"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/storage"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type config struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

func Init(db *persistencedb.PersistenceDB, log *zap.Logger) storage.Config {
	return &config{
		db:  db,
		log: log,
	}
}

func (c *config) GetConfigByName(ctx context.Context, name string) (dto.Config, bool, error) {
	resp, err := c.db.Queries.GetConfigByName(ctx, name)
	if err != nil && err.Error() != dto.ErrNoRows {
		c.log.Error("Error getting config by name", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.Config{}, false, err

	}
	if err != nil {
		return dto.Config{}, false, nil
	}
	return dto.Config{
		ID:    resp.ID,
		Name:  resp.Name,
		Value: resp.Value,
	}, true, nil
}

func (c *config) CreateConfig(ctx context.Context, cn dto.Config) (dto.Config, error) {
	resp, err := c.db.Queries.CreateConfig(ctx, db.CreateConfigParams{
		Name:  cn.Name,
		Value: cn.Value,
	})
	if err != nil {
		c.log.Error("Error creating config", zap.Error(err))
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.Config{}, err
	}
	return dto.Config{
		ID:    resp.ID,
		Name:  resp.Name,
		Value: resp.Value,
	}, nil
}

func (c *config) UpdateConfigByID(ctx context.Context, cn dto.Config) (dto.Config, error) {
	resp, err := c.db.Queries.UpdateConfigs(ctx, db.UpdateConfigsParams{
		ID:    cn.ID,
		Value: cn.Name,
	})
	if err != nil {
		c.log.Error("Error updating config", zap.Error(err))
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.Config{}, err
	}
	return dto.Config{
		ID:    resp.ID,
		Name:  resp.Name,
		Value: resp.Value,
	}, nil
}

func (c *config) UpdateConfigByName(ctx context.Context, cn dto.Config) (dto.Config, error) {
	resp, err := c.db.Queries.UpdateConfigByName(ctx, db.UpdateConfigByNameParams{
		Name:  cn.Name,
		Value: cn.Value,
	})
	if err != nil {
		c.log.Error("Error updating config", zap.Error(err))
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.Config{}, err
	}
	return dto.Config{
		ID:    resp.ID,
		Name:  resp.Name,
		Value: resp.Value,
	}, nil
}

func (c *config) UpdateSpinningWheelConfig(ctx context.Context, cn dto.Config) (dto.Config, error) {
	resp, err := c.db.Queries.UpdateConfigByName(ctx, db.UpdateConfigByNameParams{
		Name:  cn.Name,
		Value: cn.Value,
	})

	if err != nil {
		c.log.Error("Error updating config", zap.Error(err))
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.Config{}, err
	}

	return dto.Config{
		ID:    resp.ID,
		Name:  resp.Name,
		Value: resp.Value,
	}, nil
}

func (c *config) GetScratchCardConfigs(ctx context.Context) (dto.GetScratchCardConfigs, error) {
	var scratchCards []dto.ScratchCardConfig
	resp, err := c.db.Queries.GetScratchCardConfigs(ctx)
	if err != nil && err.Error() != dto.ErrNoRows {
		c.log.Error("Error getting scratch card configs", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetScratchCardConfigs{}, err
	}

	for _, v := range resp {
		prize, err := decimal.NewFromString(v.Value)
		if err != nil {
			prize = decimal.Zero
		}
		scratchCards = append(scratchCards, dto.ScratchCardConfig{
			Name:  v.Name,
			Id:    v.ID,
			Prize: prize,
		})
	}

	return dto.GetScratchCardConfigs{
		Message: constant.SUCCESS,
		Data:    scratchCards,
	}, nil
}

func (c *config) UpdateScratchGameConfig(ctx context.Context, cn dto.Config) (dto.Config, error) {

	resp, err := c.db.Queries.UpdateScratchCardConfigById(ctx, db.UpdateScratchCardConfigByIdParams{
		ID:    cn.ID,
		Name:  cn.Name,
		Value: cn.Value,
	})
	if err != nil {
		c.log.Error("Error updating scratch card config", zap.Error(err))
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.Config{}, err
	}
	return dto.Config{
		ID:    resp.ID,
		Name:  resp.Name,
		Value: resp.Value,
	}, nil

}
