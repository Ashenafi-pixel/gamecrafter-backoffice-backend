package bet

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/platform/utils"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

func (b *bet) SetCrytoKingsConfig(ctx context.Context, req dto.UpdateCryptoKingsConfigReq) (dto.UpdateCryptokingsConfigRes, error) {

	// update configuration
	// get if the config is exist if not exist create the config
	if err := b.CreateOrUpdateConfigForDecimals(ctx, constant.CRYPTO_KINGS_RANGE_MAX_VALUE, constant.CRYPTO_KINGS_RANGE_MAX_VALUE_DEFAULT, req.CryptoKingsRangeMaxValue); err != nil {
		return dto.UpdateCryptokingsConfigRes{}, err
	}
	// update crypto kings time multiplier
	if err := b.CreateOrUpdateConfigForDecimals(ctx, constant.CRYPTO_KINGS_TIME_MAX_VALUE, constant.CRYPTO_KINGS_TIME_MAX_VALUE_DEFAULT, req.CryptoKingsRangeMaxValue); err != nil {
		return dto.UpdateCryptokingsConfigRes{}, err
	}
	return dto.UpdateCryptokingsConfigRes{
		Message: constant.SUCCESS,
		Data: dto.UpdateCryptoKingsConfigResData{
			CryptoKingsRangeMaxValue: req.CryptoKingsRangeMaxValue,
			CryptoKingsTimeMaxValue:  req.CryptoKingsRangeMaxValue,
		},
	}, nil
}

