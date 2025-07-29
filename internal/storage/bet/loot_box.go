package bet

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/constant/model/db"
	"go.uber.org/zap"
)

func (b *bet) CreateLootbox(ctx context.Context, req dto.CreateLootBoxReq) (dto.CreateLootBoxRes, error) {

	resp, err := b.db.CreateLootBox(ctx, db.CreateLootBoxParams{
		Type:        req.PrizeType,
		Prizeamount: req.PrizeValue,
		Weight:      req.Probability,
	})

	if err != nil {
		b.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableTocreate.Wrap(err, "failed to create lootbox")
		return dto.CreateLootBoxRes{}, err
	}

	return dto.CreateLootBoxRes{
		Message: "lootbox created successfully",
		LootBox: dto.LootBox{
			ID:          resp.ID,
			PrizeType:   resp.Type,
			PrizeValue:  resp.Prizeamount,
			Probability: resp.Weight,
			CreatedAt:   resp.CreatedAt,
			UpdatedAt:   resp.UpdatedAt,
		},
	}, nil
}

func (b *bet) UpdateLootbox(ctx context.Context, req dto.UpdateLootBoxReq) (dto.UpdateLootBoxRes, error) {
	resp, err := b.db.UpdateLootBox(ctx, db.UpdateLootBoxParams{
		ID:          req.ID,
		Type:        req.PrizeType,
		Prizeamount: req.PrizeValue,
		Weight:      req.Probability,
	})

	if err != nil {
		b.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableToUpdate.Wrap(err, "failed to update lootbox")
		return dto.UpdateLootBoxRes{}, err
	}

	return dto.UpdateLootBoxRes{
		Message: "lootbox updated successfully",
		LootBox: dto.LootBox{
			ID:          resp.ID,
			PrizeType:   resp.Type,
			PrizeValue:  resp.Prizeamount,
			Probability: resp.Weight,
			CreatedAt:   resp.CreatedAt,
			UpdatedAt:   resp.UpdatedAt,
		},
	}, nil
}

func (b *bet) DeleteLootbox(ctx context.Context, id uuid.UUID) (dto.DeleteLootBoxRes, error) {
	err := b.db.DeleteLootBoxByID(ctx, id)
	if err != nil {
		b.log.Error(err.Error(), zap.String("id", id.String()))
		err = errors.ErrUnableToUpdate.Wrap(err, "failed to delete lootbox")
		return dto.DeleteLootBoxRes{}, err
	}

	return dto.DeleteLootBoxRes{
		Message: "lootbox deleted successfully",
	}, nil
}

func (b *bet) GetLootboxByID(ctx context.Context, id uuid.UUID) (dto.LootBox, error) {
	resp, err := b.db.GetLootBoxByID(ctx, id)
	if err != nil {
		b.log.Error(err.Error(), zap.String("id", id.String()))
		err = errors.ErrUnableToGet.Wrap(err, "failed to get lootbox by id")
		return dto.LootBox{}, err
	}

	return dto.LootBox{
		ID:          resp.ID,
		PrizeType:   resp.Type,
		PrizeValue:  resp.Prizeamount,
		Probability: resp.Weight,
	}, nil
}

func (b *bet) GetAllLootboxes(ctx context.Context) ([]dto.LootBox, error) {
	resp, err := b.db.GetAllLootBoxes(ctx)
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, "failed to get all lootboxes")
		return nil, err
	}

	var lootboxes []dto.LootBox
	for _, box := range resp {
		lootboxes = append(lootboxes, dto.LootBox{
			ID:          box.ID,
			PrizeType:   box.Type,
			PrizeValue:  box.Prizeamount,
			Probability: box.Weight,
		})
	}

	return lootboxes, nil
}

func (b *bet) PlaceLootBoxBet(ctx context.Context, req dto.PlaceLootBoxBetReq) (dto.PlaceLootBoxBetRes, error) {
	resp, err := b.db.PlaceLootBoxBet(ctx, db.PlaceLootBoxBetParams{
		UserID:  req.UserID,
		LootBox: pgtype.JSONB{Bytes: req.LootBox, Status: pgtype.Present},
	})

	if err != nil {
		b.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableTocreate.Wrap(err, "failed to place lootbox bet")
		return dto.PlaceLootBoxBetRes{}, err
	}

	return dto.PlaceLootBoxBetRes{
		ID:      resp.ID,
		Message: "bet placed successfully",
		LootBox: resp.LootBox.Bytes,
	}, nil
}

func (b *bet) UpdateLootBoxBet(ctx context.Context, req dto.PlaceLootBoxBetReq) (dto.PlaceLootBoxBetRes, error) {
	resp, err := b.db.UpdateLootBoxBet(ctx, db.UpdateLootBoxBetParams{
		ID:            req.ID,
		Status:        req.Status,
		UserSelection: uuid.NullUUID{UUID: req.UserSelection, Valid: req.UserSelection != uuid.Nil},
		Wonstatus:     req.WonStatus,
	})

	if err != nil {
		b.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableToUpdate.Wrap(err, "failed to update lootbox bet")
		return dto.PlaceLootBoxBetRes{}, err
	}

	return dto.PlaceLootBoxBetRes{
		Message: "bet updated successfully",
		LootBox: resp.LootBox.Bytes,
	}, nil
}

func (b *bet) GetLootBoxBetByID(ctx context.Context, id uuid.UUID) (dto.PlaceLootBoxBetRes, error) {
	resp, err := b.db.GetLootBoxBetByID(ctx, id)
	if err != nil {
		b.log.Error(err.Error(), zap.String("id", id.String()))
		err = errors.ErrUnableToGet.Wrap(err, "failed to get lootbox bet by id")
		return dto.PlaceLootBoxBetRes{}, err
	}

	return dto.PlaceLootBoxBetRes{
		ID:      resp.ID,
		LootBox: resp.LootBox.Bytes,
	}, nil
}
