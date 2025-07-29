package bet

import (
	"context"
	"fmt"
	"math/rand"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/joshjones612/egyptkingcrash/internal/constant"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/platform/utils"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

func (b *bet) GetSpinningWheelPrice(ctx context.Context) (dto.GetSpinningWheelPrice, error) {
	price, exist, err := b.ConfigStorage.GetConfigByName(ctx, constant.SPINNING_WHEEL)
	if err != nil {
		return dto.GetSpinningWheelPrice{}, err
	}

	if !exist {
		// save default price to database
		_, err := b.ConfigStorage.CreateConfig(ctx, dto.Config{
			Name:  constant.SPINNING_WHEEL,
			Value: fmt.Sprintf("%d", constant.SPINNING_WHEEL_DEFAULT_PRICE),
		})

		if err != nil {
			return dto.GetSpinningWheelPrice{}, err
		}

		return dto.GetSpinningWheelPrice{
			Message: constant.SUCCESS,
			Price:   decimal.NewFromInt(constant.SPINNING_WHEEL_DEFAULT_PRICE),
		}, nil
	}

	return dto.GetSpinningWheelPrice{
		Message: constant.SUCCESS,
		Price:   decimal.RequireFromString(price.Value),
	}, nil
}

func (b *bet) PlaceSpinningWheelBet(ctx context.Context, userID uuid.UUID) (dto.PlaceSpinningWheelResp, error) {
	defer b.TriggerLevelResponse(ctx, userID)
	defer b.TriggerPlayerProgressBar(ctx, userID)
	defer b.InitiateTriggerSquadsProgressBar(ctx, userID)
	defer b.userWS.TriggerBalanceWS(ctx, userID)

	freeSping := false
	var betAmount string
	var wonAmount string
	b.spinningWheelFreeSpinsLocker.Lock()
	defer b.spinningWheelFreeSpinsLocker.Unlock()
	var newBalance decimal.Decimal
	open, err := b.CheckBetLockStatus(ctx, constant.GAME_SPINNING_WHEEL)
	if err != nil {
		return dto.PlaceSpinningWheelResp{}, err
	}

	if !open {
		err := fmt.Errorf("game currently unavailable")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.PlaceSpinningWheelResp{}, err
	}

	// current  spinning wheel price
	price, err := b.GetSpinningWheelPrice(ctx)

	if err != nil {
		return dto.PlaceSpinningWheelResp{}, err
	}

	// get user balance
	// check balance
	balance, exist, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		UserId:   userID,
		Currency: constant.POINT_CURRENCY,
	})

	if err != nil {
		return dto.PlaceSpinningWheelResp{}, err
	}

	// get all spinning wheel configs
	// get spinning wheel free spins
	spinningWheelConfigs, err := b.betStorage.GetAllSpinningWheelConfigs(ctx)
	if err != nil {
		return dto.PlaceSpinningWheelResp{}, err
	}

	if len(spinningWheelConfigs) == 0 {
		err := fmt.Errorf("spinning wheel configs is empty")
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.PlaceSpinningWheelResp{}, err
	}

	// check if user has free spin or not
	spins, ok := b.spinningwheelFreeSpins[userID]
	if ok && spins > 0 {
		freeSping = true
		betAmount = "free spins"
	} else {
		betAmount = price.Price.String()
	}

	if !freeSping && (balance.RealMoney.LessThan(price.Price) || !exist) {
		err := fmt.Errorf("insufficient balance ")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.PlaceSpinningWheelResp{}, err
	}

	// do transaction if it is not free wheel
	b.spinningwheelBets = b.spinningwheelBets.Sub(price.Price)
	if !freeSping {
		// to transactions
		//update user balance
		newBalance = balance.RealMoney.Sub(price.Price)
		transactionID := utils.GenerateTransactionId()
		_, err = b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    userID,
			Currency:  constant.POINT_CURRENCY,
			Component: constant.REAL_MONEY,
			Amount:    newBalance,
		})
		if err != nil {
			return dto.PlaceSpinningWheelResp{}, err
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
			return dto.PlaceSpinningWheelResp{}, err
		}

		// save operations logs
		_, err = b.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
			UserID:             userID,
			Component:          constant.REAL_MONEY,
			Currency:           constant.POINT_CURRENCY,
			Description:        fmt.Sprintf("place spinning wheel bet amount %v, new balance is %v and  currency %s", price.Price, balance.RealMoney.Sub(price.Price), constant.POINT_CURRENCY),
			ChangeAmount:       price.Price,
			OperationalGroupID: operationalGroupAndTypeIDs.OperationalGroupID,
			OperationalTypeID:  operationalGroupAndTypeIDs.OperationalTypeID,
			BalanceAfterUpdate: &newBalance,
			TransactionID:      &transactionID,
		})
		if err != nil {
			return dto.PlaceSpinningWheelResp{}, err
		}

	} else {
		b.spinningwheelFreeSpins[userID] = b.spinningwheelFreeSpins[userID] - 1
	}

	// get spinning mystery prices

	resp, err := b.GetSpinningWheel(ctx)
	if err != nil {
		return dto.PlaceSpinningWheelResp{}, err
	}

	//get spinning wheel mystery
	switch resp.Type {
	case dto.Mystery:
		// get all mystry boxs
		spinningWheelMysteries, err := b.betStorage.GetAllSpinningWheelMysteries(ctx)
		if err != nil {
			return dto.PlaceSpinningWheelResp{}, err
		}

		if len(spinningWheelMysteries) == 0 {
			err := fmt.Errorf("spinning wheel mysteries is empty")
			err = errors.ErrInternalServerError.Wrap(err, err.Error())
			return dto.PlaceSpinningWheelResp{}, err
		}

		// get Spinning Wheel Mystery probability

		mystery, err := b.GetSpinningWheelMysteryProbability(ctx)
		if err != nil {
			return dto.PlaceSpinningWheelResp{}, err
		}

		// if mystery is better return better next time
		if mystery.Type == dto.SpinningWheelMysteryTypesBetter {
			rp, err := b.betStorage.CreateSpinningWheel(ctx, dto.SpinningWheelData{
				UserID:    userID,
				Status:    constant.CLOSED,
				BetAmount: betAmount,
				WonAmount: wonAmount,
				WonStatus: constant.LOSE,
				Type:      string(dto.SpinningWheelMysteryTypesBetter),
			})

			if err != nil {
				return dto.PlaceSpinningWheelResp{}, err
			}

			return dto.PlaceSpinningWheelResp{
				Message: constant.SUCCESS,
				Data: dto.PlaceSpinningWheelData{
					ID:        rp.ID,
					BetAmount: betAmount,
					Prize:     string(dto.SpinningWheelMysteryTypesBetter),
				},
			}, nil
		}

		// if mystery is point return point next time
		switch mystery.Type {
		case dto.SpinningWheelMysteryTypesPoint:
			operationalGroupAndTypeIDs, err := b.CreateOrGetOperationalGroupAndType(ctx, constant.TRANSFER, constant.BET_CASHOUT)
			if err != nil {
				b.log.Error(err.Error())
				err = errors.ErrInternalServerError.Wrap(err, err.Error())
				return dto.PlaceSpinningWheelResp{}, err
			}
			_, err = b.betStorage.CreateSpinningWheel(ctx, dto.SpinningWheelData{
				UserID:    userID,
				Status:    constant.CLOSED,
				BetAmount: betAmount,
				WonAmount: wonAmount,
				WonStatus: constant.WON,
				Type:      string(dto.SpinningWheelMysteryTypesPoint),
			})

			if err != nil {
				return dto.PlaceSpinningWheelResp{}, err
			}
			balance, exist, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
				UserId:   userID,
				Currency: constant.POINT_CURRENCY,
			})

			if err != nil {
				return dto.PlaceSpinningWheelResp{}, err
			}

			if !exist {
				err := fmt.Errorf("user balance not found")
				err = errors.ErrInternalServerError.Wrap(err, err.Error())
				return dto.PlaceSpinningWheelResp{}, err
			}

			// update balance
			newBalance := balance.RealMoney.Add(mystery.Amount)
			_, err = b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
				UserID:    userID,
				Currency:  constant.POINT_CURRENCY,
				Component: constant.REAL_MONEY,
				Amount:    newBalance,
			})

			if err != nil {
				return dto.PlaceSpinningWheelResp{}, err
			}

			b.SaveToSquads(ctx, dto.SquadEarns{
				UserID:   userID,
				Currency: constant.POINT_CURRENCY,
				Earn:     mystery.Amount,
				GameID:   constant.GAME_SPINNING_WHEEL,
			})

			// save operations logs
			transactionID := uuid.New().String()
			_, err = b.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
				UserID:             userID,
				Component:          constant.REAL_MONEY,
				Currency:           constant.POINT_CURRENCY,
				Description:        fmt.Sprintf("cashout spinning wheel bet amount %v, new balance is %v and  currency %s", price.Price, balance.RealMoney.Add(mystery.Amount), constant.POINT_CURRENCY),
				ChangeAmount:       price.Price,
				OperationalGroupID: operationalGroupAndTypeIDs.OperationalGroupID,
				OperationalTypeID:  operationalGroupAndTypeIDs.OperationalTypeID,
				BalanceAfterUpdate: &newBalance,
				TransactionID:      &transactionID,
			})
			if err != nil {
				return dto.PlaceSpinningWheelResp{}, err
			}
			return dto.PlaceSpinningWheelResp{
				Message: constant.SUCCESS,
				Data: dto.PlaceSpinningWheelData{
					ID:        mystery.ID,
					BetAmount: betAmount,
					Prize:     fmt.Sprintf("point %v", mystery.Amount),
				},
			}, nil
		case dto.SpinningWheelMysteryTypesInternetPackageInGB:
			// close bet
			_, err = b.betStorage.CreateSpinningWheel(ctx, dto.SpinningWheelData{
				UserID:    userID,
				Status:    constant.CLOSED,
				BetAmount: betAmount,
				WonAmount: wonAmount,
				WonStatus: constant.WON,
				Type:      string(dto.SpinningWheelMysteryTypesInternetPackageInGB),
			})

			if err != nil {
				return dto.PlaceSpinningWheelResp{}, err
			}

			return dto.PlaceSpinningWheelResp{
				Message: constant.SUCCESS,
				Data: dto.PlaceSpinningWheelData{
					ID:        mystery.ID,
					BetAmount: betAmount,
					Prize:     fmt.Sprintf("internet package %v", mystery.Amount),
				},
			}, nil
		case dto.SpinningWheelMysteryTypesBetter:
			// close bet
			_, err = b.betStorage.CreateSpinningWheel(ctx, dto.SpinningWheelData{
				UserID:    userID,
				Status:    constant.CLOSED,
				BetAmount: betAmount,
				WonAmount: wonAmount,
				WonStatus: constant.WON,
				Type:      string(dto.SpinningWheelMysteryTypesBetter),
			})

			if err != nil {
				return dto.PlaceSpinningWheelResp{}, err
			}

			return dto.PlaceSpinningWheelResp{
				Message: constant.SUCCESS,
				Data: dto.PlaceSpinningWheelData{
					ID:        mystery.ID,
					BetAmount: betAmount,
					Prize:     fmt.Sprintf("better %v", mystery.Amount),
				},
			}, nil
		case dto.SpinningWheelMysteryTypesSpin:
			// close bet
			_, err = b.betStorage.CreateSpinningWheel(ctx, dto.SpinningWheelData{
				UserID:    userID,
				Status:    constant.CLOSED,
				BetAmount: betAmount,
				WonAmount: wonAmount,
				WonStatus: constant.WON,
				Type:      string(dto.SpinningWheelMysteryTypesSpin),
			})

			if err != nil {
				return dto.PlaceSpinningWheelResp{}, err
			}

			b.spinningwheelFreeSpins[userID] = b.spinningwheelFreeSpins[userID] + int(mystery.Amount.IntPart())

			return dto.PlaceSpinningWheelResp{
				Message: constant.SUCCESS,
				Data: dto.PlaceSpinningWheelData{
					ID:        mystery.ID,
					BetAmount: betAmount,
					Prize:     fmt.Sprintf("spin %v", mystery.Amount),
				},
			}, nil
		}
	case dto.Better:
		// close bet
		_, err = b.betStorage.CreateSpinningWheel(ctx, dto.SpinningWheelData{
			UserID:    userID,
			Status:    constant.CLOSED,
			BetAmount: betAmount,
			WonAmount: wonAmount,
			WonStatus: constant.WON,
			Type:      string(dto.SpinningWheelMysteryTypesBetter),
		})

		if err != nil {
			return dto.PlaceSpinningWheelResp{}, err
		}

		return dto.PlaceSpinningWheelResp{
			Message: constant.SUCCESS,
			Data: dto.PlaceSpinningWheelData{
				ID:        resp.ID,
				BetAmount: betAmount,
				Prize:     string(dto.Better),
			},
		}, nil
	case dto.Point:
		operationalGroupAndTypeIDs, err := b.CreateOrGetOperationalGroupAndType(ctx, constant.TRANSFER, constant.BET_CASHOUT)
		if err != nil {
			b.log.Error(err.Error())
			err = errors.ErrInternalServerError.Wrap(err, err.Error())
			return dto.PlaceSpinningWheelResp{}, err
		}
		_, err = b.betStorage.CreateSpinningWheel(ctx, dto.SpinningWheelData{
			UserID:    userID,
			Status:    constant.CLOSED,
			BetAmount: betAmount,
			WonAmount: wonAmount,
			WonStatus: constant.WON,
			Type:      string(dto.SpinningWheelMysteryTypesPoint),
		})

		if err != nil {
			return dto.PlaceSpinningWheelResp{}, err
		}
		balance, exist, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
			UserId:   userID,
			Currency: constant.POINT_CURRENCY,
		})

		if err != nil {
			return dto.PlaceSpinningWheelResp{}, err
		}

		if !exist {
			err := fmt.Errorf("user balance not found")
			err = errors.ErrInternalServerError.Wrap(err, err.Error())
			return dto.PlaceSpinningWheelResp{}, err
		}

		// update balance
		newBalance := balance.RealMoney.Add(resp.Amount)
		_, err = b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    userID,
			Currency:  constant.POINT_CURRENCY,
			Component: constant.REAL_MONEY,
			Amount:    newBalance,
		})
		if err != nil {
			return dto.PlaceSpinningWheelResp{}, err
		}

		// save operations logs
		transactionID := uuid.New().String()
		_, err = b.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
			UserID:             userID,
			Component:          constant.REAL_MONEY,
			Currency:           constant.POINT_CURRENCY,
			Description:        fmt.Sprintf("cashout spinning wheel bet amount %v, new balance is %v and  currency %s", price.Price, balance.RealMoney.Add(resp.Amount), constant.POINT_CURRENCY),
			ChangeAmount:       price.Price,
			OperationalGroupID: operationalGroupAndTypeIDs.OperationalGroupID,
			OperationalTypeID:  operationalGroupAndTypeIDs.OperationalTypeID,
			BalanceAfterUpdate: &newBalance,
			TransactionID:      &transactionID,
		})
		if err != nil {
			return dto.PlaceSpinningWheelResp{}, err
		}
		return dto.PlaceSpinningWheelResp{
			Message: constant.SUCCESS,
			Data: dto.PlaceSpinningWheelData{
				ID:        resp.ID,
				BetAmount: betAmount,
				Prize:     fmt.Sprintf("point %v", resp.Amount),
			},
		}, nil
	case dto.InternetPackageInGB:
		// close bet
		_, err = b.betStorage.CreateSpinningWheel(ctx, dto.SpinningWheelData{
			UserID:    userID,
			Status:    constant.CLOSED,
			BetAmount: betAmount,
			WonAmount: wonAmount,
			WonStatus: constant.WON,
			Type:      string(dto.SpinningWheelMysteryTypesInternetPackageInGB),
		})

		if err != nil {
			return dto.PlaceSpinningWheelResp{}, err
		}

		return dto.PlaceSpinningWheelResp{
			Message: constant.SUCCESS,
			Data: dto.PlaceSpinningWheelData{
				ID:        resp.ID,
				BetAmount: betAmount,
				Prize:     string(dto.InternetPackageInGB),
			},
		}, nil
	case dto.Spin:
		_, err = b.betStorage.CreateSpinningWheel(ctx, dto.SpinningWheelData{
			UserID:    userID,
			Status:    constant.CLOSED,
			BetAmount: betAmount,
			WonAmount: wonAmount,
			WonStatus: constant.WON,
			Type:      string(dto.SpinningWheelMysteryTypesSpin),
		})

		if err != nil {
			return dto.PlaceSpinningWheelResp{}, err
		}

		b.spinningwheelFreeSpins[userID] = b.spinningwheelFreeSpins[userID] + int(resp.Amount.IntPart())

		return dto.PlaceSpinningWheelResp{
			Message: constant.SUCCESS,
			Data: dto.PlaceSpinningWheelData{
				ID:        resp.ID,
				BetAmount: betAmount,
				Prize:     fmt.Sprintf("spin %v", resp.Amount),
			},
		}, nil
	}

	return dto.PlaceSpinningWheelResp{}, nil
}

