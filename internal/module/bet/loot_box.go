package bet

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/platform/utils"
	"go.uber.org/zap"
)

func (b *bet) CreateLootBox(ctx context.Context, req dto.CreateLootBoxReq) (dto.CreateLootBoxRes, error) {
	if !dto.CheckValidTypes(req.PrizeType) {
		b.log.Error("invalid prize type", zap.Any("req", req))
		err := errors.ErrInvalidUserInput.Wrap(fmt.Errorf("invalid prize type "), "invalid prize type")
		return dto.CreateLootBoxRes{}, err
	}

	resp, err := b.betStorage.CreateLootbox(ctx, req)
	if err != nil {
		return dto.CreateLootBoxRes{}, err
	}

	return resp, nil
}

func (b *bet) GetLootBoxPrice(ctx context.Context) (decimal.Decimal, error) {
	price, exist, err := b.ConfigStorage.GetConfigByName(ctx, constant.LOOT_BOX)
	if err != nil {
		return decimal.Decimal{}, err
	}

	if !exist {
		// save default price to database
		_, err := b.ConfigStorage.CreateConfig(ctx, dto.Config{
			Name:  constant.LOOT_BOX,
			Value: fmt.Sprintf("%d", constant.LOOT_BOX_DEFAULT_PRICE),
		})

		if err != nil {
			return decimal.Decimal{}, err
		}

		return decimal.RequireFromString(fmt.Sprintf("%d", constant.LOOT_BOX_DEFAULT_PRICE)), nil
	}

	if price.Value == "" {
		err := fmt.Errorf("loot box price not set")
		err = errors.ErrInvalidUserInput.Wrap(err, "loot box price not set")
		b.log.Error("loot box price not set", zap.Error(err))
		return decimal.Decimal{}, err
	}

	resp, err := decimal.NewFromString(price.Value)
	if err != nil {
		b.log.Error("failed to parse loot box price", zap.Error(err), zap.String("price", price.Value))
		err = errors.ErrInvalidUserInput.Wrap(err, "failed to parse loot box price")
		return decimal.Decimal{}, err
	}

	return resp, nil
}

func (b *bet) UpdateLootBox(ctx context.Context, req dto.UpdateLootBoxReq) (dto.UpdateLootBoxRes, error) {
	if !dto.CheckValidTypes(req.PrizeType) {
		b.log.Error("invalid prize type", zap.Any("req", req))
		err := errors.ErrInvalidUserInput.Wrap(fmt.Errorf("invalid prize type "), "invalid prize type")
		return dto.UpdateLootBoxRes{}, err
	}

	resp, err := b.betStorage.UpdateLootbox(ctx, req)
	if err != nil {
		return dto.UpdateLootBoxRes{}, err
	}

	return resp, nil
}

func (b *bet) DeleteLootBox(ctx context.Context, id uuid.UUID) (dto.DeleteLootBoxRes, error) {
	resp, err := b.betStorage.DeleteLootbox(ctx, id)
	if err != nil {
		b.log.Error(err.Error(), zap.String("id", id.String()))
		err = errors.ErrUnableToUpdate.Wrap(err, "failed to delete lootbox")
		return dto.DeleteLootBoxRes{}, err
	}

	return resp, nil
}

func (b *bet) GetLootBox(ctx context.Context) ([]dto.LootBox, error) {
	resp, err := b.betStorage.GetAllLootboxes(ctx)
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableToUpdate.Wrap(err, "failed to get lootbox")
		return []dto.LootBox{}, err
	}

	return resp, nil
}

