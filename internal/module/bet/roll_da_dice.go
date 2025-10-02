package bet

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
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

func (b *bet) CreateRollDaDice(ctx context.Context, req dto.CreateRollDaDiceReq) (dto.CreateRollDaDiceResp, error) {
	defer b.TriggerLevelResponse(ctx, req.UserID)
	defer b.TriggerPlayerProgressBar(ctx, req.UserID)
	defer b.InitiateTriggerSquadsProgressBar(ctx, req.UserID)
	defer b.userWS.TriggerBalanceWS(ctx, req.UserID)

	var won bool
	var wonAmount decimal.Decimal
	var wonStatus string
	wonStatus = constant.LOSE
	wonAmount = decimal.Zero

	if req.UserGuessStartPoint.GreaterThanOrEqual(req.UserGuessEndPoint) {
		err := fmt.Errorf("endpoint of user guess can not be less than or equal to  startpoint")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CreateRollDaDiceResp{}, err
	}

	if req.UserGuessStartPoint.LessThan(decimal.Zero) || req.UserGuessEndPoint.LessThanOrEqual(decimal.Zero) || req.UserGuessEndPoint.GreaterThan(decimal.NewFromInt(100)) {
		err := fmt.Errorf("startpoint of user guess can not be less than zero and endpoint can not be less than or equal to zero and endpoint can not greater than 100")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CreateRollDaDiceResp{}, err
	}

	if req.BetAmount.LessThanOrEqual(decimal.Zero) {
		err := fmt.Errorf("bet amount can not be less than or equal to zero")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CreateRollDaDiceResp{}, err
	}

	// validate connections is available or not
	conns, ok := b.userSingleGameConnection[req.UserID]
	if !ok {
		err := fmt.Errorf("no active WS connection available to stream multipliers for user please use /ws/single/player endpoint to connect ")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CreateRollDaDiceResp{}, err
	}

	// check balance
	balance, exist, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		UserId:       req.UserID,
		CurrencyCode: constant.POINT_CURRENCY,
	})

	if err != nil {
		return dto.CreateRollDaDiceResp{}, err
	}

	if !exist || balance.RealMoney.LessThan(req.BetAmount) {
		err := fmt.Errorf("insufficient balance ")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CreateRollDaDiceResp{}, err
	}

	// create random point based on the % of possibility
	//multiplier y=−0.0194x+2.91
	multiplier := decimal.NewFromFloat(2.91).Sub((req.UserGuessEndPoint.Sub(req.UserGuessStartPoint)).Mul(decimal.NewFromFloat(0.0194)))
	possibility := req.UserGuessEndPoint.Sub(req.UserGuessStartPoint)
	crashPoint := b.guessCrashPointBasedOnRangeAndPossibilityRange(ctx, req.UserGuessStartPoint, req.UserGuessEndPoint, possibility)

	// wonOrLoss
	req.Multiplier = multiplier
	req.CrashPoint = crashPoint
	if crashPoint.LessThanOrEqual(req.UserGuessEndPoint) && crashPoint.GreaterThanOrEqual(req.UserGuessStartPoint) {
		won = true
		wonAmount = req.BetAmount.Mul(multiplier)
		wonStatus = constant.WON
	}
	req.WonStatus = wonStatus
	req.WonAmount = wonAmount

	resp, err := b.betStorage.CreateRollDaDice(ctx, req)
	if err != nil {
		return dto.CreateRollDaDiceResp{}, err
	}

	// do transaction
	//update user balance
	newBalance := balance.RealMoney.Sub(req.BetAmount)
	transactionID := utils.GenerateTransactionId()
	_, err = b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
		UserID:    req.UserID,
		Currency:  constant.POINT_CURRENCY,
		Component: constant.REAL_MONEY,
		Amount:    newBalance,
	})
	if err != nil {
		return dto.CreateRollDaDiceResp{}, err
	}

	// save transaction logs
	operationalGroupAndTypeIDs, err := b.CreateOrGetOperationalGroupAndType(ctx, constant.TRANSFER, constant.BET_OPERATIONAL_TYPE)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("placeBetReq", req))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    req.UserID,
			Currency:  constant.POINT_CURRENCY,
			Component: constant.REAL_MONEY,
			Amount:    balance.RealMoney,
		})
		return dto.CreateRollDaDiceResp{}, err
	}

	// save operations logs
	_, err = b.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
		UserID:             req.UserID,
		Component:          constant.REAL_MONEY,
		Currency:           constant.POINT_CURRENCY,
		Description:        fmt.Sprintf("place roll da dice bet amount %v, new balance is %v and  currency %s", req.BetAmount, balance.RealMoney.Sub(req.BetAmount), constant.POINT_CURRENCY),
		ChangeAmount:       req.BetAmount,
		OperationalGroupID: operationalGroupAndTypeIDs.OperationalGroupID,
		OperationalTypeID:  operationalGroupAndTypeIDs.OperationalTypeID,
		BalanceAfterUpdate: &newBalance,
		TransactionID:      &transactionID,
	})

	if err != nil {
		return dto.CreateRollDaDiceResp{}, err
	}

	if won {
		// update user balance
		balance, _, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
			UserId:       req.UserID,
			CurrencyCode: constant.POINT_CURRENCY,
		})

		if err != nil {
			return dto.CreateRollDaDiceResp{}, err
		}
		balanceAfterWin := balance.RealMoney.Add(wonAmount)
		updatedBalance, err := b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    req.UserID,
			Currency:  constant.POINT_CURRENCY,
			Component: constant.REAL_MONEY,
			Amount:    balanceAfterWin,
		})
		if err != nil {
			return dto.CreateRollDaDiceResp{}, err
		}
		b.SaveToSquads(ctx, dto.SquadEarns{
			UserID:   req.UserID,
			Currency: constant.POINT_CURRENCY,
			Earn:     wonAmount,
			GameID:   constant.GAME_ROLL_DA_DICE,
		})
		// save cashout balance logs
		operationalGroupAndTypeIDsResp, err := b.CreateOrGetOperationalGroupAndType(ctx, constant.TRANSFER, constant.BET_CASHOUT)
		if err != nil {
			b.log.Error(err.Error(), zap.Any("cashoutReq", req))
			err = errors.ErrInternalServerError.Wrap(err, err.Error())
			// reverse balance
			_, err := b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
				UserID:    req.UserID,
				Currency:  constant.POINT_CURRENCY,
				Component: constant.REAL_MONEY,
				Amount:    balance.RealMoney,
			})
			// reverse cashout

			return dto.CreateRollDaDiceResp{}, err
		}
		// save balance log
		transactionId := utils.GenerateTransactionId()

		b.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
			UserID:             req.UserID,
			Component:          constant.REAL_MONEY,
			Currency:           constant.POINT_CURRENCY,
			Description:        fmt.Sprintf("cash out roll da dice bet  %v  amount, new balance is %v s currency balance is  %s", wonAmount, updatedBalance.RealMoney, constant.POINT_CURRENCY),
			ChangeAmount:       updatedBalance.RealMoney,
			OperationalGroupID: operationalGroupAndTypeIDsResp.OperationalGroupID,
			OperationalTypeID:  operationalGroupAndTypeIDsResp.OperationalTypeID,
			BalanceAfterUpdate: &balanceAfterWin,
			TransactionID:      &transactionId,
		})
	}
	resp.WonAmount = wonAmount
	go b.StreamRollDaDice(ctx, crashPoint, conns, resp)
	return dto.CreateRollDaDiceResp{
		Message: constant.SUCCESS,
		Data: dto.RollDaDiceRespData{
			ID:                  resp.ID,
			UserID:              resp.UserID,
			BetAmount:           resp.BetAmount,
			UserGuessStartPoint: resp.UserGuessStartPoint,
			UserGuessEndPoint:   resp.UserGuessEndPoint,
			Multiplier:          resp.Multiplier,
			Timestamp:           resp.Timestamp,
		},
	}, nil
}