func (b *bet) GetSpinningWheel(ctx context.Context) (dto.SpinningWheelConfigData, error) {
	spinningWheelConfigs, err := b.betStorage.GetAllSpinningWheelConfigs(ctx)
	if err != nil {
		return dto.SpinningWheelConfigData{}, err
	}

	if len(spinningWheelConfigs) == 0 {
		b.log.Error("spinning wheel configs not found")
		err := errors.ErrInternalServerError.Wrap(fmt.Errorf("spinning wheel configs not found"), "spinning wheel configs not found")
		return dto.SpinningWheelConfigData{}, err
	}

	var configs []dto.SpinningWheelConfigData
	for _, config := range spinningWheelConfigs {
		if config.Status == constant.ACTIVE {
			configs = append(configs, config)
		}
	}

	if len(configs) == 0 {
		b.log.Error("spinning wheel configs not found")
		err := errors.ErrInternalServerError.Wrap(fmt.Errorf("spinning wheel configs not found"), "spinning wheel configs not found")
		return dto.SpinningWheelConfigData{}, err
	}

	totalFrequency := 0
	for _, config := range configs {
		if config.Frequency < 0 {
			b.log.Error("invalid frequency for config", zap.Any("config", config))
			continue
		}
		totalFrequency += config.Frequency
	}

	if totalFrequency == 0 {
		b.log.Error("total frequency is zero")
		err := errors.ErrInternalServerError.Wrap(fmt.Errorf("total frequency is zero"), "no valid frequencies found")
		return dto.SpinningWheelConfigData{}, err
	}

	rand.Seed(time.Now().UnixNano())

	randValue := rand.Intn(totalFrequency)

	currentWeight := 0
	for _, config := range configs {
		currentWeight += config.Frequency
		if randValue < currentWeight {
			return config, nil
		}
	}

	return configs[len(configs)-1], nil
}

