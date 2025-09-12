package balance

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/platform/utils"
	"go.uber.org/zap"
)

func (b *balance) AddManualFunds(ctx context.Context, fund dto.ManualFundReq) (dto.ManualFundRes, error) {

	var err error
	userLock := b.getUserLock(fund.UserID)
	userLock.Lock()
	defer userLock.Unlock()
	// validate inputs

	if err := b.ValidateFundReq(ctx, fund); err != nil {
		return dto.ManualFundRes{}, err
	}

	//fund user with the specified amount
	//create or get operational group and type
	operationalGroupAndType, err := b.CreateOrGetOperationalGroupAndType(ctx, constant.FUND, constant.ADD_FUND)
	if err != nil {
		return dto.ManualFundRes{}, err
	}
	usrAmount, exist, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		UserId:       fund.UserID,
		CurrencyCode: fund.Currency,
	})
	if err != nil {
		return dto.ManualFundRes{}, err
	}

	if !exist {
		_, err = b.balanceStorage.CreateBalance(ctx, dto.Balance{
			UserId:       fund.UserID,
			CurrencyCode: fund.Currency,
			AmountUnits:  fund.Amount,
		})
		if err != nil {
			return dto.ManualFundRes{}, err
		}
	} else {
		// update existing balance
		_, err = b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    fund.UserID,
			Currency:  fund.Currency,
			Component: constant.REAL_MONEY,
			Amount:    usrAmount.AmountUnits.Add(fund.Amount),
		})
		if err != nil {
			return dto.ManualFundRes{}, err
		}
	}
	//save transaction_log
	blog, err := b.SaveBalanceLogs(ctx, dto.SaveBalanceLogsReq{
		OperationalGroupID:   operationalGroupAndType.OperationalGroupID,
		OperationalGroupType: operationalGroupAndType.OperationalTypeID,
		UpdateReq: dto.UpdateBalanceReq{
			Component:   constant.REAL_MONEY,
			Description: fund.Reason,
			Amount:      fund.Amount,
		},
		UpdateRes: dto.UpdateBalanceRes{
			Data: dto.BalanceData{
				UserID:     fund.UserID,
				Currency:   fund.Currency,
				NewBalance: fund.Amount.Add(usrAmount.AmountUnits),
			},
		},
	})
	if err != nil {
		// reverse amount
		b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    fund.UserID,
			Currency:  fund.Currency,
			Component: constant.REAL_MONEY,
			Amount:    usrAmount.AmountUnits,
		})
	}

	// save it to manual fund
	manualFund, err := b.balanceStorage.SaveManualFunds(ctx, dto.ManualFundReq{
		UserID:        fund.UserID,
		AdminID:       fund.AdminID,
		TransactionID: *blog.TransactionID,
		Type:          constant.ADD_FUND,
		Amount:        fund.Amount,
		Reason:        fund.Reason,
		Currency:      fund.Currency,
		Note:          fund.Reason,
	})
	if err != nil {
		// reverse transaction
		b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    fund.UserID,
			Currency:  fund.Currency,
			Component: constant.REAL_MONEY,
			Amount:    usrAmount.AmountUnits,
		})

		//remove from balance log
		b.balanceLogStorage.DeleteBalanceLog(ctx, blog.ID)
		return dto.ManualFundRes{}, err
	}
	return manualFund, nil
}

