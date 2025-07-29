package bet

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/joshjones612/egyptkingcrash/internal/constant"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/platform/utils"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

func (b *bet) CreateRounds() {
	b.streetkingsLocker.Lock()
	defer b.streetkingsLocker.Unlock()
	var exist bool
	var err error
	var inProgessBet dto.BetRound
	var openRound dto.BetRound
	var randomCrashPoint decimal.Decimal
	openRound, exist = b.currentRound[constant.BET_OPEN]
	open, err := b.CheckBetLockStatus(context.Background(), constant.GAME_EGYPTKING)
	if err != nil {
		return
	}

	if !open {
		err := fmt.Errorf("game currently unavailable")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return
	}

	if exist {
		now := time.Now().In(time.Now().Location()).UTC().Add(time.Second)
		createdAt := openRound.CreatedAt

		if createdAt == nil {
			b.log.Warn(fmt.Sprintf("openRound ID %v has nil CreatedAt", openRound.ID))
			return
		}

		createdAtVal := *createdAt
		adjustedCreatedAt := createdAtVal.Add(b.betOpenForDuration)
		if adjustedCreatedAt.Before(now) {
			b.currentRound[constant.BET_INPROGRESS] = openRound
			delete(b.currentRound, constant.BET_OPEN)
			openBet, err := b.betStorage.UpdateRoundStatusByID(context.Background(), openRound.ID, constant.BET_INPROGRESS)
			if err != nil {
				b.log.Warn(fmt.Sprintf("Error updating round status for ID %v: %v", openRound.ID, err))
				return
			}
			b.streaming[openBet.ID] = true
			closedAt := time.Now().In(time.Now().Location()).UTC()
			openBet.ClosedAt = &closedAt
			go b.StreamRound(context.Background(), openBet)
		}

		return
	}

	inProgessBet, exist = b.currentRound[constant.BET_INPROGRESS]

	if exist {

		_, ok := b.streaming[inProgessBet.ID]
		if !ok {
			// bet is not streaming
			b.streaming[inProgessBet.ID] = true
			go b.StreamRound(context.Background(), inProgessBet)
			return
		}

		return
	}

	min := int64(1)
	max := decimal.NewFromInt(5.0)

	// 1% chance for the crash point max
	roll := rand.Intn(100) + 1 // Random roll for probabilities
	if roll <= 1 {
		randomCrashPoint, err = utils.GenerateRandomCrashPoint(min, max.IntPart())
		if err != nil {
			b.log.Error("unable to generate random crashpoints ", zap.Any("err", err.Error()))
			return
		}
		b.lowerCounter = b.lowerCounter - 1
	} else {
		// Else, proceed with weighted probabilities
		if b.lowerCounter == 0 {

			// 90% chance for low range
			if roll <= 90 {
				randomCrashPoint, err = utils.GenerateRandomCrashPoint(min, 2)
			} else if roll <= 98 {
				// 8% chance for mid-range 2.0x - 5.0x
				randomCrashPoint, err = utils.GenerateRandomCrashPoint(2, 5)
			} else {
				// 2% chance for high range 5.0x - 10.0x
				randomCrashPoint, err = utils.GenerateRandomCrashPoint(5, 10)
			}

			if err != nil {
				b.log.Error("unable to generate random crashpoints ", zap.Any("err", err.Error()))
				return
			}

			lower, err := utils.GenerateRandomCrashPoint(min, int64(6))
			if err != nil {
				b.log.Error("unable to generate random crashpoints ", zap.Any("err", err.Error()))
				return
			}
			b.lowerCounter = int(lower.IntPart())

		} else {
			randomCrashPoint, err = utils.GenerateRandomCrashPoint(min, max.IntPart())
			if err != nil {
				b.log.Error("unable to generate random crashpoints ", zap.Any("err", err.Error()))
				return
			}
			b.lowerCounter = b.lowerCounter - 1
		}
	}

	bt, _ := b.betStorage.SaveRounds(context.Background(), dto.BetRound{
		Status:     constant.BET_OPEN,
		CrashPoint: randomCrashPoint,
	})

	b.currentRound[constant.BET_OPEN] = bt

}

