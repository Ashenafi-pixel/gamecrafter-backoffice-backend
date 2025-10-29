package balance

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/platform/utils"
	"go.uber.org/zap"
)

func (b *balance) Exchange(ctx context.Context, exchngeReq dto.ExchangeBalanceReq) (dto.ExchangeBalanceRes, error) {
	//lock account to write
	var operationalGroupTypeID dto.OperationalGroupType
	var exist bool
	var userToBalance dto.Balance
	var operationalGroup dto.OperationalGroup
	var err error
	userLock := b.getUserLock(exchngeReq.UserID)
	userLock.Lock()
	defer userLock.Unlock()

	// validate exchange request
	if valid := dto.IsValidCurrency(exchngeReq.CurrencyFrom); !valid {
		err := fmt.Errorf("invalid from_currency is given")
		b.log.Warn(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.ExchangeBalanceRes{}, err
	}
	if valid := dto.IsValidCurrency(exchngeReq.CurrencyTo); !valid {
		err := fmt.Errorf("invalid to_currency is given")
		b.log.Warn(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.ExchangeBalanceRes{}, err
	}
	// validate if the amount is not negative
	if exchngeReq.Amount.LessThanOrEqual(decimal.Zero) {
		err := fmt.Errorf("amount can not be negative or zero")
		b.log.Error(err.Error(), zap.Any("exhcnageReq", exchngeReq))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.ExchangeBalanceRes{}, err
	}
	// get transfer operational  group and  type if not exist create group transfer and type transfer-internal
	operationalGroup, exist, err = b.operationalGroupStorage.GetOperationalGroupByName(ctx, constant.TRANSFER)
	if err != nil {
		return dto.ExchangeBalanceRes{}, err
	}
	if !exist {
		// create transfer internal group and type
		operationalGroup, err = b.operationalGroupStorage.CreateOperationalGroup(ctx, dto.OperationalGroup{
			Name:        constant.TRANSFER,
			Description: "internal transaction",
			CreatedAt:   time.Now(),
		})
		if err != nil {
			return dto.ExchangeBalanceRes{}, err
		}

		// create operation group type
		operationalGroupTypeID, err = b.operationalGroupTypeStorage.CreateOperationalType(ctx, dto.OperationalGroupType{
			GroupID:     operationalGroup.ID,
			Name:        constant.INTERNAL_TRANSACTION,
			Description: "internal transactions",
		})
		if err != nil {
			return dto.ExchangeBalanceRes{}, err
		}
	}
	// create or get operational group type if operational group exist
	if exist {
		// get operational group type
		operationalGroupTypeID, exist, err = b.operationalGroupTypeStorage.GetOperationalGroupByGroupIDandName(ctx, dto.OperationalGroupType{
			GroupID: operationalGroup.ID,
			Name:    constant.INTERNAL_TRANSACTION,
		})
		if err != nil {
			return dto.ExchangeBalanceRes{}, err
		}
		if !exist {
			operationalGroupTypeID, err = b.operationalGroupTypeStorage.CreateOperationalType(ctx, dto.OperationalGroupType{
				GroupID: operationalGroup.ID,
				Name:    constant.INTERNAL_TRANSACTION,
			})
			if err != nil {
				return dto.ExchangeBalanceRes{}, err
			}

		}
	}
	// check balance
	userFromBalance, exist, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		UserId:       exchngeReq.UserID,
		CurrencyCode: exchngeReq.CurrencyFrom,
	})
	if err != nil {
		return dto.ExchangeBalanceRes{}, err
	}
	if !exist {
		err := fmt.Errorf("user dose not have balance with %s currency ", exchngeReq.CurrencyFrom)
		b.log.Warn(err.Error(), zap.Any("exchngeReq", exchngeReq))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.ExchangeBalanceRes{}, err
	}

	// check if the user has enough balance
	if userFromBalance.AmountUnits.LessThan(exchngeReq.Amount) {
		err := fmt.Errorf("insufficient amount")
		b.log.Warn(err.Error(), zap.Any("exchngeReq", exchngeReq))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.ExchangeBalanceRes{}, err
	}

	// check if to_currency exist if not create balance with that currency
	// check balance
	userToBalance, exist, err = b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		UserId:       exchngeReq.UserID,
		CurrencyCode: exchngeReq.CurrencyTo,
	})
	if err != nil {
		return dto.ExchangeBalanceRes{}, err
	}
	if !exist {
		// create balance with current currency
		userToBalance, err = b.balanceStorage.CreateBalance(ctx, dto.Balance{
			UserId:        exchngeReq.UserID,
			CurrencyCode:  exchngeReq.CurrencyTo,
			AmountUnits:   decimal.Zero,
			ReservedUnits: decimal.Zero,
		})
		if err != nil {
			return dto.ExchangeBalanceRes{}, err
		}
	}

	// get ConversionRate From to To currency
	conversionRate, exist, err := b.exchangeStorage.GetExchange(ctx, dto.ExchangeReq{
		CurrencyFrom: exchngeReq.CurrencyFrom,
		CurrencyTo:   exchngeReq.CurrencyTo,
	})
	if err != nil {
		return dto.ExchangeBalanceRes{}, err
	}
	if !exist {
		err := fmt.Errorf("unable to get conversion for %s to %s", exchngeReq.CurrencyFrom, exchngeReq.CurrencyTo)
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.ExchangeBalanceRes{}, err
	}

	// convert currency
	exchangedCurrency := exchngeReq.Amount.Mul(conversionRate.Rate)

	// save changed currency
	userToBalanceTemp := userToBalance.AmountUnits
	userToBalance.AmountUnits = userToBalance.AmountUnits.Add(exchangedCurrency)
	exchangedBalance, err := b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
		UserID:    exchngeReq.UserID,
		Currency:  userToBalance.CurrencyCode,
		Component: constant.REAL_MONEY,
		Amount:    userToBalance.AmountUnits,
	})
	if err != nil {
		return dto.ExchangeBalanceRes{}, err
	}
	// substract currency_from account
	userFromBalance.AmountUnits = userFromBalance.AmountUnits.Sub(exchngeReq.Amount)
	balanceAfterSubstracted, err := b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
		UserID:    exchngeReq.UserID,
		Currency:  userFromBalance.CurrencyCode,
		Component: constant.REAL_MONEY,
		Amount:    userFromBalance.AmountUnits,
	})
	if err != nil {
		// reverse user balance to the previous balance
		userToBalance.AmountUnits = userToBalanceTemp
		_, err = b.balanceStorage.UpdateBalance(ctx, userToBalance)
		if err != nil {
			err = fmt.Errorf("unable to revert user exchange %s", err.Error())
			b.log.Error(err.Error())
			err = errors.ErrInternalServerError.Wrap(err, err.Error())
			return dto.ExchangeBalanceRes{}, err
		}
	}
	// save operations logs
	transactionID := utils.GenerateTransactionId()
	_, err = b.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
		UserID:             exchngeReq.UserID,
		Component:          constant.REAL_MONEY,
		Currency:           exchngeReq.CurrencyTo,
		Description:        fmt.Sprintf("transfer from %s to %s amount %v new  %s currency balance is  %v", exchngeReq.CurrencyFrom, exchngeReq.CurrencyTo, exchngeReq.Amount, exchngeReq.CurrencyTo, exchangedCurrency),
		ChangeAmount:       exchngeReq.Amount,
		OperationalGroupID: operationalGroup.ID,
		OperationalTypeID:  operationalGroupTypeID.ID,
		TransactionID:      &transactionID,
		BalanceAfterUpdate: &userToBalance.AmountUnits,
		Status:             constant.COMPLTE,
	})
	if err != nil {
		return dto.ExchangeBalanceRes{}, err
	}
	// transfer internal

	return dto.ExchangeBalanceRes{
		Status:  constant.SUCCESS,
		Message: "balance exchanged successfully",
		Date: dto.ExchangeBalanceResData{
			NewFromBalance: dto.NewBalance{
				Currency: balanceAfterSubstracted.CurrencyCode,
				Balance:  balanceAfterSubstracted.AmountUnits,
			},
			NewToBalance: dto.NewBalance{
				Currency: exchangedBalance.CurrencyCode,
				Balance:  exchangedBalance.AmountUnits,
			},
		},
	}, nil
}