func (b *bet) CreateOrUpdateConfigForDecimals(ctx context.Context, name, defaultValue string, newValue decimal.Decimal) error {
	resp, exist, err := b.ConfigStorage.GetConfigByName(ctx, constant.CRYPTO_KINGS_RANGE_MAX_VALUE)

	if err != nil {
		return err
	}

	if !exist {
		// create new  with the given value
		if newValue.GreaterThan(decimal.Zero) {
			if _, err := b.ConfigStorage.CreateConfig(ctx, dto.Config{
				Name:  constant.CRYPTO_KINGS_RANGE_MAX_VALUE,
				Value: newValue.String(),
			},
			); err != nil {
				return err
			}
		} else {
			// create  config with default value
			if _, err := b.ConfigStorage.CreateConfig(ctx, dto.Config{
				Name:  constant.CRYPTO_KINGS_RANGE_MAX_VALUE,
				Value: constant.CRYPTO_KINGS_RANGE_MAX_VALUE_DEFAULT,
			},
			); err != nil {
				return err
			}
		}
	} else {
		// update the value
		if newValue.GreaterThan(decimal.Zero) {
			_, err := b.ConfigStorage.UpdateConfigByID(ctx, dto.Config{
				ID:    resp.ID,
				Name:  constant.CRYPTO_KINGS_RANGE_MAX_VALUE,
				Value: newValue.String(),
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *bet) PlaceCryptoKingsBet(ctx context.Context, req dto.PlaceCryptoKingsBetReq, userID uuid.UUID) (dto.PlaceCryptoKingsBetRes, error) {
	defer b.TriggerLevelResponse(ctx, userID)
	defer b.TriggerPlayerProgressBar(ctx, userID)
	defer b.InitiateTriggerSquadsProgressBar(ctx, userID)
	defer b.userWS.TriggerBalanceWS(ctx, userID)
	userLock := b.getUserLock(userID)
	userLock.Lock()
	var wonOrLose bool
	defer userLock.Unlock()

	open, err := b.CheckBetLockStatus(ctx, constant.GAME_CRYPTO_KINGS)
	if err != nil {
		return dto.PlaceCryptoKingsBetRes{}, err
	}

	if !open {
		err := fmt.Errorf("game currently unavailable")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.PlaceCryptoKingsBetRes{}, err
	}

	var response dto.PlaceCryptoKingsBetRes
	//check for the users is blocked or not
	if err := b.CheckGameBlocks(ctx, userID); err != nil {
		return dto.PlaceCryptoKingsBetRes{}, err
	}

	// validate connections is available or not
	conns, ok := b.userSingleGameConnection[userID]
	if !ok {
		err := fmt.Errorf("no active WS connection available to stream multipliers for user please use /ws/single/player endpoint to connect ")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.PlaceCryptoKingsBetRes{}, err
	}

	// validate possibe types

	if req.Type == "" || !constant.CRYPTO_KINGS_PLACE_BET_TYPES[req.Type] {
		m := "invalid type is given, possible types are "
		for k := range constant.CRYPTO_KINGS_PLACE_BET_TYPES {
			m = m + " " + k
		}
		err := fmt.Errorf("error: %s", m)
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.PlaceCryptoKingsBetRes{}, err
	}
	// if type is range check validate range
	if req.Type == constant.CRYPTO_KING_RANGE && (req.MinValue.LessThanOrEqual(decimal.Zero) || req.MaxValue.LessThanOrEqual(decimal.Zero)) {
		err := fmt.Errorf("error: for type range min and max value can not zero ")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.PlaceCryptoKingsBetRes{}, err
	}
	if req.Type == constant.CRYPTO_KING_RANGE && req.MaxValue.LessThanOrEqual(req.MinValue) {
		// min should be greater than max
		err := fmt.Errorf("error: for type range min value can not be less than or equal to max value ")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.PlaceCryptoKingsBetRes{}, err

	}

	if req.Type == constant.CRYPTO_KING_TIME && (req.Second <= 0 || req.Second > 10) {
		// second should be between 1 to 10
		err := fmt.Errorf("error: for type time second must be between 1 to 10")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.PlaceCryptoKingsBetRes{}, err

	}

	//check if the given max value  are not less than zero or equal to min
	if req.BetAmount <= 0 {
		err := fmt.Errorf("bet amount can not be less than or equal to zero")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.PlaceCryptoKingsBetRes{}, err
	}

	// check balance
	balance, exist, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		UserId:   userID,
		CurrencyCode: constant.POINT_CURRENCY,
	})

	if err != nil {
		return dto.PlaceCryptoKingsBetRes{}, err
	}

	if !exist || balance.AmountUnits.LessThan(decimal.NewFromInt(req.BetAmount)) {
		err := fmt.Errorf("insufficient balance ")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.PlaceCryptoKingsBetRes{}, err
	}

	// check if user should be lose or not
	if _, ok := b.crptoKingsWonLoseMap[userID]; !ok {
		// generate random lose amount
		loseAmounts, err := utils.GenerateRandomCrashPoint(1, 6)
		if err != nil {
			b.log.Error(err.Error())
			err = errors.ErrInternalServerError.Wrap(err, err.Error())
			return dto.PlaceCryptoKingsBetRes{}, err
		}

		b.crptoKingsWonLoseMap[userID] = int(loseAmounts.IntPart())
		b.cryptoKingsCurrentUsersValue[userID] = decimal.NewFromInt(constant.CRYPTO_KING_DEFAULT_CURRENT_VALUE)
	}

	// if req.type == HIGHER or req.type LOWER
	if req.Type == constant.CRYPTO_KING_HIGHER || req.Type == constant.CRYPTO_KING_LOWER {
		// check lose value if lose value == 0 then the user should win else user should lose
		resp, postibleWon, err := b.CalculateLoseOrWonAmountForHighOrLowOptions(ctx, userID, req)
		if err != nil {
			return dto.PlaceCryptoKingsBetRes{}, err
		}
		if resp.WonStatus == constant.WON {
			wonOrLose = true
		} else {
			wonOrLose = false
		}
		go b.GenerateAndStreamRandomBetValues(dto.StreamTradingReq{
			ID:          resp.ID,
			UserID:      userID,
			StartValue:  b.cryptoKingsCurrentUsersValue[userID],
			TargetValue: resp.EndCryptoValue,
			TypeSecond:  false,
			UserConns:   conns,
			Second:      10,
			WonStatus:   resp.WonStatus,
			WonAmount:   int(resp.WonAmount.IntPart()),
		})

		response = dto.PlaceCryptoKingsBetRes{
			Message: constant.SUCCESS,
			Data: dto.PlaceCryptoKingsBetResData{
				ID:                 resp.ID,
				UserID:             resp.UserID,
				BetAmount:          resp.BetAmount,
				Type:               req.Type,
				PotentialWinAmount: postibleWon,
				Timestamp:          resp.Timestamp,
			},
		}

	} else if req.Type == constant.CRYPTO_KING_RANGE || req.Type == constant.CRYPTO_KING_TIME {
		// if req.type == HIGHER or req.type LOWER

		// calculate range
		var isTypeTime bool
		var stopTime int
		var wonAmount int64
		var won bool

		if req.Type == constant.CRYPTO_KING_RANGE {
			stopTime = 10
			wonAmount, won, err = b.CalculateCryptoKingsRangeAndTimeWon(ctx, userID, int(req.MaxValue.Sub(req.MinValue).IntPart()), int(req.BetAmount), constant.CRYPTO_KINGS_RANGE_MAX_VALUE, constant.CRYPTO_KINGS_RANGE_MAX_VALUE_DEFAULT)
		} else {
			isTypeTime = true
			stopTime = req.Second
			wonAmount, won, err = b.CalculateCryptoKingsRangeAndTimeWon(ctx, userID, req.Second, int(req.BetAmount), constant.CRYPTO_KINGS_TIME_MAX_VALUE, constant.CRYPTO_KINGS_TIME_MAX_VALUE_DEFAULT)

		}

		if err != nil {
			return dto.PlaceCryptoKingsBetRes{}, err
		}

		// trigger level
		b.TriggerLevelResponse(ctx, userID)
		b.TriggerPlayerProgressBar(ctx, userID)

		// save to database
		if won {
			wonOrLose = true
			var cryptEndValue float64
			cryptoStartValue := b.cryptoKingsCurrentUsersValue[userID]

			if req.Type == constant.CRYPTO_KING_RANGE {
				cryptEndValue = utils.RandomDecimal(int(req.MinValue.IntPart()), int(req.MaxValue.IntPart()))
			} else {
				cryptEndValue = utils.RandomDecimal(int(cryptoStartValue.IntPart()), int(cryptoStartValue.IntPart()+200))
			}
			b.cryptoKingsCurrentUsersValue[userID] = decimal.NewFromFloat(cryptEndValue).Round(2)
			b.cryptoKingsRangeBets[userID] = b.cryptoKingsRangeBets[userID] - int(wonAmount)

			resp, err := b.betStorage.CreateCryptoKings(ctx, dto.CreateCryptoKingData{
				UserID:           userID,
				Status:           constant.CLOSED,
				BetAmount:        req.BetAmount,
				Type:             req.Type,
				WonStatus:        constant.WON,
				Timestamp:        time.Now(),
				StartCryptoValue: cryptoStartValue,
				EndCryptoValue:   decimal.NewFromFloat(cryptEndValue).Round(2),
				WonAmount:        decimal.NewFromInt(wonAmount),
			})
			if err != nil {
				return dto.PlaceCryptoKingsBetRes{}, err
			}

			go b.GenerateAndStreamRandomBetValues(dto.StreamTradingReq{
				ID:          resp.ID,
				UserID:      userID,
				StartValue:  cryptoStartValue,
				TargetValue: decimal.NewFromFloat(cryptEndValue).Round(2),
				TypeSecond:  isTypeTime,
				UserConns:   conns,
				Second:      stopTime,
				WonStatus:   constant.WON,
				WonAmount:   int(resp.WonAmount.IntPart()),
			})

			response = dto.PlaceCryptoKingsBetRes{
				Message: constant.SUCCESS,
				Data: dto.PlaceCryptoKingsBetResData{
					ID:                 resp.ID,
					UserID:             userID,
					BetAmount:          req.BetAmount,
					Type:               req.Type,
					PotentialWinAmount: wonAmount,
					Timestamp:          resp.Timestamp,
				},
			}
		} else {
			wonOrLose = false
			cryptoStartValue := b.cryptoKingsCurrentUsersValue[userID]
			var cryptEndValue float64
			cryptEndValue = utils.RandomDecimal(1, 6)
			if decimal.NewFromFloat(cryptEndValue).LessThanOrEqual(decimal.NewFromInt(3)) && cryptoStartValue.GreaterThan(decimal.NewFromInt(300)) {
				cryptEndValue = utils.RandomDecimal(int(req.MinValue.IntPart()-200), int(req.MinValue.IntPart()-2))
			} else {
				cryptEndValue = utils.RandomDecimal(int(req.MaxValue.IntPart()+2), int(req.MaxValue.IntPart()+200))
			}
			if req.Type == constant.CRYPTO_KING_TIME {
				cryptEndValue = utils.RandomDecimal(int(cryptoStartValue.IntPart()), int(cryptoStartValue.IntPart()+200))
			}
			b.cryptoKingsRangeBets[userID] = b.cryptoKingsRangeBets[userID] + int(req.BetAmount)
			b.cryptoKingsCurrentUsersValue[userID] = decimal.NewFromFloat(cryptEndValue)

			resp, err := b.betStorage.CreateCryptoKings(ctx, dto.CreateCryptoKingData{
				UserID:             userID,
				Status:             constant.CLOSED,
				BetAmount:          req.BetAmount,
				Type:               req.Type,
				WonStatus:          constant.LOSE,
				Timestamp:          time.Now(),
				StartCryptoValue:   cryptoStartValue,
				EndCryptoValue:     decimal.NewFromFloat(cryptEndValue),
				SelectedStartValue: req.MinValue,
				SelectedEndValue:   req.MaxValue,
				WonAmount:          decimal.Zero,
			})
			if err != nil {
				return dto.PlaceCryptoKingsBetRes{}, err
			}

			go b.GenerateAndStreamRandomBetValues(dto.StreamTradingReq{
				ID:          resp.ID,
				UserID:      userID,
				StartValue:  cryptoStartValue,
				TargetValue: decimal.NewFromFloat(cryptEndValue),
				TypeSecond:  isTypeTime,
				UserConns:   conns,
				Second:      stopTime,
				WonStatus:   constant.LOSE,
				WonAmount:   int(resp.WonAmount.IntPart()),
			})

			response = dto.PlaceCryptoKingsBetRes{
				Message: constant.SUCCESS,
				Data: dto.PlaceCryptoKingsBetResData{
					ID:                 resp.ID,
					UserID:             userID,
					BetAmount:          req.BetAmount,
					Type:               req.Type,
					PotentialWinAmount: wonAmount,
					Timestamp:          resp.Timestamp,
				},
			}
		}
	}

	// do transaction
	//update user balance
	newBalance := balance.AmountUnits.Sub(decimal.NewFromInt(req.BetAmount))
	transactionID := utils.GenerateTransactionId()
	_, err = b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
		UserID:    userID,
		Currency:  constant.POINT_CURRENCY,
		Component: constant.REAL_MONEY,
		Amount:    newBalance,
	})
	if err != nil {
		return dto.PlaceCryptoKingsBetRes{}, err
	}

	// save transaction logs
	operationalGroupAndTypeIDs, err := b.CreateOrGetOperationalGroupAndType(ctx, constant.TRANSFER, constant.BET_OPERATIONAL_TYPE)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("placeBetReq", req))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    userID,
			Currency:  constant.POINT_CURRENCY,
			Component: constant.REAL_MONEY,
			Amount:    balance.AmountUnits,
		})
		return dto.PlaceCryptoKingsBetRes{}, err
	}

	// save operations logs
	_, err = b.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
		UserID:             userID,
		Component:          constant.REAL_MONEY,
		Currency:           constant.POINT_CURRENCY,
		Description:        fmt.Sprintf("place cryto kings bet amount %v, new balance is %v and  currency %s", req.BetAmount, balance.AmountUnits.Sub(decimal.NewFromInt(req.BetAmount)), constant.POINT_CURRENCY),
		ChangeAmount:       decimal.NewFromInt(req.BetAmount),
		OperationalGroupID: operationalGroupAndTypeIDs.OperationalGroupID,
		OperationalTypeID:  operationalGroupAndTypeIDs.OperationalTypeID,
		BalanceAfterUpdate: &newBalance,
		TransactionID:      &transactionID,
	})
	if err != nil {
		return dto.PlaceCryptoKingsBetRes{}, err
	}

	// update balance after the user won
	if wonOrLose {
		balanceAfterWin := newBalance.Add(decimal.NewFromInt(response.Data.PotentialWinAmount))
		updatedBalance, err := b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    userID,
			Currency:  constant.POINT_CURRENCY,
			Component: constant.REAL_MONEY,
			Amount:    balanceAfterWin,
		})
		if err != nil {
			return dto.PlaceCryptoKingsBetRes{}, err
		}
		// save to squad
		b.SaveToSquads(ctx, dto.SquadEarns{
			UserID:   userID,
			Currency: constant.POINT_CURRENCY,
			Earn:     decimal.NewFromInt(response.Data.PotentialWinAmount),
			GameID:   constant.GAME_CRYPTO_KINGS,
		})
		// save cashout balance logs
		operationalGroupAndTypeIDsResp, err := b.CreateOrGetOperationalGroupAndType(ctx, constant.TRANSFER, constant.BET_CASHOUT)
		if err != nil {
			b.log.Error(err.Error(), zap.Any("cashoutReq", req))
			err = errors.ErrInternalServerError.Wrap(err, err.Error())
			// reverse balance
			_, err := b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
				UserID:    userID,
				Currency:  constant.POINT_CURRENCY,
				Component: constant.REAL_MONEY,
				Amount:    balance.AmountUnits,
			})
			// reverse cashout

			return dto.PlaceCryptoKingsBetRes{}, err
		}
		// save balance log
		transactionId := utils.GenerateTransactionId()

		b.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
			UserID:             userID,
			Component:          constant.REAL_MONEY,
			Currency:           constant.POINT_CURRENCY,
			Description:        fmt.Sprintf("cash out cyrpto kings bet  %v  amount, new balance is %v s currency balance is  %s", response.Data.PotentialWinAmount, updatedBalance.AmountUnits, constant.POINT_CURRENCY),
			ChangeAmount:       updatedBalance.AmountUnits,
			OperationalGroupID: operationalGroupAndTypeIDsResp.OperationalGroupID,
			OperationalTypeID:  operationalGroupAndTypeIDsResp.OperationalTypeID,
			BalanceAfterUpdate: &balanceAfterWin,
			TransactionID:      &transactionId,
		})

	}

	return response, nil
}