func (b *bet) PlaceLootBoxBet(ctx context.Context, userID uuid.UUID) ([]dto.PlaceLootBoxResp, error) {
	var logBox []dto.LootBox
	defer b.TriggerLevelResponse(ctx, userID)
	defer b.TriggerPlayerProgressBar(ctx, userID)
	defer b.InitiateTriggerSquadsProgressBar(ctx, userID)
	defer b.userWS.TriggerBalanceWS(ctx, userID)
	var lootboxResp []dto.PlaceLootBoxResp
	var templootboxResp []dto.PlaceLootBoxResp
	userLock := b.getUserLock(userID)
	userLock.Lock()
	defer userLock.Unlock()

	open, err := b.CheckBetLockStatus(ctx, constant.GAME_LOOT_BOX)
	if err != nil {
		return []dto.PlaceLootBoxResp{}, err
	}

	if !open {
		err := fmt.Errorf("game currently unavailable")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return []dto.PlaceLootBoxResp{}, err
	}

	resp, err := b.GetLootBoxPrice(ctx)
	if err != nil {
		b.log.Error("failed to get loot box price", zap.Error(err))
		return []dto.PlaceLootBoxResp{}, errors.ErrUnableToGet.Wrap(err, "failed to get loot box price")
	}

	//check for the users is blocked or not
	if err := b.CheckGameBlocks(ctx, userID); err != nil {
		return []dto.PlaceLootBoxResp{}, err
	}

	lootBoxs, err := b.GetLootBox(ctx)
	if err != nil {
		b.log.Error("failed to get loot boxes", zap.Error(err))
		return []dto.PlaceLootBoxResp{}, errors.ErrUnableToGet.Wrap(err, "failed to get loot boxes")
	}

	loots, err := b.SelectLootBoxes(lootBoxs)
	if err != nil {
		b.log.Error("failed to select loot boxes", zap.Error(err))
		return []dto.PlaceLootBoxResp{}, errors.ErrUnableToGet.Wrap(err, "failed to select loot boxes")
	}

	balance, exist, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		UserId:       userID,
		CurrencyCode: constant.POINT_CURRENCY,
	})

	if err != nil {
		b.log.Error("failed to get user balance", zap.Error(err), zap.Any("req", userID.String()))
		return []dto.PlaceLootBoxResp{}, errors.ErrUnableToGet.Wrap(err, "failed to get user balance")
	}
	if !exist || balance.RealMoney.LessThanOrEqual(decimal.Zero) {
		err := fmt.Errorf("insufficient balance")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		b.log.Error("insufficient balance", zap.Error(err), zap.Any("req", userID.String()))
		return []dto.PlaceLootBoxResp{}, err
	}

	if balance.RealMoney.LessThan(resp) {
		err := fmt.Errorf("insufficient balance to play loot box")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		b.log.Error("insufficient balance to play loot box", zap.Error(err), zap.Any("req", userID.String()))
		return []dto.PlaceLootBoxResp{}, err
	}

	// deduct the balance
	newBalance := balance.RealMoney.Sub(resp)
	transactionID := utils.GenerateTransactionId()
	_, err = b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
		UserID:    userID,
		Currency:  constant.POINT_CURRENCY,
		Component: constant.REAL_MONEY,
		Amount:    newBalance,
	})
	if err != nil {
		return []dto.PlaceLootBoxResp{}, err
	}

	// save transaction logs
	operationalGroupAndTypeIDs, err := b.CreateOrGetOperationalGroupAndType(ctx, constant.TRANSFER, constant.BET_OPERATIONAL_TYPE)
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    userID,
			Currency:  constant.POINT_CURRENCY,
			Component: constant.REAL_MONEY,
			Amount:    balance.RealMoney,
		})
		return []dto.PlaceLootBoxResp{}, err
	}

	// save operations logs
	_, err = b.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
		UserID:             userID,
		Component:          constant.REAL_MONEY,
		Currency:           constant.POINT_CURRENCY,
		Description:        fmt.Sprintf("place loot box bet amount %v, new balance is %v and  currency %s", resp, balance.RealMoney.Sub(resp), constant.POINT_CURRENCY),
		ChangeAmount:       resp,
		OperationalGroupID: operationalGroupAndTypeIDs.OperationalGroupID,
		OperationalTypeID:  operationalGroupAndTypeIDs.OperationalTypeID,
		BalanceAfterUpdate: &newBalance,
		TransactionID:      &transactionID,
	})
	if err != nil {
		return []dto.PlaceLootBoxResp{}, err
	}

	for _, lb := range loots {
		lb.AssociatedID = uuid.New()
		templootboxResp = append(lootboxResp, dto.PlaceLootBoxResp{
			LootBoxID: lb.AssociatedID,
		})

		logBox = append(logBox, dto.LootBox{
			ID:           lb.ID,
			AssociatedID: lb.AssociatedID,
			PrizeType:    lb.PrizeType,
			PrizeValue:   lb.PrizeValue,
			Probability:  lb.Probability,
			CreatedAt:    lb.CreatedAt,
			UpdatedAt:    lb.UpdatedAt,
		})

	}

	data, err := json.Marshal(logBox)
	if err != nil {
		b.log.Error("failed to marshal loot boxes", zap.Error(err), zap.Any("loot_boxes", logBox))
		err = errors.ErrInternalServerError.Wrap(err, "failed to marshal loot boxes")
		return []dto.PlaceLootBoxResp{}, err
	}

	// save to the database
	lbb, err := b.betStorage.PlaceLootBoxBet(ctx, dto.PlaceLootBoxBetReq{
		UserID:  userID,
		LootBox: data,
	})

	for _, lb := range templootboxResp {
		lb.ID = lbb.ID
		lootboxResp = append(lootboxResp, lb)
	}

	if err != nil {
		b.log.Error("failed to place loot box bet", zap.Error(err), zap.Any("req", userID.String()))
		b.betStorage.PlaceLootBoxBet(ctx, dto.PlaceLootBoxBetReq{})
		return lootboxResp, nil
	}

	return lootboxResp, nil
}

