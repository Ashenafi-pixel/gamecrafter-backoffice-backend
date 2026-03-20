package player

import (
	"context"
	"fmt"

	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/module"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"

	"github.com/google/uuid"
)

type player struct {
	log           *zap.Logger
	playerStorage storage.Player
}

func Init(playerStorage storage.Player, log *zap.Logger) module.Player {
	return &player{
		log:           log,
		playerStorage: playerStorage,
	}
}

func (p *player) CreatePlayer(ctx context.Context, req dto.CreatePlayerReq) (dto.CreatePlayerRes, error) {
	if err := dto.ValidateCreatePlayer(req); err != nil {
		userFriendlyMsg := "Please check your input data. Some required fields are missing or invalid."
		err = errors.ErrInvalidUserInput.Wrap(err, userFriendlyMsg)
		return dto.CreatePlayerRes{}, err
	}

	player := dto.Player{
		Email:                 req.Email,
		Username:              req.Username,
		Password:              req.Password,
		Phone:                 req.Phone,
		FirstName:             req.FirstName,
		LastName:              req.LastName,
		DefaultCurrency:       req.DefaultCurrency,
		Brand:                 req.Brand,
		DateOfBirth:           req.DateOfBirth.Time,
		Country:               req.Country,
		State:                 req.State,
		StreetAddress:         req.StreetAddress,
		PostalCode:            req.PostalCode,
		TestAccount:           req.TestAccount,
		EnableWithdrawalLimit: req.EnableWithdrawalLimit,
		BrandID:               req.BrandID,
	}

	createdPlayer, err := p.playerStorage.CreatePlayer(ctx, player)
	if err != nil {
		return dto.CreatePlayerRes{}, err
	}

	return dto.CreatePlayerRes{
		ID:        createdPlayer.ID,
		Email:     createdPlayer.Email,
		Username:  createdPlayer.Username,
		CreatedAt: createdPlayer.CreatedAt,
	}, nil
}

func (p *player) GetPlayerByID(ctx context.Context, playerID uuid.UUID) (dto.Player, error) {
	if playerID == uuid.Nil {
		return dto.Player{}, errors.ErrInvalidUserInput.New("Please provide a valid player ID.")
	}

	player, exists, err := p.playerStorage.GetPlayerByID(ctx, playerID)
	if err != nil {
		return dto.Player{}, err
	}

	if !exists {
		p.log.Warn("player not found", zap.String("id", playerID.String()))
		userFriendlyMsg := fmt.Sprintf("The requested player with ID %s could not be found. Please check the player ID and try again.", playerID.String())
		err := errors.ErrResourceNotFound.New(userFriendlyMsg)
		return dto.Player{}, err
	}

	return player, nil
}

func (p *player) GetPlayers(ctx context.Context, req dto.GetPlayersReqs) (dto.GetPlayersRess, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 {
		req.PerPage = 10
	}

	playersRes, err := p.playerStorage.GetPlayers(ctx, req)
	if err != nil {
		return dto.GetPlayersRess{}, err
	}
	return playersRes, nil
}

func (p *player) UpdatePlayer(ctx context.Context, req dto.UpdatePlayerReq) (dto.UpdatePlayerRes, error) {
	player, exists, err := p.playerStorage.GetPlayerByID(ctx, req.ID)
	if err != nil {
		return dto.UpdatePlayerRes{}, err
	}
	if !exists {
		p.log.Error("player not found", zap.String("playerID", req.ID.String()))
		userFriendlyMsg := fmt.Sprintf("The player with ID %s could not be found. Please check the player ID and try again.", req.ID.String())
		err := errors.ErrResourceNotFound.New(userFriendlyMsg)
		return dto.UpdatePlayerRes{}, err
	}

	if err := dto.ValidateUpdatePlayer(req); err != nil {
		userFriendlyMsg := "Please check your input data. Some fields may be invalid or missing."
		err = errors.ErrInvalidUserInput.Wrap(err, userFriendlyMsg)
		return dto.UpdatePlayerRes{}, err
	}

	// Fill in missing fields from existing player
	updatePlayer := dto.Player{
		ID:                    req.ID,
		Email:                 player.Email,
		Username:              player.Username,
		Phone:                 player.Phone,
		FirstName:             player.FirstName,
		LastName:              player.LastName,
		DefaultCurrency:       player.DefaultCurrency,
		Brand:                 player.Brand,
		DateOfBirth:           player.DateOfBirth,
		Country:               player.Country,
		State:                 player.State,
		StreetAddress:         player.StreetAddress,
		PostalCode:            player.PostalCode,
		TestAccount:           player.TestAccount,
		EnableWithdrawalLimit: player.EnableWithdrawalLimit,
		BrandID:               player.BrandID,
	}

	if req.Email != nil {
		updatePlayer.Email = *req.Email
	}
	if req.Username != nil {
		updatePlayer.Username = *req.Username
	}
	if req.Phone != nil {
		updatePlayer.Phone = req.Phone
	}
	if req.FirstName != nil {
		updatePlayer.FirstName = req.FirstName
	}
	if req.LastName != nil {
		updatePlayer.LastName = req.LastName
	}
	if req.DefaultCurrency != nil {
		updatePlayer.DefaultCurrency = *req.DefaultCurrency
	}
	if req.Brand != nil {
		updatePlayer.Brand = req.Brand
	}
	if req.DateOfBirth != nil {
		updatePlayer.DateOfBirth = *req.DateOfBirth
	}
	if req.Country != nil {
		updatePlayer.Country = *req.Country
	}
	if req.State != nil {
		updatePlayer.State = req.State
	}
	if req.StreetAddress != nil {
		updatePlayer.StreetAddress = req.StreetAddress
	}
	if req.PostalCode != nil {
		updatePlayer.PostalCode = req.PostalCode
	}
	if req.TestAccount != nil {
		updatePlayer.TestAccount = *req.TestAccount
	}
	if req.EnableWithdrawalLimit != nil {
		updatePlayer.EnableWithdrawalLimit = *req.EnableWithdrawalLimit
	}
	if req.BrandID != nil {
		updatePlayer.BrandID = req.BrandID
	}

	updatedPlayer, err := p.playerStorage.UpdatePlayer(ctx, updatePlayer)
	if err != nil {
		return dto.UpdatePlayerRes{}, err
	}

	return dto.UpdatePlayerRes{
		Player: updatedPlayer,
	}, nil
}

func (p *player) DeletePlayer(ctx context.Context, playerID uuid.UUID) error {
	if playerID == uuid.Nil {
		return errors.ErrInvalidUserInput.New("Please provide a valid player ID.")
	}

	// Check if player exists before deleting
	_, exists, err := p.playerStorage.GetPlayerByID(ctx, playerID)
	if err != nil {
		return err
	}
	if !exists {
		p.log.Warn("player not found for deletion", zap.String("id", playerID.String()))
		userFriendlyMsg := fmt.Sprintf("The player with ID %s could not be found. Please check the player ID and try again.", playerID.String())
		err := errors.ErrResourceNotFound.New(userFriendlyMsg)
		return err
	}

	// Delete the player
	err = p.playerStorage.DeletePlayer(ctx, playerID)
	if err != nil {
		return err
	}

	return nil
}