func (b *bet) GetSpinningWheelMysteryProbability(ctx context.Context) (dto.SpinningWheelMysteryResData, error) {
	spinningWheelConfigs, err := b.betStorage.GetAllSpinningWheelMysteries(ctx)
	if err != nil {
		return dto.SpinningWheelMysteryResData{}, err
	}

	if len(spinningWheelConfigs) == 0 {
		b.log.Error("spinning wheel configs not found")
		err := errors.ErrInternalServerError.Wrap(fmt.Errorf("spinning wheel configs not found"), "spinning wheel configs not found")
		return dto.SpinningWheelMysteryResData{}, err
	}

	var configs []dto.SpinningWheelMysteryResData
	for _, config := range spinningWheelConfigs {
		if config.Status == constant.ACTIVE {
			configs = append(configs, config)
		}
	}

	if len(configs) == 0 {
		configs = append(configs, dto.SpinningWheelMysteryResData{
			ID:        uuid.UUID{},
			Amount:    decimal.Decimal{},
			Type:      dto.SpinningWheelMysteryTypesBetter,
			Frequency: 1,
			Status:    constant.ACTIVE,
		})
	}

	totalFrequency := 0
	for _, config := range configs {
		if config.Frequency < 0 {
			b.log.Error("invalid frequency for config", zap.Any("config", config))
			continue
		}
		totalFrequency += config.Frequency
	}

	if totalFrequency == 0 {
		b.log.Error("total frequency is zero")
		err := errors.ErrInternalServerError.Wrap(fmt.Errorf("total frequency is zero"), "no valid frequencies found")
		return dto.SpinningWheelMysteryResData{}, err
	}

	rand.Seed(time.Now().UnixNano())

	randValue := rand.Intn(totalFrequency)

	currentWeight := 0
	for _, config := range configs {
		currentWeight += config.Frequency
		if randValue < currentWeight {
			return config, nil
		}
	}

	return configs[len(configs)-1], nil
}