func (b *bet) UpdateLootBoxBet(ctx context.Context, req dto.PlaceLootBoxBetReq) (dto.PlaceLootBoxBetRes, error) {
	resp, err := b.betStorage.UpdateLootBoxBet(ctx, req)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableToUpdate.Wrap(err, "failed to update lootbox bet")
		return dto.PlaceLootBoxBetRes{}, err
	}

	return resp, nil
}

func (b *bet) SelectLootBoxes(lootBoxes []dto.LootBox) ([]dto.LootBox, error) {
	if len(lootBoxes) == 0 {
		err := fmt.Errorf("no loot boxes provided")
		return nil, errors.ErrAccountingError.Wrap(err, "no loot boxes provided")
	}
	if len(lootBoxes) < 2 {
		err := fmt.Errorf("at least two loot boxes are required to allow one repeat")
		return nil, errors.ErrInternalServerError.Wrap(err, "at least two loot boxes are required to allow one repeat")
	}

	for _, lb := range lootBoxes {
		if lb.Probability.LessThanOrEqual(decimal.Zero) {
			return nil, fmt.Errorf("invalid probability for loot box ID %s: must be positive", lb.ID)
		}
	}

	totalProb := decimal.Zero
	for _, lb := range lootBoxes {
		totalProb = totalProb.Add(lb.Probability)
	}

	rand.Seed(time.Now().UnixNano())

	selectionCount := make(map[uuid.UUID]int)
	selected := make([]dto.LootBox, 0, 3)

	for i := 0; i < 3; i++ {
		attempts := 0
		maxAttempts := 100
		for {
			attempts++
			if attempts > maxAttempts {
				return nil, errors.ErrInternalServerError.Wrap(fmt.Errorf("failed to select loot box "), "failed to select loot boxes within attempt limit")
			}

			randVal := decimal.NewFromFloat(rand.Float64()).Mul(totalProb)

			cumulative := decimal.Zero
			for _, lb := range lootBoxes {

				if selectionCount[lb.ID] >= 2 {
					continue
				}
				cumulative = cumulative.Add(lb.Probability)
				if randVal.LessThanOrEqual(cumulative) {

					if i == 2 && selectionCount[lb.ID] == 1 {

						uniqueCount := len(selectionCount)
						if _, exists := selectionCount[lb.ID]; !exists {
							uniqueCount++
						}
						if uniqueCount < 2 {
							break
						}
					}

					selected = append(selected, lb)
					selectionCount[lb.ID]++
					break
				}
			}

			if len(selected) == i+1 {
				break
			}
		}
	}

	if len(selectionCount) < 2 {
		err := fmt.Errorf("not enough unique loot boxes selected: %d unique boxes, expected at least 2", len(selectionCount))
		err = errors.ErrInternalServerError.Wrap(err, "failed to select loot boxes")
		b.log.Error("failed to select loot boxes", zap.Error(err), zap.Int("unique_boxes", len(selectionCount)))
		return nil, err
	}
	return selected, nil
}

