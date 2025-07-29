package balance

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/joshjones612/egyptkingcrash/internal/constant"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	customerrors "github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/module"
	"github.com/joshjones612/egyptkingcrash/internal/storage"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type balance struct {
	balanceStorage              storage.Balance
	balanceLogStorage           storage.BalanceLogs
	exchangeStorage             storage.Exchage
	userStorage                 storage.User
	operationalGroupStorage     storage.OperationalGroup
	operationalGroupTypeStorage storage.OperationalGroupType
	log                         *zap.Logger
	locker                      map[uuid.UUID]*sync.Mutex
	lockerMux                   sync.Mutex
}

func Init(balanceStorage storage.Balance,
	blanceLogStorage storage.BalanceLogs,
	exchangeRateStorage storage.Exchage,
	userStorage storage.User,
	opgStorage storage.OperationalGroup,
	opgtStorage storage.OperationalGroupType,
	log *zap.Logger,
	locker map[uuid.UUID]*sync.Mutex) module.Balance {
	return &balance{
		balanceStorage:              balanceStorage,
		log:                         log,
		locker:                      locker,
		operationalGroupStorage:     opgStorage,
		operationalGroupTypeStorage: opgtStorage,
		balanceLogStorage:           blanceLogStorage,
		exchangeStorage:             exchangeRateStorage,
		userStorage:                 userStorage,
	}
}

func (b *balance) getUserLock(userID uuid.UUID) *sync.Mutex {
	b.lockerMux.Lock()
	defer b.lockerMux.Unlock()

	if _, exists := b.locker[userID]; !exists {
		b.locker[userID] = &sync.Mutex{}
	}
	return b.locker[userID]
}

func (b *balance) Update(ctx context.Context, updateBalanceReq dto.UpdateBalanceReq) (dto.UpdateBalanceRes, error) {
	// Validate create balance request

	//get operational group
	operationalGroup, exist, err := b.operationalGroupStorage.GetOperationalGroupByID(ctx, updateBalanceReq.OperationGroupID)
	if err != nil {
		return dto.UpdateBalanceRes{}, err
	}
	if !exist {
		err := fmt.Errorf("operational group not found")
		b.log.Error(err.Error(), zap.Any("updateBalanceReq", updateBalanceReq))
		err = customerrors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UpdateBalanceRes{}, err
	}

	// get operational group type
	operationalGroupType, exist, err := b.operationalGroupTypeStorage.GetOperationalGroupTypeByID(ctx, updateBalanceReq.OperationTypeID)
	if err != nil {
		return dto.UpdateBalanceRes{}, err
	}
	if !exist {
		err := fmt.Errorf("operational group type not found")
		b.log.Error(err.Error(), zap.Any("updateBalanceReq", updateBalanceReq))
		err = customerrors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UpdateBalanceRes{}, err
	}

	switch operationalGroup.Name {
	case constant.DEPOSIT:
		deposited, err := b.Deposite(ctx, updateBalanceReq)
		if err != nil {
			return dto.UpdateBalanceRes{}, err
		}
		deposited.Data.OperationGroup = operationalGroup.Name
		deposited.Data.OperationType = operationalGroupType.Name

		// save balance logs
		b.SaveBalanceLogs(ctx, dto.SaveBalanceLogsReq{
			UpdateReq:            updateBalanceReq,
			UpdateRes:            deposited,
			OperationalGroupID:   operationalGroup.ID,
			OperationalGroupType: operationalGroupType.ID,
		})
		return deposited, nil
	case constant.WITHDRAWAL:
		updatedBalance, err := b.Withdraw(ctx, updateBalanceReq)
		if err != nil {
			return dto.UpdateBalanceRes{}, err
		}
		updatedBalance.Data.OperationGroup = operationalGroup.Name
		updatedBalance.Data.OperationType = operationalGroupType.Name
		b.SaveBalanceLogs(ctx, dto.SaveBalanceLogsReq{
			UpdateReq:            updateBalanceReq,
			UpdateRes:            updatedBalance,
			OperationalGroupID:   operationalGroup.ID,
			OperationalGroupType: operationalGroupType.ID,
		})
		return updatedBalance, nil
	}

	// if the operation not found
	err = fmt.Errorf("unkown operational group")
	b.log.Error(err.Error(), zap.Any("saveReq", updateBalanceReq))
	err = customerrors.ErrInvalidUserInput.Wrap(err, err.Error())
	return dto.UpdateBalanceRes{}, err
}

