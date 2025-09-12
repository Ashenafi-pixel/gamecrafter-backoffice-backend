package bet

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/platform/utils"
	"go.uber.org/zap"
)

func (b *bet) CreateStreetKingsGame(ctx context.Context, req dto.CreateCrashKingsReq, userID uuid.UUID) (dto.CreateStreetKingsResp, error) {
	defer b.TriggerLevelResponse(ctx, userID)
	defer b.TriggerPlayerProgressBar(ctx, userID)
	defer b.InitiateTriggerSquadsProgressBar(ctx, userID)
	defer b.userWS.TriggerBalanceWS(ctx, userID)

	userLock := b.getUserLock(userID)
	userLock.Lock()
	defer userLock.Unlock()
	var randomCrashPoint decimal.Decimal
	open, err := b.CheckBetLockStatus(ctx, constant.GAME_STREET_KINGS)
	if err != nil {
		return dto.CreateStreetKingsResp{}, err
	}

	if !open {
		err := fmt.Errorf("game currently unavailable")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CreateStreetKingsResp{}, err
	}

	versions := ""
	if _, ok := constant.STREET_KINGS_VERSIONS[req.Version]; !ok {
		for k, v := range constant.STREET_KINGS_VERSIONS {
			versions = fmt.Sprintf("%s %s for %s ", versions, k, v)
		}
		err := fmt.Errorf("invalid version given available versions are %s", versions)
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CreateStreetKingsResp{}, err
	}

	balance, exist, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		UserId:       userID,
		CurrencyCode: constant.POINT_CURRENCY,
	})

	if err != nil {
		return dto.CreateStreetKingsResp{}, err
	}

	if !exist {
		err := fmt.Errorf("insufficent point ")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CreateStreetKingsResp{}, err
	}

	if balance.AmountUnits.LessThan(req.BetAmount) {
		err := fmt.Errorf("insufficent point ")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CreateStreetKingsResp{}, err
	}
	if req.BetAmount.LessThan(decimal.NewFromInt(1)) || req.BetAmount.GreaterThan(decimal.NewFromInt(1000)) {
		err := fmt.Errorf("minimum bet is $1 and maximum bet amount is $1000")
		b.log.Warn(err.Error(), zap.Any("placeBetReq", req))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CreateStreetKingsResp{}, err
	}

	if err := b.CheckGameBlocks(ctx, userID); err != nil {
		return dto.CreateStreetKingsResp{}, err
	}

	conns, ok := b.userSingleGameConnection[userID]
	if !ok {
		err := fmt.Errorf("no active WS connection available to stream multipliers for user please use /ws/single/player endpoint to connect ")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CreateStreetKingsResp{}, err
	}
	randomCrashPoint, err = b.GenerateCrashPoint()
	if err != nil {
		return dto.CreateStreetKingsResp{}, err
	}

	newBalance := balance.AmountUnits.Sub(req.BetAmount)
	transactionID := utils.GenerateTransactionId()
	_, err = b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
		UserID:    userID,
		Currency:  constant.POINT_CURRENCY,
		Component: constant.REAL_MONEY,
		Amount:    newBalance,
	})
	if err != nil {
		return dto.CreateStreetKingsResp{}, err
	}

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
		return dto.CreateStreetKingsResp{}, err
	}

	_, err = b.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
		UserID:             userID,
		Component:          constant.REAL_MONEY,
		Currency:           constant.POINT_CURRENCY,
		Description:        fmt.Sprintf("place street kings bet amount %v, new balance is %v and  currency %s", req.BetAmount, balance.AmountUnits.Sub(req.BetAmount), constant.POINT_CURRENCY),
		ChangeAmount:       req.BetAmount,
		OperationalGroupID: operationalGroupAndTypeIDs.OperationalGroupID,
		OperationalTypeID:  operationalGroupAndTypeIDs.OperationalTypeID,
		BalanceAfterUpdate: &newBalance,
		TransactionID:      &transactionID,
	})
	if err != nil {
		return dto.CreateStreetKingsResp{}, err
	}
	// save to squads

	resp, err := b.betStorage.CreateStreetKingsGame(ctx, dto.CreateStreetKingsReqData{
		CreateCrashKingsReq: dto.CreateCrashKingsReq{
			Version:   req.Version,
			BetAmount: req.BetAmount,
		},
		UserID:     userID,
		CrashPoint: randomCrashPoint,
		Status:     constant.FOOTBALL_MATCH_STATUS_ACTIVE,
		Timestamp:  time.Now(),
	})

	if err != nil {
		return dto.CreateStreetKingsResp{}, err
	}

	b.activeSingleBet[resp.ID] = decimal.Zero
	b.activeSingleBetSync[resp.ID] = &sync.Mutex{}
	go b.StreamToSinglePlayerConnection(ctx, conns, resp.ID, resp.CrashPoint, userID, req.BetAmount, time.Now())
	return dto.CreateStreetKingsResp{
		Message: constant.SUCCESS,
		Data: dto.CreateStreetKingsData{
			RoundID:   resp.ID,
			UserID:    userID,
			Version:   resp.Version,
			BetAmount: resp.BetAmount,
			Timestamp: resp.Timestamp,
		},
	}, nil
}

