package bet

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
)

func (b *bet) GetGameByID(ctx context.Context, ID uuid.UUID) (dto.Game, error) {
	return b.betStorage.GetGameByID(ctx, ID)
}

func (b *bet) UpdateGame(ctx context.Context, game dto.Game) (dto.Game, error) {
	if game.Status != "" {
		if game.Status != constant.ACTIVE && game.Status != constant.INACTIVE {
			err := fmt.Errorf("only ACTIVE OR INACTIVE status allowed")
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			return dto.Game{}, err
		}
	}

	_, err := b.betStorage.GetGameByID(ctx, game.ID)
	if err != nil {
		return dto.Game{}, err
	}

	return b.betStorage.UpdageGame(ctx, game)
}

func (b *bet) GetGames(ctx context.Context, req dto.GetRequest) (dto.GetGamesResp, error) {
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset
	return b.betStorage.GetGames(ctx, req)
}

func (b *bet) GetGameSummary(ctx context.Context) (dto.GetGameSummaryResp, error) {
	return b.betStorage.GetGameSummary(ctx)
}

func (b *bet) GetTransactionSummary(ctx context.Context) (dto.GetTransactionSummaryResp, error) {
	return b.betStorage.GetTransactionSummary(ctx)
}

func (b *bet) DisableAllGames(ctx context.Context) (dto.BlockGamesResp, error) {
	var games []dto.Game
	for _, gameIDs := range constant.GAMES {
		g, err := b.betStorage.GetGameByID(ctx, gameIDs)
		if err != nil {
			return dto.BlockGamesResp{}, err
		}

		gm, err := b.betStorage.UpdageGame(ctx, dto.Game{
			ID:     g.ID,
			Status: constant.INACTIVE,
			Name:   g.Name,
		})

		if err != nil {
			return dto.BlockGamesResp{}, err
		}
		games = append(games, gm)
	}
	return dto.BlockGamesResp{
		Message: constant.SUCCESS,
		Data:    games,
	}, nil
}

func (b *bet) ListInactiveGames(ctx context.Context) ([]dto.Game, error) {
	empty := []dto.Game{}
	resp, err := b.betStorage.ListInactiveGames(ctx)
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return empty, err
	}
	if len(resp) == 0 {
		return empty, nil
	}
	games := make([]dto.Game, 0, len(resp))
	for _, game := range resp {
		games = append(games, dto.Game{
			ID:     game.ID,
			Status: game.Status,
			Name:   game.Name,
			Photo:  game.Photo,
		})
	}
	return games, nil
}

func (b *bet) DeleteGame(ctx context.Context, req dto.Game) (dto.DeleteResponse, error) {

	if err := b.betStorage.DeleteGame(ctx, req.ID); err != nil {
		return dto.DeleteResponse{}, err
	}

	return dto.DeleteResponse{
		Message: constant.SUCCESS,
	}, nil
}

func (b *bet) AddGame(ctx context.Context, game dto.Game) (dto.Game, error) {
	return b.betStorage.AddGame(ctx, game.ID)
}

func (b *bet) UpdateGameStatus(ctx context.Context, game dto.Game) (dto.Game, error) {
	return b.betStorage.UpdateEnableStatus(ctx, game)
}