func (b *balance) Deposite(ctx context.Context, depositeReq dto.UpdateBalanceReq) (dto.UpdateBalanceRes, error) {

	userLock := b.getUserLock(depositeReq.UserID)
	userLock.Lock()
	defer userLock.Unlock()

	// Check if the balance exists or notD
	blnc, exist, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		UserId:   depositeReq.UserID,
		Currency: depositeReq.Currency,
	})
	if err != nil {
		return dto.UpdateBalanceRes{}, err
	}

	// update the balance
	if exist {
		updatedMoney, err := b.UpdateMoney(ctx, blnc, depositeReq)
		if err != nil {
			return dto.UpdateBalanceRes{}, nil
		}
		var newBalance decimal.Decimal
		if depositeReq.Component == constant.REAL_MONEY {
			newBalance = updatedMoney.RealMoney
		} else {
			newBalance = updatedMoney.BonusMoney
		}
		return dto.UpdateBalanceRes{
			Status:  constant.SUCCESS,
			Message: constant.BALANCE_SUCCESS,
			Data: dto.BalanceData{
				UserID:           updatedMoney.UserId,
				Currency:         updatedMoney.Currency,
				UpdatedComponent: depositeReq.Component,
				NewBalance:       newBalance,
			},
		}, nil
	}

	// Create balance
	realMoney := decimal.Zero
	bonusMoney := decimal.Zero
	if depositeReq.Component == constant.REAL_MONEY {
		realMoney = depositeReq.Amount
	} else if depositeReq.Component == constant.BONUS_MONEY {
		bonusMoney = depositeReq.Amount
	}

	createdBalance, err := b.balanceStorage.CreateBalance(ctx, dto.Balance{
		UserId:     depositeReq.UserID,
		Currency:   depositeReq.Currency,
		RealMoney:  realMoney,
		BonusMoney: bonusMoney,
	})
	if err != nil {
		return dto.UpdateBalanceRes{}, err
	}
	return dto.UpdateBalanceRes{
		Status:  constant.SUCCESS,
		Message: constant.BALANCE_SUCCESS,
		Data: dto.BalanceData{
			UserID:           createdBalance.UserId,
			Currency:         createdBalance.Currency,
			UpdatedComponent: depositeReq.Component,
			NewBalance:       depositeReq.Amount,
		},
	}, nil
}

func (b *balance) Withdraw(ctx context.Context, withdrawalReq dto.UpdateBalanceReq) (dto.UpdateBalanceRes, error) {

	userLock := b.getUserLock(withdrawalReq.UserID)
	userLock.Lock()
	defer userLock.Unlock()

	// Check if the balance exists or not
	blnc, exist, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		UserId:   withdrawalReq.UserID,
		Currency: withdrawalReq.Currency,
	})
	if err != nil {
		return dto.UpdateBalanceRes{}, err
	}
	if !exist {
		err := fmt.Errorf("unable to find blance with this currencys")
		b.log.Error(err.Error(), zap.Any("withdrawalReq", withdrawalReq))
		return dto.UpdateBalanceRes{}, err
	}
	if withdrawalReq.Component == constant.REAL_MONEY {
		newBalance := blnc.RealMoney.Sub(withdrawalReq.Amount)
		if newBalance.LessThan(decimal.Zero) {
			err := fmt.Errorf("insufficient amount")
			b.log.Warn(err.Error(), zap.Any("withdrawalReq", withdrawalReq))
			err = customerrors.ErrInvalidUserInput.Wrap(err, err.Error())
			return dto.UpdateBalanceRes{}, err
		}
		blnc.RealMoney = newBalance
		updatedBlance, err := b.balanceStorage.UpdateBalance(ctx, blnc)
		if err != nil {
			return dto.UpdateBalanceRes{}, err
		}
		return dto.UpdateBalanceRes{
			Status:  constant.SUCCESS,
			Message: constant.BALANCE_SUCCESS,
			Data: dto.BalanceData{
				UserID:           updatedBlance.UserId,
				Currency:         updatedBlance.Currency,
				UpdatedComponent: constant.REAL_MONEY,
				NewBalance:       newBalance,
			},
		}, nil

	} else if withdrawalReq.Component == constant.BONUS_MONEY {
		newBalance := blnc.BonusMoney.Sub(withdrawalReq.Amount)
		if newBalance.LessThan(decimal.Zero) {
			err := fmt.Errorf("insufficient amount")
			b.log.Warn(err.Error(), zap.Any("withdrawalReq", withdrawalReq))
			return dto.UpdateBalanceRes{}, err
		}
		blnc.RealMoney = newBalance
		updatedBlance, err := b.balanceStorage.UpdateBalance(ctx, blnc)
		if err != nil {
			return dto.UpdateBalanceRes{}, err
		}
		return dto.UpdateBalanceRes{
			Status:  constant.SUCCESS,
			Message: constant.BALANCE_SUCCESS,
			Data: dto.BalanceData{
				UserID:           updatedBlance.UserId,
				Currency:         updatedBlance.Currency,
				UpdatedComponent: constant.REAL_MONEY,
				NewBalance:       newBalance,
			},
		}, nil
	}

	// unkown component
	err = fmt.Errorf("unkown component")
	b.log.Error(err.Error())
	err = customerrors.ErrInvalidUserInput.Wrap(err, err.Error())
	return dto.UpdateBalanceRes{}, err
}