func (b *bet) StreamToSinglePlayerConnection(ctx context.Context, conns map[*websocket.Conn]bool, betID uuid.UUID, crashPoint decimal.Decimal, userID uuid.UUID, betAmount decimal.Decimal, createdAt time.Time) {
	cons := conns
	var err error
	var byteData []byte
	mutex := b.activeSingleBetSync[betID]
	if _, ok := b.singleBetSocketSync[userID]; !ok {
		b.singleBetSocketSync[userID] = &sync.Mutex{}
	}
	b.singleBetSocketSync[userID].Lock()
	defer b.singleBetSocketSync[userID].Unlock()

	mutex.Lock()
	defer mutex.Unlock()

	i := decimal.NewFromInt(0)
	b.activeSingleBet[betID] = i
	for i.LessThanOrEqual(crashPoint) {
		time.Sleep(time.Millisecond * 100)
		for conn := range cons {

			if conn != nil {

				byteData, err = json.Marshal(dto.WSRes{
					Type: constant.WS_CURRENT_MULTIPLIER,
					Data: dto.BetRoundResp{
						ID:          betID,
						Multiplayer: i,
					},
				})
				if err != nil {
					b.log.Error(err.Error())
					continue
				}

				conn.WriteMessage(websocket.TextMessage, byteData)
				if _, ok := b.activeSingleBet[betID]; ok {
					b.activeSingleBet[betID] = i
				} else {
					delete(b.activeSingleBetSync, betID)
					return
				}
			}
		}

		i = i.Add(decimal.NewFromFloat(0.1))

	}
	for conn := range cons {

		if conn != nil {

			byteData, err = json.Marshal(dto.WSRes{
				Type: constant.WS_CURRENT_MULTIPLIER,
				Data: dto.BetRoundResp{
					ID:          betID,
					Multiplayer: i,
				},
			})
			if err != nil {
				b.log.Error(err.Error())
				continue
			}

			conn.WriteMessage(websocket.TextMessage, byteData)
			if _, ok := b.activeSingleBet[betID]; ok {
				b.activeSingleBet[betID] = i
			} else {
				delete(b.activeSingleBetSync, betID)
				return
			}
		}
	}
	delete(b.activeSingleBet, betID)
	for conn := range cons {

		if conn != nil {
			t := time.Now()
			byteData, err = json.Marshal(dto.WSRes{
				Type: constant.WS_CRASH,
				Data: dto.CrashPointRes{
					Round: dto.BetRound{
						ID:         betID,
						Status:     constant.CLOSED,
						CrashPoint: i,
						ClosedAt:   &t,
						UserID:     userID,
						Currency:   constant.POINT_CURRENCY,
						Amount:     betAmount,
						CreatedAt:  &createdAt,
						BetID:      betID,
					},
					Message: constant.BET_CRASHPOINT_REACHED,
				},
			})
			if err != nil {
				b.log.Error(err.Error(), zap.Any("betRound", betID))
			}
			conn.WriteMessage(websocket.TextMessage, byteData)
		}
	}

	b.betStorage.CloseStreetKingsCrash(ctx, dto.StreetKingsCashoutReq{
		ID:        betID,
		Status:    constant.CLOSED,
		WonAmount: decimal.Zero,
	})

}