func (b *bet) CalculateCryptoWonAmount(betAmount int64) int64 {
	n := float64(betAmount)
	result := (9.0 / 5.0) * n
	return int64(result)
}

func (b *bet) CalculateLoseOrWonAmountForHighOrLowOptions(ctx context.Context, userID uuid.UUID, req dto.PlaceCryptoKingsBetReq) (dto.CreateCryptoKingData, int64, error) {
	cryptoChange, err := utils.GenerateRandomCrashPoint(5, 100)
	defer b.userWS.TriggerBalanceWS(ctx, userID)

	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.CreateCryptoKingData{}, 0, err
	}
	var cryptEndValue decimal.Decimal
	if b.crptoKingsWonLoseMap[userID] == 0 {
		// user won
		loseAmounts, err := utils.GenerateRandomCrashPoint(0, 6)
		if err != nil {
			b.log.Error(err.Error())
			err = errors.ErrInternalServerError.Wrap(err, err.Error())
			return dto.CreateCryptoKingData{}, 0, err
		}

		b.crptoKingsWonLoseMap[userID] = int(loseAmounts.IntPart())

		// add value to the current value of crypto
		cryptoStartValue := b.cryptoKingsCurrentUsersValue[userID]
		if req.Type == constant.CRYPTO_KING_HIGHER {
			cryptEndValue = b.cryptoKingsCurrentUsersValue[userID].Add(cryptoChange)

		} else {
			cryptEndValue = b.cryptoKingsCurrentUsersValue[userID].Sub(cryptoChange)
		}
		b.cryptoKingsCurrentUsersValue[userID] = cryptEndValue.Round(3)

		//save the value to database
		wonAmount := b.CalculateCryptoWonAmount(req.BetAmount)

		resp, err := b.betStorage.CreateCryptoKings(ctx, dto.CreateCryptoKingData{
			UserID:           userID,
			Status:           constant.CLOSED,
			BetAmount:        req.BetAmount,
			Type:             req.Type,
			WonStatus:        constant.WON,
			Timestamp:        time.Now(),
			StartCryptoValue: cryptoStartValue,
			EndCryptoValue:   cryptEndValue,
			WonAmount:        decimal.NewFromInt(wonAmount),
		})

		if err != nil {
			return dto.CreateCryptoKingData{}, 0, err
		}
		return resp, wonAmount, nil
	} else {
		// lose user

		wonAmount := b.CalculateCryptoWonAmount(req.BetAmount)
		cryptoStartValue := b.cryptoKingsCurrentUsersValue[userID]
		if req.Type == constant.CRYPTO_KING_LOWER {
			cryptEndValue = b.cryptoKingsCurrentUsersValue[userID].Add(cryptoChange)
		} else {
			cryptEndValue = b.cryptoKingsCurrentUsersValue[userID].Sub(cryptoChange)
		}

		b.cryptoKingsCurrentUsersValue[userID] = cryptEndValue.Round(3)

		resp, err := b.betStorage.CreateCryptoKings(ctx, dto.CreateCryptoKingData{
			UserID:           userID,
			Status:           constant.CLOSED,
			BetAmount:        req.BetAmount,
			Type:             req.Type,
			WonStatus:        constant.LOSE,
			Timestamp:        time.Now(),
			StartCryptoValue: cryptoStartValue,
			EndCryptoValue:   cryptEndValue,
			WonAmount:        decimal.Zero,
		})

		if err != nil {
			return dto.CreateCryptoKingData{}, 0, err
		}
		b.crptoKingsWonLoseMap[userID]--
		return resp, wonAmount, nil
	}

}