func (b *bet) SaveSpinningWheelPrize(ctx context.Context, userID uuid.UUID, wonAmount decimal.Decimal) error {
	// update user balance
	defer b.userWS.TriggerBalanceWS(ctx, userID)

	balance, _, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		UserId:   userID,
		Currency: constant.POINT_CURRENCY,
	})

	if err != nil {
		return err
	}
	balanceAfterWin := balance.RealMoney.Add(wonAmount)
	updatedBalance, err := b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
		UserID:    userID,
		Currency:  constant.POINT_CURRENCY,
		Component: constant.REAL_MONEY,
		Amount:    balanceAfterWin,
	})
	if err != nil {
		return err
	}
	// save cashout balance logs
	operationalGroupAndTypeIDsResp, err := b.CreateOrGetOperationalGroupAndType(ctx, constant.TRANSFER, constant.BET_CASHOUT)
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		// reverse balance
		_, err := b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    userID,
			Currency:  constant.POINT_CURRENCY,
			Component: constant.REAL_MONEY,
			Amount:    balance.RealMoney,
		})
		// reverse cashout

		return err
	}
	// save balance log
	transactionId := utils.GenerateTransactionId()

	b.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
		UserID:             userID,
		Component:          constant.REAL_MONEY,
		Currency:           constant.POINT_CURRENCY,
		Description:        fmt.Sprintf("cash out spinning wheels bet  %v  amount, new balance is %v s currency balance is  %s", wonAmount, updatedBalance.RealMoney, constant.POINT_CURRENCY),
		ChangeAmount:       updatedBalance.RealMoney,
		OperationalGroupID: operationalGroupAndTypeIDsResp.OperationalGroupID,
		OperationalTypeID:  operationalGroupAndTypeIDsResp.OperationalTypeID,
		BalanceAfterUpdate: &balanceAfterWin,
		TransactionID:      &transactionId,
	})
	return nil
}

