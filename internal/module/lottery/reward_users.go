package lottery

import (
	"context"
	"fmt"
	"time"

	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/platform/utils"
	"github.com/shopspring/decimal"

	"github.com/tucanbit/internal/constant/errors"
	"go.uber.org/zap"
)

func (l *lottery) RewardWinnersAndSaveLogs(ctx context.Context, event dto.KafkaLotteryEvent) (bool, error) {
	// Process the winners and save logs
	for _, winner := range event.Rewards {
		logEntry := dto.LotteryKafkaLog{
			LotteryID:       event.LotteryID,
			LotteryRewardID: winner.ID,
			DrawNumbers:     winner.DrawedNumbers,
			Prize:           winner.Prize,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			UniqIdentifier:  event.UniqueID,
		}
		_, err := l.lotteryStorage.CreateLotteryLog(ctx, logEntry)
		if err != nil {
			l.log.Error("failed to create lottery log", zap.Error(err))
			return false, errors.ErrInternalServerError.Wrap(err, "failed to create lottery log")
		}
	}

	// Reward the winners
	for _, winner := range event.Winners {
		logEntry := dto.LotteryLog{
			LotteryID:    event.LotteryID,
			UserID:       winner.UserID,
			RewardID:     winner.RewardID,
			WonAmount:    winner.Prize,
			Currency:     winner.PrizeCurrency,
			TicketNumber: winner.TicketNumber,
			Status:       "won",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		_, err := l.lotteryStorage.CreateLotteryWinnersLogs(ctx, logEntry)
		if err != nil {
			l.log.Error("failed to create lottery winners log", zap.Error(err))
			return false, errors.ErrInternalServerError.Wrap(err, "failed to create lottery winners log")
		}

		// Update user balance
		balance, exist, err := l.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
			UserId:   winner.UserID,
			Currency: winner.PrizeCurrency,
		})

		if err != nil {
			l.log.Error("failed to get user balance", zap.Error(err), zap.Any("winner", winner))
			continue
		}

		if !exist {
			l.log.Warn("user balance does not exist", zap.Any("winner", winner))
			continue
		}
		newBalance := balance.RealMoney.Add(winner.Prize.Mul(decimal.NewFromInt(int64(winner.NumberOfTickets))))
		transactionID := utils.GenerateTransactionId()
		_, err = l.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    winner.UserID,
			Currency:  winner.PrizeCurrency,
			Component: constant.REAL_MONEY,
			Amount:    newBalance,
		})

		if err != nil {
			l.log.Error("failed to update user balance", zap.Error(err), zap.Any("winner", winner))
			continue
		}

		// Save balance log
		operationalGroupAndTypeIDs, err := l.CreateOrGetOperationalGroupAndType(ctx, constant.TRANSFER, constant.BET_LOTTERY_CASHOUT)
		if err != nil {
			l.log.Error("failed to create or get operational group and type", zap.Error(err), zap.Any("winner", winner))
			continue
		}
		// save operations logs
		_, err = l.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
			UserID:             winner.UserID,
			Component:          constant.REAL_MONEY,
			Currency:           winner.PrizeCurrency,
			Description:        fmt.Sprintf("lottery won  amount %v, new balance is %v and  currency %s", winner.Prize.Mul(decimal.NewFromInt(int64(winner.NumberOfTickets))), newBalance, winner.PrizeCurrency),
			ChangeAmount:       decimal.NewFromInt(int64(winner.NumberOfTickets)).Mul(winner.Prize),
			OperationalGroupID: operationalGroupAndTypeIDs.OperationalGroupID,
			OperationalTypeID:  operationalGroupAndTypeIDs.OperationalTypeID,
			BalanceAfterUpdate: &newBalance,
			TransactionID:      &transactionID,
		})

		if err != nil {
			l.log.Error("failed to save balance log", zap.Error(err), zap.Any("winner", winner))
			continue
		}

	}

	return true, nil
}

func (b *lottery) CreateOrGetOperationalGroupAndType(ctx context.Context, operationalGroupName, operationalType string) (dto.OperationalGroupAndType, error) {
	// get transfer operational  group and  type if not exist create group transfer and type transfer-internal
	var operationalGroup dto.OperationalGroup
	var exist bool
	var err error
	var operationalGroupTypeID dto.OperationalGroupType
	operationalGroup, exist, err = b.operationalGroupStorage.GetOperationalGroupByName(ctx, constant.TRANSFER)
	if err != nil {
		return dto.OperationalGroupAndType{}, err
	}
	if !exist {
		// create transfer internal group and type
		operationalGroup, err = b.operationalGroupStorage.CreateOperationalGroup(ctx, dto.OperationalGroup{
			Name:        constant.TRANSFER,
			Description: "internal transaction",
			CreatedAt:   time.Now(),
		})
		if err != nil {
			return dto.OperationalGroupAndType{}, err
		}

		// create operation group type
		operationalGroupTypeID, err = b.operationalGroupTypeStorage.CreateOperationalType(ctx, dto.OperationalGroupType{
			GroupID:     operationalGroup.ID,
			Name:        operationalType,
			Description: "internal transactions",
		})
		if err != nil {
			return dto.OperationalGroupAndType{}, err
		}
	}
	// create or get operational group type if operational group exist
	if exist {
		// get operational group type
		operationalGroupTypeID, exist, err = b.operationalGroupTypeStorage.GetOperationalGroupByGroupIDandName(ctx, dto.OperationalGroupType{
			GroupID: operationalGroup.ID,
			Name:    operationalType,
		})
		if err != nil {
			return dto.OperationalGroupAndType{}, err
		}
		if !exist {
			operationalGroupTypeID, err = b.operationalGroupTypeStorage.CreateOperationalType(ctx, dto.OperationalGroupType{
				GroupID: operationalGroup.ID,
				Name:    operationalType,
			})
			if err != nil {
				return dto.OperationalGroupAndType{}, err
			}

		}
	}
	return dto.OperationalGroupAndType{
		OperationalGroupID: operationalGroup.ID,
		OperationalTypeID:  operationalGroupTypeID.ID,
	}, nil
}