func (b *bet) guessCrashPointBasedOnRangeAndPossibilityRange(ctx context.Context, startPoint, endPoint, possibility decimal.Decimal) decimal.Decimal {
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Calculate win probability as the fraction of the total range (0 to 100)
	totalRange := decimal.NewFromInt(100)
	winProbability := possibility.Div(totalRange)

	// Decide if the crash point falls within the guessed range
	if decimal.NewFromFloat(rand.Float64()).LessThan(winProbability) {
		return startPoint.Add(decimal.NewFromFloat(rand.Float64()).Mul(endPoint.Sub(startPoint)))
	}

	// Generate number outside the specified range
	for {
		num := decimal.NewFromFloat(rand.Float64() * 100.0)
		if num.LessThan(startPoint) || num.GreaterThan(endPoint) {
			return num
		}
	}
}
func (b *bet) StreamRollDaDice(ctx context.Context, crashPoint decimal.Decimal, conns map[*websocket.Conn]bool, resp dto.RollDaDiceData) {

	i := decimal.Zero
	data := dto.StreamRollDaDiceData{
		ID:        resp.ID,
		WonStatus: constant.PENDING,
	}

	for {
		i = decimal.NewFromInt(1).Add(i)
		data.CurrentValue = i
		if i.LessThan(crashPoint) {
			byteData, err := json.Marshal(data)
			if err != nil {
				b.log.Error(err.Error())
				break
			}
			b.StreamToSingleConnection(ctx, conns, byteData, resp.UserID)
			time.Sleep(time.Millisecond * 125)
			continue
		}
		data.WonStatus = resp.WonStatus
		data.WonAmount = resp.WonAmount
		byteData, err := json.Marshal(data)
		if err != nil {
			b.log.Error(err.Error())
			break
		}
		b.StreamToSingleConnection(ctx, conns, byteData, resp.UserID)
		break
	}
}

func (b *bet) GetRollDaDiceHistory(ctx context.Context, req dto.GetRequest, userID uuid.UUID) (dto.GetRollDaDiceResp, error) {
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset
	resp, _, err := b.betStorage.GetUserRollDicBetHistory(ctx, req, userID)
	if err != nil {
		return dto.GetRollDaDiceResp{}, err
	}

	return dto.GetRollDaDiceResp{
		Message: constant.SUCCESS,
		Data:    resp,
	}, nil
}