func (b *bet) StreamRound(ctx context.Context, betRound dto.BetRound) error {
	i := decimal.NewFromInt(0.0)
	b.InProgressRounds[betRound.ID] = betRound
	defer func() {
		b.log.Info("cleanup for round", zap.Any("roundID", betRound.ID))
		delete(b.InProgressRounds, betRound.ID)
		delete(b.streaming, betRound.ID)
		delete(b.betRoundMultiplayerHolder, betRound.ID)
		b.CreateRounds()
	}()
	var mutex sync.Mutex
	defer b.betStorage.CloseRound(ctx, betRound.ID)
	conn := b.broadcastConn
	for betRound.CrashPoint.GreaterThanOrEqual(i) {

		time.Sleep(time.Millisecond * 110)
		mutex.Lock()
		b.betRoundMultiplayerHolder[betRound.ID] = i
		byteDate, err := json.Marshal(dto.WSRes{
			Type: constant.WS_CURRENT_MULTIPLIER,
			Data: dto.BetRoundResp{
				ID:          betRound.ID,
				Multiplayer: i,
			},
		})
		if err != nil {
			mutex.Unlock()
			b.log.Error(err.Error())
			return err
		}
		mutex.Unlock()
		i = i.Add(decimal.NewFromFloat(0.01))
		b.BroadcastToAllMessage(ctx, byteDate, conn)
	}
	message, err := json.Marshal(dto.WSRes{
		Type: constant.WS_CRASH,
		Data: dto.CrashPointRes{
			Round:   betRound,
			Message: constant.BET_CRASHPOINT_REACHED,
		},
	})
	if err != nil {
		b.log.Error(err.Error(), zap.Any("betRound", betRound))
	}
	b.BroadcastToAllMessage(ctx, message, conn)
	delete(b.currentRound, constant.BET_INPROGRESS)
	b.CreateRounds()
	return nil
}

func (b *bet) GetOpenRound(ctx context.Context) (dto.OpenRoundRes, error) {
	cp := b.currentRound
	if _, ok := cp[constant.BET_OPEN]; !ok {
		return dto.OpenRoundRes{
			Message: constant.BET_NO_OPEN_ROUND_AVAILABLE,
		}, nil
	} else {
		startTime := time.Until(cp[constant.BET_OPEN].CreatedAt.Add(b.betOpenForDuration)) / 1000000000
		if startTime < 0 {
			startTime = 0
		}

		startT := strings.Split(startTime.String(), "n")[0]

		return dto.OpenRoundRes{
			Message:   constant.SUCCESS,
			BetRound:  cp[constant.BET_OPEN].ID,
			StartTime: &startT,
		}, nil
	}

}

