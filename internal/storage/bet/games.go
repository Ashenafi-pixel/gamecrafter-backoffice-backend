package bet

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/db"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

func (b *bet) CreateGame(ctx context.Context, req dto.Game) (dto.Game, error) {
	resp, err := b.db.Queries.CreateGame(ctx, db.CreateGameParams{
		ID:   req.ID,
		Name: req.Name,
	})
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.Game{}, err
	}

	return dto.Game{
		ID:     resp.ID,
		Status: resp.Status,
		Name:   resp.Name,
	}, nil
}

func (b *bet) GetGames(ctx context.Context, req dto.GetRequest) (dto.GetGamesResp, error) {
	//game list
	var games []dto.Game
	var gameResp dto.GetGamesResp
	resp, err := b.db.Queries.GetGames(ctx, db.GetGamesParams{
		Status: constant.ACTIVE,
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetGamesResp{}, err
	}

	if err != nil || len(resp) == 0 {
		return dto.GetGamesResp{}, nil
	}
	totalPages := 1
	for i, game := range resp {
		games = append(games, dto.Game{
			ID:      game.ID,
			Status:  game.Status,
			Name:    game.Name,
			Photo:   game.Photo.String,
			Enabled: game.Enabled.Bool,
		})

		if i == 0 {
			totalPages = int(int(game.TotalRows) / req.PerPage)
			if int(game.TotalRows)%req.PerPage != 0 {
				totalPages++
			}
		}
	}
	gameResp.Message = constant.SUCCESS
	gameResp.Data.TotalPages = totalPages
	gameResp.Data.Games = games

	return gameResp, nil
}

func (b *bet) GetAllGames(ctx context.Context) (dto.GetGamesResp, error) {
	var games []dto.Game
	resp, err := b.db.Queries.GetAllGames(ctx)
	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetGamesResp{}, err
	}
	if err != nil || len(resp) == 0 {
		return dto.GetGamesResp{}, nil
	}

	for _, game := range resp {
		games = append(games, dto.Game{
			ID:     game.ID,
			Status: game.Status,
			Name:   game.Name,
		})
	}
	return dto.GetGamesResp{
		Message: constant.SUCCESS,
		Data: dto.GetGamesData{
			TotalPages: 1,
			Games:      games,
		},
	}, nil
}

func (b *bet) GetGameByID(ctx context.Context, ID uuid.UUID) (dto.Game, error) {
	game, err := b.db.Queries.GetGameByID(ctx, ID)
	if err != nil {
		return dto.Game{}, err
	}
	return dto.Game{ID: game.ID, Status: game.Status, Name: game.Name}, nil
}

func (b *bet) UpdageGame(ctx context.Context, game dto.Game) (dto.Game, error) {
	resp, err := b.db.Queries.UpdateGame(ctx, db.UpdateGameParams{
		Name:   game.Name,
		Status: game.Status,
		ID:     game.ID,
	})
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.Game{}, err
	}
	return dto.Game{
		ID:     resp.ID,
		Status: resp.Status,
		Name:   resp.Name,
	}, nil
}

func (b *bet) DeleteGame(ctx context.Context, ID uuid.UUID) error {
	err := b.db.Queries.DeleteGame(ctx, ID)
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return err
	}

	return nil
}

func (b *bet) ListInactiveGames(ctx context.Context) ([]dto.Game, error) {
	resp, err := b.db.Queries.GetGames(ctx, db.GetGamesParams{
		Status: constant.INACTIVE,
		Limit:  1000,
		Offset: 0,
	})
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return nil, err
	}
	var games []dto.Game
	for _, game := range resp {
		games = append(games, dto.Game{
			ID:     game.ID,
			Status: game.Status,
			Name:   game.Name,
			Photo:  game.Photo.String,
		})
	}
	return games, nil
}

func (b *bet) AddGame(ctx context.Context, ID uuid.UUID) (dto.Game, error) {
	resp, err := b.db.Queries.AddGame(ctx, ID)
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.Game{}, err
	}
	return dto.Game{
		ID:      resp.ID,
		Status:  resp.Status,
		Name:    resp.Name,
		Photo:   resp.Photo.String,
		Enabled: resp.Enabled.Bool,
	}, nil
}

func (b *bet) UpdateEnableStatus(ctx context.Context, game dto.Game) (dto.Game, error) {
	resp, err := b.db.Queries.ChangeEnableStatus(ctx, db.ChangeEnableStatusParams{
		ID:      game.ID,
		Enabled: sql.NullBool{Bool: game.Enabled, Valid: true},
	})
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.Game{}, err
	}
	return dto.Game{
		ID:      resp.ID,
		Status:  resp.Status,
		Name:    resp.Name,
		Photo:   resp.Photo.String,
		Enabled: resp.Enabled.Bool,
	}, nil
}

func (s *bet) AddFakeBalanceLog(ctx context.Context, userID uuid.UUID, changeAmount decimal.Decimal, currency string) error {
	_, err := s.db.Queries.AddFakeBalanceLog(ctx, db.AddFakeBalanceLogParams{
		UserID:       uuid.NullUUID{UUID: userID, Valid: true},
		ChangeAmount: decimal.NullDecimal{Decimal: changeAmount, Valid: true},
		Currency:     sql.NullString{String: "P", Valid: true},
	})

	if err != nil {
		s.log.Error(err.Error(), zap.Any("user_id", userID), zap.Any("change_amount", changeAmount), zap.Any("currency", currency))
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return err
	}

	return nil
}