func (b *bet) GetSpinningWheelUserBetHistory(ctx context.Context, req dto.GetRequest, userID uuid.UUID) (dto.GetSpinningWheelHistoryResp, error) {
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset

	resp, _, err := b.betStorage.GetSpinningWheelUserBetHistory(ctx, req, userID)
	if err != nil {
		return dto.GetSpinningWheelHistoryResp{}, err
	}

	return resp, nil
}

func (b *bet) CreateSpinningWheelMystery(ctx context.Context, req dto.CreateSpinningWheelMysteryReq) (dto.CreateSpinningWheelMysteryRes, error) {
	if req.Frequency <= 0 {
		req.Frequency = 1
	}

	if req.Amount.LessThan(decimal.Zero) {
		req.Amount = decimal.Zero
	}

	if req.Frequency < 0 {
		req.Frequency = 0
	}

	if req.Type != dto.SpinningWheelMysteryTypesPoint &&
		req.Type != dto.SpinningWheelMysteryTypesInternetPackageInGB &&
		req.Type != dto.SpinningWheelMysteryTypesBetter &&
		req.Type != dto.SpinningWheelMysteryTypesSpin {
		b.log.Error(fmt.Sprintf("invalid type %v", req.Type), zap.Any("req", req))
		err := errors.ErrInvalidUserInput.Wrap(nil, fmt.Sprintf("invalid type %v", req.Type))
		return dto.CreateSpinningWheelMysteryRes{}, err
	}

	return b.betStorage.CreateSpinningWheelMystery(ctx, req)
}

