package bet

import (
	"context"
	"fmt"
	"time"

	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/platform/utils"
	"go.uber.org/zap"
)

func (b *bet) GetFailedRounds(ctx context.Context, req dto.GetFailedRoundsReq) (dto.GetFailedRoundsRes, error) {
	//validate req status
	if req.Status == "" || (req.Status != constant.BET_STATUS_FAILED && req.Status != constant.BET_STATUS_COMPLETED) {
		err := fmt.Errorf("required field status  which should be failed or completed")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.GetFailedRoundsRes{}, err
	}

	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset

	if req.Status == constant.BET_STATUS_COMPLETED {
		return b.betStorage.GetAllFailedRounds(ctx, req)
	}

	return b.betStorage.GetNotRefundedFailedRounds(ctx, req)
}

func (b *bet) RefundFailedRounds(ctx context.Context, failedRounds []dto.BetRound) {
	// get failed rounds
	for _, round := range failedRounds {
		// get bets with that round with no cashout
		// updateUsersBalance
		b.betStorage.UpdateRoundStatusByID(ctx, round.ID, constant.ROUND_FAILED)

		usr, exist, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
			UserId:       round.UserID,
			CurrencyCode: round.Currency,
		})

		if err != nil {
			b.log.Error(err.Error())
			continue
		}

		if !exist {
			b.log.Error("user balance not found with user_id", zap.Any("user_id", round.UserID.String()), zap.Any("currency", round.Currency))
			continue
		}
		newBalance := usr.RealMoney.Add(round.Amount)
		b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    round.UserID,
			Currency:  round.Currency,
			Component: constant.REAL_MONEY,
			Amount:    usr.RealMoney.Add(round.Amount),
		})
		operationalGroupAndTypeIDs, err := b.CreateOrGetOperationalGroupAndType(ctx, constant.TRANSFER, constant.REFUND)
		if err != nil {
			b.log.Error(err.Error())
			//reverse fund
			b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
				UserID:    round.UserID,
				Currency:  round.Currency,
				Component: constant.REAL_MONEY,
				Amount:    usr.RealMoney,
			})
			continue
		}

		transactionID := utils.GenerateTransactionId()
		balanceLogs, err := b.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
			UserID:             round.UserID,
			Component:          constant.REAL_MONEY,
			Currency:           round.Currency,
			Description:        fmt.Sprintf("refund bet amount %v, new balance is %v and currency  %s", round.Amount, newBalance, round.Currency),
			ChangeAmount:       round.Amount,
			OperationalGroupID: operationalGroupAndTypeIDs.OperationalGroupID,
			OperationalTypeID:  operationalGroupAndTypeIDs.OperationalTypeID,
			BalanceAfterUpdate: &newBalance,
			TransactionID:      &transactionID,
		})
		if err != nil {
			//reverse fund
			b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
				UserID:    round.UserID,
				Currency:  round.Currency,
				Component: constant.REAL_MONEY,
				Amount:    usr.RealMoney,
			})
			continue
		}

		//save operation to fund logs
		err = b.betStorage.SaveFailedBetsLogaAuto(ctx, dto.SaveFailedBetsLog{
			UserID:        round.UserID,
			RoundID:       round.ID,
			BetID:         round.BetID,
			Status:        constant.COMPLTE,
			Manual:        false,
			TransactionID: balanceLogs.ID,
		})
		if err != nil {
			// reverse delete transaction logs
			b.balanceLogStorage.DeleteBalanceLog(ctx, balanceLogs.ID)
			//substract balance back to the original
			b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
				UserID:    round.UserID,
				Currency:  round.Currency,
				Component: constant.REAL_MONEY,
				Amount:    usr.RealMoney,
			})
			continue
		}

	}

}