func (b *bet) CalculateCryptoKingsRangeAndTimeWon(ctx context.Context, userID uuid.UUID, rangeSize int, betAmount int, maxValueName string, maxDefualtValue string) (int64, bool, error) {
	// only won if possible won less than total lost
	// get  current lost amount by range
	b.cryptoKingsBetLocker.Lock()
	defer b.cryptoKingsBetLocker.Unlock()
	totalUserLost := b.cryptoKingsRangeBets[userID]
	var maxRangMeltiplier decimal.Decimal

	//possible won amount
	resp, exist, err := b.ConfigStorage.GetConfigByName(ctx, maxValueName)
	if err != nil {
		return 0, false, err
	}

	if !exist || resp.Value == "" {
		// use default value
		maxRangMeltiplier, err = decimal.NewFromString(maxDefualtValue)
		if err != nil {
			b.log.Error(err.Error())
			err = errors.ErrInternalServerError.Wrap(err, err.Error())
			return 0, false, err
		}
	} else {
		maxRangMeltiplier, err = decimal.NewFromString(resp.Value)
		if err != nil {
			b.log.Error(err.Error())
			err = errors.ErrInternalServerError.Wrap(err, err.Error())
			return 0, false, err

		}
	}
	potentialWonAmount := (maxRangMeltiplier.IntPart() / int64(rangeSize)) * int64(betAmount)

	//check if user is valid for get reward
	if potentialWonAmount >= int64(totalUserLost) {
		return potentialWonAmount, false, nil
	} else {
		// allow user to get chance of winning
		randVal, err := utils.GenerateRandomCrashPoint(1, 30)
		if err != nil {
			b.log.Error(err.Error())
			err = errors.ErrInternalServerError.Wrap(err, err.Error())
			return 0, false, err
		}
		if randVal.GreaterThanOrEqual(decimal.NewFromInt(15)) {
			return potentialWonAmount, true, nil
		}
	}

	return potentialWonAmount, true, nil
}