func (b *bet) GetSpinningWheelMystery(ctx context.Context, req dto.GetRequest) (dto.GetSpinningWheelMysteryRes, error) {
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset
	return b.betStorage.GetSpinningWheelMystery(ctx, req)
}

func (b *bet) UpdateSpinningWheelMystery(ctx context.Context, req dto.UpdateSpinningWheelMysteryReq) (dto.UpdateSpinningWheelMysteryRes, error) {
	return b.betStorage.UpdateSpinningWheelMystery(ctx, req)
}

func (b *bet) DeleteSpinningWheelMystery(ctx context.Context, req dto.DeleteReq) error {
	return b.betStorage.DeleteSpinningWheelMystery(ctx, req.ID)
}

func (b *bet) CreateSpinningWheelConfig(ctx context.Context, req dto.CreateSpinningWheelConfigReq) (dto.CreateSpinningWheelConfigRes, error) {
	if req.Frequency <= 0 {
		req.Frequency = 1
	}

	if req.Amount.LessThan(decimal.Zero) {
		req.Amount = decimal.Zero
	}

	if req.Type != dto.Point &&
		req.Type != dto.InternetPackageInGB &&
		req.Type != dto.Better &&
		req.Type != dto.Mystery &&
		req.Type != dto.Spin {
		b.log.Error(fmt.Sprintf("invalid type %v", req.Type), zap.Any("req", req))
		err := errors.ErrInvalidUserInput.Wrap(nil, fmt.Sprintf("invalid type %v", req.Type))
		return dto.CreateSpinningWheelConfigRes{}, err
	}

	return b.betStorage.CreateSpinningWheelConfig(ctx, req)
}