func (b *bet) CreateOrGetOperationalGroupAndType(ctx context.Context, operationalGroupName, operationalType string) (dto.OperationalGroupAndType, error) {
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

func (b *bet) ManualRefundFailedRounds(ctx context.Context, req dto.ManualRefundFailedRoundsReq) (dto.ManualRefundFailedRoundsRes, error) {
	// get user by user id
	usrP, exist, err := b.userStorage.GetUserByID(ctx, req.UserID)
	if err != nil {
		return dto.ManualRefundFailedRoundsRes{}, err
	}
	if !exist {
		err := fmt.Errorf("user name found %v ", zap.Any("refund_req", req))
		b.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.ManualRefundFailedRoundsRes{}, err
	}

	//get active bet for the given user
	bt, exist, err := b.betStorage.GetUserActiveBetWithRound(ctx, usrP.ID, req.RoundID)
	if err != nil {
		return dto.ManualRefundFailedRoundsRes{}, err
	}

	if !exist {
		err = fmt.Errorf("unable to find active bet for the given user  userID %v and roundID %v", req.UserID, req.RoundID)
		b.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.ManualRefundFailedRoundsRes{}, err
	}

	usr, exist, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		UserId:       usrP.ID,
		CurrencyCode: bt.Currency,
	})

	if err != nil {
		return dto.ManualRefundFailedRoundsRes{}, err
	}

	if !exist {
		err := fmt.Errorf("unable to find user account with specified currency %v", bt.Currency)
		b.log.Error("user balance not found with user_id")
		return dto.ManualRefundFailedRoundsRes{}, err
	}
	newBalance := usr.RealMoney.Add(bt.Amount)
	b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
		UserID:    bt.UserID,
		Currency:  bt.Currency,
		Component: constant.REAL_MONEY,
		Amount:    usr.RealMoney.Add(bt.Amount),
	})
	operationalGroupAndTypeIDs, err := b.CreateOrGetOperationalGroupAndType(ctx, constant.TRANSFER, constant.REFUND)
	if err != nil {
		b.log.Error(err.Error())
		//reverse fund
		b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    bt.UserID,
			Currency:  bt.Currency,
			Component: constant.REAL_MONEY,
			Amount:    usr.RealMoney,
		})
		return dto.ManualRefundFailedRoundsRes{}, err
	}

	transactionID := utils.GenerateTransactionId()
	balanceLogs, err := b.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
		UserID:             bt.UserID,
		Component:          constant.REAL_MONEY,
		Currency:           bt.Currency,
		Description:        fmt.Sprintf("refund bet amount %v, new balance is %v and currency  %s", bt.Amount, newBalance, bt.Currency),
		ChangeAmount:       bt.Amount,
		OperationalGroupID: operationalGroupAndTypeIDs.OperationalGroupID,
		OperationalTypeID:  operationalGroupAndTypeIDs.OperationalTypeID,
		BalanceAfterUpdate: &newBalance,
		TransactionID:      &transactionID,
	})
	if err != nil {
		//reverse fund
		b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    bt.UserID,
			Currency:  bt.Currency,
			Component: constant.REAL_MONEY,
			Amount:    usr.RealMoney,
		})
		return dto.ManualRefundFailedRoundsRes{}, err
	}

	//save operation to fund logs
	err = b.betStorage.SaveFailedBetsLogaAuto(ctx, dto.SaveFailedBetsLog{
		UserID:        bt.UserID,
		RoundID:       req.RoundID,
		BetID:         bt.ID,
		Status:        constant.COMPLTE,
		TransactionID: balanceLogs.ID,
		Manual:        true,
		AdminID:       req.AdminID,
		CreatedAt:     time.Now(),
	})
	if err != nil {
		// reverse delete transaction logs
		b.balanceLogStorage.DeleteBalanceLog(ctx, balanceLogs.ID)
		//substract balance back to the original
		b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    bt.UserID,
			Currency:  bt.Currency,
			Component: constant.REAL_MONEY,
			Amount:    usr.RealMoney,
		})
		return dto.ManualRefundFailedRoundsRes{}, err
	}
	return dto.ManualRefundFailedRoundsRes{
		Message: constant.SUCCESS,
		Data: dto.ManualRefundFailedRoundsData{
			UserID:        req.UserID,
			BetID:         bt.ID,
			RefundAmount:  bt.Amount,
			TransactionID: transactionID,
		},
	}, nil
}