func (b *bet) GenerateAndStreamRandomBetValues(req dto.StreamTradingReq) {
	var isFinal bool
	first := true
	state := dto.TradingState{
		CurrentValue: b.cryptoKingsCurrentUsersValue[req.UserID],
		History:      []decimal.Decimal{req.StartValue},
		StartTime:    time.Now(),
	}

	for {

		state, isFinal = b.GenerateNextTradingValue(req, state)
		if isFinal {
			data := dto.StreamTradingRes{
				ID:           req.ID,
				CurrentValue: state.CurrentValue,
				WonStatus:    req.WonStatus,
				WonAmount:    req.WonAmount,
			}
			byteDate, err := json.Marshal(data)
			if err != nil {
				b.log.Error(err.Error())
				break
			}

			b.StreamToSingleConnection(context.Background(), req.UserConns, byteDate, req.UserID)
			break
		}
		if req.TypeSecond && isFinal && first {
			// stream if the user is lose or won

			data := dto.StreamTradingRes{
				ID:           req.ID,
				CurrentValue: state.CurrentValue,
				WonStatus:    req.WonStatus,
				WonAmount:    req.WonAmount,
			}
			byteDate, err := json.Marshal(data)
			if err != nil {
				b.log.Error(err.Error())
				continue
			}

			b.StreamToSingleConnection(context.Background(), req.UserConns, byteDate, req.UserID)
			// then update state target to fit 10s
			if req.Second < 10 {
				req.Second = 10 - req.Second
				r, _ := utils.GenerateRandomCrashPoint(5, 20)
				req.TargetValue = req.TargetValue.Add(r)
			} else {
				break
			}
			first = false
		}

		//stream the current state of game
		if state.CurrentValue.Equal(decimal.Zero) {
			state.CurrentValue = b.cryptoKingsCurrentUsersValue[req.UserID]
		}
		data := dto.StreamTradingRes{
			ID:           req.ID,
			CurrentValue: state.CurrentValue,
			WonStatus:    constant.PENDING,
		}
		byteDate, err := json.Marshal(data)
		if err != nil {
			b.log.Error(err.Error())
			continue
		}

		b.StreamToSingleConnection(context.Background(), req.UserConns, byteDate, req.UserID)

		time.Sleep(time.Millisecond * 200)
	}

}