func (b *bet) CashOutStreetKings(ctx context.Context, req dto.CashOutReq) (dto.StreetKingsCrashResp, error) {
	userLock := b.getUserLock(req.UserID)
	defer b.userWS.TriggerBalanceWS(ctx, req.UserID)

	userLock.Lock()
	defer userLock.Unlock()
	wonAmount := decimal.Zero

	mutex, ok := b.activeSingleBetSync[req.RoundID]
	if !ok {
		err := fmt.Errorf("round closed ")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.StreetKingsCrashResp{}, err
	}

	resp, ok := b.activeSingleBet[req.RoundID]
	if !ok {
		err := fmt.Errorf("round closed")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.StreetKingsCrashResp{}, err
	}
	delete(b.activeSingleBet, req.RoundID)
	delete(b.activeSingleBetSync, req.RoundID)
	mutex.Lock()
	defer mutex.Unlock()

	crash, err := b.betStorage.GetStreetKingsCrashByID(ctx, req.RoundID)
	if err != nil {
		return dto.StreetKingsCrashResp{}, err
	}
	wonAmount = wonAmount.Add(crash.Data.BetAmount.Mul(resp))

	userBalance, _, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		UserId:       req.UserID,
		CurrencyCode: constant.POINT_CURRENCY,
	})

	if err != nil {
		return dto.StreetKingsCrashResp{}, err
	}

	b.betStorage.CloseStreetKingsCrash(ctx, dto.StreetKingsCashoutReq{
		ID:           req.RoundID,
		Status:       constant.CLOSED,
		WonAmount:    wonAmount,
		CashoutPoint: resp,
	})
	crash.Data.WonAmount = wonAmount
	crash.Data.CashoutPoint = resp
	crash.Data.Status = constant.CLOSED

	operationalGroupAndTypeIDs, err := b.CreateOrGetOperationalGroupAndType(ctx, constant.TRANSFER, constant.BET_CASHOUT)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("cashoutReq", req))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())

		_, err := b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    req.UserID,
			Currency:  constant.POINT_CURRENCY,
			Component: constant.REAL_MONEY,
			Amount:    userBalance.AmountUnits,
		})

		b.betStorage.ReverseCashOut(ctx, req.RoundID)
		return dto.StreetKingsCrashResp{}, err
	}

	betAfterUpdate := userBalance.AmountUnits.Add(wonAmount)
	_, err = b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
		UserID:    req.UserID,
		Currency:  constant.POINT_CURRENCY,
		Component: constant.REAL_MONEY,
		Amount:    betAfterUpdate,
	})
	if err != nil {
		return dto.StreetKingsCrashResp{}, err
	}
	b.SaveToSquads(ctx, dto.SquadEarns{
		UserID:   req.UserID,
		Currency: constant.POINT_CURRENCY,
		Earn:     wonAmount,
		GameID:   constant.GAME_CRYPTO_KINGS,
	})
	transactionId := utils.GenerateTransactionId()
	b.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
		UserID:             req.UserID,
		Component:          constant.REAL_MONEY,
		Currency:           constant.POINT_CURRENCY,
		Description:        fmt.Sprintf("cash out street kings bet  %v  amount, new balance is %v s currency balance is  %s", wonAmount, userBalance.AmountUnits.Add(wonAmount), constant.POINT_CURRENCY),
		ChangeAmount:       wonAmount,
		OperationalGroupID: operationalGroupAndTypeIDs.OperationalGroupID,
		OperationalTypeID:  operationalGroupAndTypeIDs.OperationalTypeID,
		BalanceAfterUpdate: &betAfterUpdate,
		TransactionID:      &transactionId,
	})

	return dto.StreetKingsCrashResp{
		Message: crash.Message,
		Data:    crash.Data,
	}, nil
}

func (b *bet) GetStreetkingHistory(ctx context.Context, req dto.GetStreetkingHistoryReq, userID uuid.UUID) (dto.GetStreetkingHistoryRes, error) {

	versions := ""
	if _, ok := constant.STREET_KINGS_VERSIONS[req.Version]; !ok {
		for k, v := range constant.STREET_KINGS_VERSIONS {
			versions = fmt.Sprintf("%s %s for %s ", versions, k, v)
		}
		err := fmt.Errorf("invalid version given available versions are %s", versions)
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.GetStreetkingHistoryRes{}, err
	}

	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset
	resp, _, err := b.betStorage.GetStreetKingsGamesByUserIDAndVersion(ctx, req, userID)

	if err != nil {
		return dto.GetStreetkingHistoryRes{}, err
	}

	return resp, nil
}

func generateSkewedLowCrashPoint() (decimal.Decimal, error) {

	roll, err := secureRandInt(1, 100)
	if err != nil {
		return decimal.Zero, err
	}

	if roll <= 80 {

		return generateDecimalCrashPoint(1, 1.5, 2)
	}

	return generateDecimalCrashPoint(0.51, 0.9, 2)
}