func (b *bet) SelectLootBox(ctx context.Context, lootBox dto.PlaceLootBoxResp, userID uuid.UUID) (dto.LootBoxBetResp, error) {
	defer b.TriggerLevelResponse(ctx, userID)
	defer b.TriggerPlayerProgressBar(ctx, userID)
	defer b.InitiateTriggerSquadsProgressBar(ctx, userID)
	defer b.userWS.TriggerBalanceWS(ctx, userID)
	var wonStatus string
	var lootbox []dto.LootBox
	var response dto.LootBox
	resp, err := b.betStorage.GetLootBoxBetByID(ctx, lootBox.ID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("req", lootBox))
		err = errors.ErrUnableToGet.Wrap(err, "failed to get lootbox by id")
		return dto.LootBoxBetResp{}, err
	}

	if err := json.Unmarshal(resp.LootBox, &lootbox); err != nil {
		b.log.Error("failed to unmarshal loot box", zap.Error(err), zap.Any("loot_box", resp.LootBox))
		err = errors.ErrInternalServerError.Wrap(err, "failed to unmarshal loot box")
		return dto.LootBoxBetResp{}, err
	}
	if len(lootbox) == 0 {
		err := fmt.Errorf("no loot boxes found for the given ID")
		err = errors.ErrInternalServerError.Wrap(err, "no loot boxes found for the given ID")
		b.log.Error("no loot boxes found for the given ID", zap.Error(err), zap.Any("loot_box_id", lootBox.LootBoxID))
		return dto.LootBoxBetResp{}, err
	}

	for _, lb := range lootbox {
		if lb.AssociatedID == lootBox.LootBoxID {
			response = lb
			break
		}
	}

	if response.ID == uuid.Nil {
		err := fmt.Errorf("loot box with ID %s not found", lootBox.LootBoxID)
		err = errors.ErrInternalServerError.Wrap(err, "loot box with ID not found")
		b.log.Error("loot box with ID not found", zap.Error(err), zap.Any("loot_box_id", lootBox.LootBoxID))
		return dto.LootBoxBetResp{}, err
	}

	if response.PrizeValue.GreaterThan(decimal.Zero) {
		wonStatus = constant.WON
		// update user balance
		newBalance, exist, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
			UserId:       userID,
			CurrencyCode: constant.POINT_CURRENCY,
		})
		if err != nil {
			b.log.Error("failed to get user balance", zap.Error(err), zap.Any("req", userID.String()))
			return dto.LootBoxBetResp{}, errors.ErrUnableToGet.Wrap(err, "failed to get user balance")
		}
		if !exist {
			err := fmt.Errorf("user balance not found")
			err = errors.ErrInternalServerError.Wrap(err, "user balance not found")
			b.log.Error("user balance not found", zap.Error(err), zap.Any("req", userID.String()))
			return dto.LootBoxBetResp{}, err
		}
		newBalance.RealMoney = newBalance.RealMoney.Add(response.PrizeValue)
		transactionID := utils.GenerateTransactionId()
		// update user balance
		_, err = b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    userID,
			Currency:  constant.POINT_CURRENCY,
			Component: constant.REAL_MONEY,
			Amount:    newBalance.RealMoney,
		})

		if err != nil {
			b.log.Error("failed to update user balance", zap.Error(err), zap.Any("req", userID.String()))
			return dto.LootBoxBetResp{}, errors.ErrUnableToUpdate.Wrap(err, "failed to update user balance")
		}
		// save transaction logs
		operationalGroupAndTypeIDsResp, err := b.CreateOrGetOperationalGroupAndType(ctx, constant.TRANSFER, constant.BET_CASHOUT)
		if err != nil {
			b.log.Error("failed to create or get operational group and type", zap.Error(err))
			return dto.LootBoxBetResp{}, errors.ErrInternalServerError.Wrap(err, "failed to create or get operational group and type")
		}
		// save operations logs
		_, err = b.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
			UserID:             userID,
			Component:          constant.REAL_MONEY,
			Currency:           constant.POINT_CURRENCY,
			Description:        fmt.Sprintf("won loot box bet amount %v, new balance is %v and currency %s", response.PrizeValue, newBalance.RealMoney, constant.POINT_CURRENCY),
			ChangeAmount:       response.PrizeValue,
			OperationalGroupID: operationalGroupAndTypeIDsResp.OperationalGroupID,
			OperationalTypeID:  operationalGroupAndTypeIDsResp.OperationalTypeID,
			BalanceAfterUpdate: &newBalance.RealMoney,
			TransactionID:      &transactionID,
		})
		if err != nil {
			b.log.Error("failed to save balance logs", zap.Error(err), zap.Any("req", userID.String()))
			return dto.LootBoxBetResp{}, errors.ErrUnableToUpdate.Wrap(err, "failed to save balance logs")
		}
		// update loot box bet status
		_, err = b.UpdateLootBoxBet(ctx, dto.PlaceLootBoxBetReq{
			ID:            resp.ID,
			Status:        constant.CLOSED,
			UserID:        userID,
			UserSelection: lootBox.LootBoxID,
			LootBox:       resp.LootBox,
			WonStatus:     constant.WON,
		})

	} else {
		wonStatus = constant.LOSE
		// update loot box bet status
		_, err = b.UpdateLootBoxBet(ctx, dto.PlaceLootBoxBetReq{
			ID:            resp.ID,
			Status:        constant.CLOSED,
			UserID:        userID,
			UserSelection: lootBox.LootBoxID,
			LootBox:       resp.LootBox,
			WonStatus:     constant.LOSE,
		})

		if err != nil {
			b.log.Error("failed to update loot box bet", zap.Error(err), zap.Any("req", lootBox))
			err = errors.ErrUnableToUpdate.Wrap(err, "failed to update loot box bet")
			return dto.LootBoxBetResp{}, err
		}

	}

	return dto.LootBoxBetResp{
		LootBoxID:  response.ID,
		PrizeType:  response.PrizeType,
		PrizeValue: response.PrizeValue,
		WonStatus:  wonStatus,
		Status:     constant.CLOSED,
	}, nil

}