func (b *balance) RemoveFundManualy(ctx context.Context, fund dto.ManualFundReq) (dto.ManualFundRes, error) {

	var err error
	userLock := b.getUserLock(fund.UserID)
	userLock.Lock()
	defer userLock.Unlock()
	// validate inputs
	if err := b.ValidateFundReq(ctx, fund); err != nil {
		return dto.ManualFundRes{}, err
	}
	//fund user with the specified amount
	//create or get operational group and type
	operationalGroupAndType, err := b.CreateOrGetOperationalGroupAndType(ctx, constant.FUND, constant.REMOVE_FUND)
	if err != nil {
		return dto.ManualFundRes{}, err
	}
	usrAmount, exist, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		UserId:       fund.UserID,
		CurrencyCode: fund.Currency,
	})
	if err != nil {
		return dto.ManualFundRes{}, err
	}

	if !exist {
		err = fmt.Errorf("user dose not have balance with this currency")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.ManualFundRes{}, err
	} else {
		// update existing balance
		//check if the balance is greater than or equal
		if fund.Amount.GreaterThan(usrAmount.AmountUnits) {
			err = fmt.Errorf("user dose not have enough balance with this currency to substract. user current balance %v", usrAmount.AmountUnits)
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			return dto.ManualFundRes{}, err
		}
		_, err = b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    fund.UserID,
			Currency:  fund.Currency,
			Component: constant.REAL_MONEY,
			Amount:    usrAmount.AmountUnits.Sub(fund.Amount),
		})
		if err != nil {
			return dto.ManualFundRes{}, err
		}
	}
	//save transaction_log
	blog, err := b.SaveBalanceLogs(ctx, dto.SaveBalanceLogsReq{
		OperationalGroupID:   operationalGroupAndType.OperationalGroupID,
		OperationalGroupType: operationalGroupAndType.OperationalTypeID,
		UpdateReq: dto.UpdateBalanceReq{
			Component:   constant.REAL_MONEY,
			Description: fund.Reason,
			Amount:      fund.Amount,
		},
		UpdateRes: dto.UpdateBalanceRes{
			Data: dto.BalanceData{
				UserID:     fund.UserID,
				Currency:   fund.Currency,
				NewBalance: fund.Amount.Add(usrAmount.AmountUnits),
			},
		},
	})
	if err != nil {
		// reverse amount
		b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    fund.UserID,
			Currency:  fund.Currency,
			Component: constant.REAL_MONEY,
			Amount:    usrAmount.AmountUnits,
		})
	}

	// save it to manual fund
	manualFund, err := b.balanceStorage.SaveManualFunds(ctx, dto.ManualFundReq{
		UserID:        fund.UserID,
		AdminID:       fund.AdminID,
		TransactionID: *blog.TransactionID,
		Type:          constant.REMOVE_FUND,
		Amount:        fund.Amount,
		Reason:        fund.Reason,
		Currency:      fund.Currency,
		Note:          fund.Note,
	})
	if err != nil {
		// reverse transaction
		b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    fund.UserID,
			Currency:  fund.Currency,
			Component: constant.REAL_MONEY,
			Amount:    usrAmount.AmountUnits,
		})

		//remove from balance log
		b.balanceLogStorage.DeleteBalanceLog(ctx, blog.ID)
		return dto.ManualFundRes{}, err
	}
	return manualFund, nil
}
func (b *balance) CreateOrGetOperationalGroupAndType(ctx context.Context, operationalGroupName, operationalType string) (dto.OperationalGroupAndType, error) {
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

func (b *balance) ValidateFundReq(ctx context.Context, fund dto.ManualFundReq) error {
	var err error
	if fund.Amount.LessThanOrEqual(decimal.Zero) {
		err = fmt.Errorf("invalid amount is given. fund can not be less than zero")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return err
	}
	//validate currency
	if yes := dto.IsValidCurrency(fund.Currency); !yes {
		err = fmt.Errorf("invalid currency is given.")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return err
	}
	// validate and verify user
	if fund.UserID == uuid.Nil {
		err = fmt.Errorf("invalid user_id is given. ")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
	}

	if fund.Reason == "" {
		err = fmt.Errorf("reason can not be empty")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return err
	}
	_, exist, err := b.userStorage.GetUserByID(ctx, fund.UserID)
	if err != nil {
		return err
	}
	if !exist {
		err = fmt.Errorf("user not found ")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return err
	}

	//validate admin
	_, exist, err = b.userStorage.GetUserByID(ctx, fund.AdminID)
	if err != nil {
		return err
	}
	if !exist {
		err = fmt.Errorf("unable to find admin with this id")
		b.log.Error(err.Error(), zap.Any("adminID", fund.AdminID.String()))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return err
	}
	return nil
}

func (b *balance) GetManualFundLogs(ctx context.Context, filter dto.GetManualFundReq) (dto.GetManualFundRes, error) {
	offset := (filter.Page - 1) * filter.PerPage
	filter.Page = offset
	if filter.Sort.Amount != "" {
		if err := utils.ValidateSortOptions("amount", filter.Sort.Amount); err != nil {
			return dto.GetManualFundRes{}, err
		}
	}

	if filter.Sort.AdminEmail != "" {
		if err := utils.ValidateSortOptions("admin amount", filter.Sort.AdminEmail); err != nil {
			return dto.GetManualFundRes{}, err
		}
	}

	if filter.Sort.Date != "" {
		if err := utils.ValidateSortOptions("date", filter.Sort.Date); err != nil {
			return dto.GetManualFundRes{}, err
		}
	}

	return b.balanceStorage.GetManualFundLogs(ctx, filter)
}