func generateHighCrashPoint() (decimal.Decimal, error) {

	lambda := 0.1
	randVal, err := secureRandFloat()
	if err != nil {
		return decimal.Zero, err
	}

	one := decimal.NewFromInt(1)
	randDec := decimal.NewFromFloat(randVal)
	exponent := one.Sub(randDec).Neg().Div(decimal.NewFromFloat(lambda))

	min := decimal.NewFromInt(2)
	max := decimal.NewFromInt(75)
	crashPoint := min.Add(exponent.Mul(max.Sub(min).Div(decimal.NewFromInt(100))))

	if crashPoint.GreaterThan(max) {
		crashPoint = max
	}
	crashPoint = crashPoint.Round(2)

	return crashPoint, nil
}

func generateDecimalCrashPoint(min, max float64, decimalPlaces int32) (decimal.Decimal, error) {
	if min > max {
		return decimal.Zero, fmt.Errorf("min cannot be greater than max")
	}

	minDec := decimal.NewFromFloat(min)
	maxDec := decimal.NewFromFloat(max)

	rangeDec := maxDec.Sub(minDec)

	randFloat, err := secureRandFloat()
	if err != nil {
		return decimal.Zero, err
	}
	randDec := decimal.NewFromFloat(randFloat)

	result := minDec.Add(randDec.Mul(rangeDec))

	result = result.Round(decimalPlaces)

	return result, nil
}

func (b *bet) GenerateCrashPoint() (decimal.Decimal, error) {
	b.streetkingsLocker.Lock()
	defer b.streetkingsLocker.Unlock()
	roll, err := secureRandInt(1, 100)
	if err != nil {
		b.log.Error("unable to generate secure random roll", zap.Any("err", err.Error()))
		return decimal.Zero, errors.ErrInternalServerError.Wrap(err, "failed to generate random roll")
	}

	var randomCrashPoint decimal.Decimal

	if b.lowerCounter == 0 {
		randomCrashPoint, err = generateSkewedLowCrashPoint()
		if err != nil {
			b.log.Error("unable to generate skewed low crashpoint", zap.Any("err", err.Error()))
			return decimal.Zero, errors.ErrInternalServerError.Wrap(err, err.Error())
		}

		newCounter, err := secureRandInt(1, 6)
		if err != nil {
			b.log.Error("unable to generate new lowerCounter", zap.Any("err", err.Error()))
			return decimal.Zero, errors.ErrInternalServerError.Wrap(err, "failed to reset lowerCounter")
		}

		b.lowerCounter = int(newCounter)

		if randomCrashPoint.LessThanOrEqual(decimal.NewFromInt(1)) {
			randomCrashPoint = decimal.NewFromFloat(1.01)
		}
		return randomCrashPoint, nil
	}

	switch {
	case roll <= 70:
		randomCrashPoint, err = generateSkewedLowCrashPoint()
		if err != nil {
			b.log.Error("unable to generate skewed low crashpoint", zap.Any("err", err.Error()))
			return decimal.Zero, errors.ErrInternalServerError.Wrap(err, err.Error())
		}

	case roll <= 90:
		randomCrashPoint, err = generateDecimalCrashPoint(1, 1.9, 2)
		if err != nil {
			b.log.Error("unable to generate mid-range crashpoint", zap.Any("err", err.Error()))
			return decimal.Zero, errors.ErrInternalServerError.Wrap(err, err.Error())
		}

	default:
		randomCrashPoint, err = generateHighCrashPoint()
		if err != nil {
			b.log.Error("unable to generate high-range crashpoint", zap.Any("err", err.Error()))
			return decimal.Zero, errors.ErrInternalServerError.Wrap(err, err.Error())
		}
	}

	if randomCrashPoint.LessThanOrEqual(decimal.NewFromInt(1)) {
		randomCrashPoint = decimal.NewFromFloat(1.01)
	}

	b.lowerCounter--

	return randomCrashPoint, nil
}

func secureRandInt(min, max int) (int, error) {
	rangeBig := big.NewInt(int64(max - min + 1))
	n, err := rand.Int(rand.Reader, rangeBig)
	if err != nil {
		return 0, err
	}
	return int(n.Int64()) + min, nil
}

func secureRandFloat() (float64, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1<<52))
	if err != nil {
		return 0, err
	}
	return float64(n.Int64()) / float64(1<<52), nil
}