func (b *bet) CashOut(ctx context.Context, cashOutReq dto.CashOutReq) (dto.CashOutRes, error) {
	// check if the round id is still in progress or not
	inProgressRound, exist, err := b.betStorage.GetRoundByID(ctx, cashOutReq.RoundID)
	if err != nil {
		return dto.CashOutRes{}, err
	}
	if !exist {
		err = fmt.Errorf("unable to find round with this id")
		b.log.Warn(err.Error(), zap.Any("cashOutReq", cashOutReq))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CashOutRes{}, err
	}
	if inProgressRound.Status != constant.BET_INPROGRESS {
		if inProgressRound.Status == constant.BET_OPEN {
			err = fmt.Errorf("round not start yet")
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			return dto.CashOutRes{}, err
		} else {
			err = fmt.Errorf("round closed")
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			return dto.CashOutRes{}, err
		}
	}
	// check current multiplayer for this round and multiplay with current bet amount
	// get user with this round
	userBet, exist, err := b.betStorage.GetUserBetByUserIDAndRoundID(ctx, dto.Bet{
		UserID:  cashOutReq.UserID,
		RoundID: cashOutReq.RoundID,
		Status:  constant.ACTIVE,
	})
	if err != nil {
		return dto.CashOutRes{}, err
	}
	if !exist || len(userBet) < 1 {
		err = fmt.Errorf("unable to find active user bet")
		b.log.Warn(err.Error(), zap.Any("cashoutReq", cashOutReq))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CashOutRes{}, err
	}

	// get current multiplayer
	currentMultiplayer, ok := b.betRoundMultiplayerHolder[cashOutReq.RoundID]
	if !ok {
		err = fmt.Errorf("round closed")
		b.log.Info(err.Error(), zap.Any("cashoutReq", cashOutReq))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CashOutRes{}, err
	}

	multiplayedBalance := userBet[0].Amount.Mul(currentMultiplayer)
	// update user balance
	//get user balance with that currency
	// lock user before making transactions
	userLock := b.getUserLock(cashOutReq.UserID)
	userLock.Lock()
	defer userLock.Unlock()
	userBalance, _, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		UserId:   cashOutReq.UserID,
		Currency: userBet[0].Currency,
	})
	if err != nil {
		return dto.CashOutRes{}, err
	}
	// lock user account before updating it

	updatedBalance, err := b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
		UserID:    cashOutReq.UserID,
		Currency:  userBet[0].Currency,
		Component: constant.REAL_MONEY,
		Amount:    userBalance.RealMoney.Add(multiplayedBalance),
	})
	if err != nil {
		return dto.CashOutRes{}, err
	}
	// update userBetTable
	savedCashOut, err := b.betStorage.CashOut(ctx, dto.SaveCashoutReq{
		ID:         userBet[0].BetID,
		Multiplier: currentMultiplayer,
		Payout:     multiplayedBalance,
	})

	if err != nil {
		// reverse
		_, err := b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    cashOutReq.UserID,
			Currency:  userBet[0].Currency,
			Component: constant.REAL_MONEY,
			Amount:    userBalance.RealMoney,
		})
		if err != nil {
			return dto.CashOutRes{}, err
		}
	}
	// save cashout balance logs
	operationalGroupAndTypeIDs, err := b.CreateOrGetOperationalGroupAndType(ctx, constant.TRANSFER, constant.BET_CASHOUT)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("cashoutReq", cashOutReq))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		// reverse balance
		_, err := b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    cashOutReq.UserID,
			Currency:  userBet[0].Currency,
			Component: constant.REAL_MONEY,
			Amount:    userBalance.RealMoney,
		})
		// reverse cashout
		b.betStorage.ReverseCashOut(ctx, userBet[0].BetID)
		return dto.CashOutRes{}, err
	}
	// save balance log
	betAfterUpdate := userBalance.RealMoney.Add(multiplayedBalance)
	transactionId := utils.GenerateTransactionId()
	b.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
		UserID:             cashOutReq.UserID,
		Component:          constant.REAL_MONEY,
		Currency:           userBet[0].Currency,
		Description:        fmt.Sprintf("cash out bet  %v  amount, new balance is %v s currency balance is  %s", multiplayedBalance, updatedBalance.RealMoney, userBet[0].Currency),
		ChangeAmount:       updatedBalance.RealMoney,
		OperationalGroupID: operationalGroupAndTypeIDs.OperationalGroupID,
		OperationalTypeID:  operationalGroupAndTypeIDs.OperationalTypeID,
		BalanceAfterUpdate: &betAfterUpdate,
		TransactionID:      &transactionId,
	})
	// send Response to user using websocket
	byteData, err := json.Marshal(dto.WSRes{
		Type: constant.WS_CASHOUT,
		Data: savedCashOut,
	})
	if err != nil {
		b.log.Error(err.Error(), zap.Any("marshalreq", savedCashOut))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.CashOutRes{}, err
	}
	// findUserConnections
	userConnections := make(map[uuid.UUID]map[*websocket.Conn]bool)
	userConnections[cashOutReq.UserID] = b.userConn[cashOutReq.UserID]
	b.BroadcastMessage(ctx, byteData, userConnections)
	return dto.CashOutRes{}, nil
}

func (b *bet) GetLeaders(ctx context.Context) (dto.LeadersResp, error) {
	leadersRes := dto.LeadersResp{}
	leaders, err := b.betStorage.GetLeaders(ctx)
	if err != nil {
		return dto.LeadersResp{}, err
	}
	for _, leader := range leaders.Leaders {
		profilePicture := ""
		if leader.ProfileURL != "" {
			profilePicture = utils.GetFromS3Bucket(b.bucketName, leader.ProfileURL)
		}
		leadersRes.Leaders = append(leadersRes.Leaders, dto.Leader{
			ProfileURL: profilePicture,
			Payout:     leader.Payout,
		})
		leadersRes.TotalPlayers = leaders.TotalPlayers

	}
	return leadersRes, nil
}