func (b *bet) GenerateNextTradingValue(req dto.StreamTradingReq, state dto.TradingState) (dto.TradingState, bool) {

	// Calculate elapsed time
	elapsed := time.Since(state.StartTime).Seconds()
	isFinal := elapsed >= float64(req.Second)

	// Linear interpolation to reach target at 10 seconds, with fluctuations
	totalDuration := 10.0
	progressFraction := math.Min(elapsed/totalDuration, 1.0)
	baseValue := req.StartValue.Add(req.TargetValue.Sub(req.StartValue)).Mul(decimal.NewFromFloat(progressFraction))

	volatility := 0.02 * (1 - math.Abs(2*progressFraction-1)) // Peaks at middle
	fluctuation := decimal.NewFromFloat(rand.NormFloat64()).Mul(decimal.NewFromFloat(volatility)).Mul(req.StartValue)
	if rand.Float64() < 0.05 {
		fluctuation = fluctuation.Mul(decimal.NewFromInt(2))
	}

	// Ensure exact target at 10 seconds
	if isFinal {
		state.CurrentValue = req.TargetValue.Round(2)
	} else {
		state.CurrentValue = baseValue.Add(fluctuation).Round(2)
	}

	state.History = append(state.History, state.CurrentValue)
	if len(state.History) > 50 {
		state.History = state.History[1:]
	}

	return state, isFinal
}

func (b *bet) GetCryptoKingsBetHistory(ctx context.Context, req dto.GetRequest, userID uuid.UUID) (dto.GetCryptoKingsUserBetHistoryRes, error) {
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset
	res, exist, err := b.betStorage.GetCrytoKingsBetHistoryByUserID(ctx, req, userID)
	if err != nil {
		return dto.GetCryptoKingsUserBetHistoryRes{}, err
	}

	if !exist {
		return dto.GetCryptoKingsUserBetHistoryRes{}, nil
	}

	return res, nil
}

func (b *bet) GetCryptoKingsCurrentCryptoPrice(ctx context.Context, userID uuid.UUID) (dto.GetCryptoCurrencyPriceResp, error) {
	if _, ok := b.cryptoKingsCurrentUsersValue[userID]; !ok {

		return dto.GetCryptoCurrencyPriceResp{
			Message: constant.SUCCESS,
			Price:   decimal.NewFromInt(constant.CRYPTO_KING_DEFAULT_CURRENT_VALUE),
		}, nil
	}

	return dto.GetCryptoCurrencyPriceResp{
		Message: constant.SUCCESS,
		Price:   b.cryptoKingsCurrentUsersValue[userID],
	}, nil
}