func (b *balance) UpdateMoney(ctx context.Context, blnc dto.Balance, updateBalanceReq dto.UpdateBalanceReq) (dto.Balance, error) {
	if updateBalanceReq.Component == constant.REAL_MONEY || updateBalanceReq.Component == constant.BONUS_MONEY {
		if updateBalanceReq.Component == constant.REAL_MONEY {
			updateBalanceReq.Amount = updateBalanceReq.Amount.Add(blnc.RealMoney)
		} else {
			updateBalanceReq.Amount = updateBalanceReq.Amount.Add(blnc.BonusMoney)
		}
		updatedMoney, err := b.balanceStorage.UpdateMoney(ctx, updateBalanceReq)
		if err != nil {
			return dto.Balance{}, err
		}
		return updatedMoney, nil
	} else {
		err := fmt.Errorf("unkown component is given")
		b.log.Warn(err.Error(), zap.Any("updateBalanceReq", updateBalanceReq))
		return dto.Balance{}, err
	}

}

func (b *balance) GetBalanceByUserID(ctx context.Context, userID uuid.UUID) ([]dto.Balance, error) {
	return b.balanceStorage.GetBalancesByUserID(ctx, userID)
}

// CreditWallet credits a user's wallet after payment confirmation.
func (b *balance) CreditWallet(ctx context.Context, req dto.CreditWalletReq) (dto.CreditWalletRes, error) {
	// 1. Check if payment_reference is unique (query transaction logs)
	unique, err := b.IsPaymentReferenceUnique(ctx, req.PaymentReference)
	if err != nil {
		return dto.CreditWalletRes{Success: false, Reason: err.Error()}, err
	}
	if !unique {
		err := customerrors.ErrDuplicatePaymentReference.New("payment reference already processed: %s", req.PaymentReference)
		return dto.CreditWalletRes{Success: false, Reason: "Payment reference already processed"}, err
	}

	// 2. Add amount to user's balance
	userLock := b.getUserLock(req.UserID)
	userLock.Lock()
	defer userLock.Unlock()

	blnc, exist, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		UserId:   req.UserID,
		Currency: req.Currency,
	})
	if err != nil {
		return dto.CreditWalletRes{Success: false, Reason: err.Error()}, err
	}
	if !exist {
		// create balance if not exist
		blnc, err = b.balanceStorage.CreateBalance(ctx, dto.Balance{
			UserId:    req.UserID,
			Currency:  req.Currency,
			RealMoney: req.Amount,
			UpdateAt:  time.Now(),
		})
		if err != nil {
			return dto.CreditWalletRes{Success: false, Reason: err.Error()}, err
		}
	} else {
		blnc.RealMoney = blnc.RealMoney.Add(req.Amount)
		blnc, err = b.balanceStorage.UpdateBalance(ctx, blnc)
		if err != nil {
			return dto.CreditWalletRes{Success: false, Reason: err.Error()}, err
		}
	}

	// 3. Log the transaction (with payment_reference, provider, etc.)
	transactionID := req.PaymentReference
	operationalGroupAndType, err := b.CreateOrGetOperationalGroupAndType(ctx, constant.FUND, constant.ADD_FUND)
	if err != nil {
		return dto.CreditWalletRes{Success: false, Reason: err.Error()}, err
	}
	_, err = b.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
		UserID:             req.UserID,
		Component:          constant.REAL_MONEY,
		Currency:           req.Currency,
		Description:        fmt.Sprintf("wallet credit via %s, ref: %s, type: %s", req.Provider, req.PaymentReference, req.TxType),
		ChangeAmount:       req.Amount,
		OperationalGroupID: operationalGroupAndType.OperationalGroupID,
		OperationalTypeID:  operationalGroupAndType.OperationalTypeID,
		BalanceAfterUpdate: &blnc.RealMoney,
		TransactionID:      &transactionID,
	})
	if err != nil {
		return dto.CreditWalletRes{Success: false, Reason: err.Error()}, err
	}

	return dto.CreditWalletRes{
		Success: true,
		Reason:  "Wallet credited successfully",
	}, nil
}

// IsPaymentReferenceUnique checks if a payment reference is unique in the transaction logs.
func (b *balance) IsPaymentReferenceUnique(ctx context.Context, paymentReference string) (bool, error) {
	res, err := b.balanceLogStorage.GetBalanceLogByTransactionID(ctx, paymentReference)
	if err != nil {
		return false, err
	}
	if res.ID == uuid.Nil {
		return true, nil
	}
	return false, nil
}