func (b *bet) GetSpinningWheelConfig(ctx context.Context, req dto.GetRequest) (dto.GetSpinningWheelConfigRes, error) {
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset
	return b.betStorage.GetSpinningWheelConfig(ctx, req)
}

func (b *bet) UpdateSpinningWheelConfig(ctx context.Context, req dto.UpdateSpinningWheelConfigReq) (dto.UpdateSpinningWheelConfigRes, error) {
	if req.Frequency <= 0 {
		req.Frequency = 1
	}

	if req.Amount.LessThan(decimal.Zero) {
		req.Amount = decimal.Zero
	}

	if req.Type != dto.Point &&
		req.Type != dto.InternetPackageInGB &&
		req.Type != dto.Better &&
		req.Type != dto.Mystery &&
		req.Type != dto.Spin {
		b.log.Error(fmt.Sprintf("invalid type %v", req.Type), zap.Any("req", req))
		err := errors.ErrInvalidUserInput.Wrap(nil, fmt.Sprintf("invalid type %v", req.Type))
		return dto.UpdateSpinningWheelConfigRes{}, err
	}
	return b.betStorage.UpdateSpinningWheelConfig(ctx, req)
}

func (b *bet) DeleteSpinningWheelConfig(ctx context.Context, req dto.DeleteReq) error {
	return b.betStorage.DeleteSpinningWheelConfig(ctx, req.ID)
}

func (u *bet) UploadBetIcons(ctx context.Context, img multipart.File, header *multipart.FileHeader) (dto.UploadIconsResp, error) {
	// Extract the original file name and get the extension
	fileExtension := filepath.Ext(header.Filename)
	if fileExtension == "" {
		err := fmt.Errorf("invalid file extension")
		u.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UploadIconsResp{}, err
	}

	profilePictureName := uuid.New().String() + fileExtension

	// Create S3 instance
	s3Instance := utils.NewS3Instance(u.log, constant.VALID_IMGS)
	if s3Instance == nil {
		err := fmt.Errorf("unable to create s3 session")
		u.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.UploadIconsResp{}, err
	}

	_, err := s3Instance.UploadToS3Bucket(u.bucketName, img, profilePictureName, header.Header.Get("Content-Type"))
	if err != nil {
		u.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.UploadIconsResp{}, err
	}

	return dto.UploadIconsResp{
		Status: constant.SUCCESS,
		Url:    fmt.Sprintf("https://%s.s3.amazonaws.com/%s", u.bucketName, profilePictureName),
	}, nil
}
