package exchange

import (
	"context"
	"fmt"

	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/module"
	"github.com/joshjones612/egyptkingcrash/internal/storage"
	"go.uber.org/zap"
)

type exchange struct {
	exchangeStorage storage.Exchage
	log             *zap.Logger
}

func Init(exchangeStorage storage.Exchage, log *zap.Logger) module.Exchange {
	return &exchange{
		exchangeStorage: exchangeStorage,
		log:             log,
	}
}

func (e *exchange) GetExchange(ctx context.Context, exchangeReq dto.ExchangeReq) (dto.ExchangeRes, error) {
	// validate exchange
	if err := dto.ValidateExchangeRequest(exchangeReq); err != nil {
		e.log.Warn(err.Error(), zap.Any("exchangeReq", exchangeReq))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.ExchangeRes{}, err
	}
	// validate currency exist or not
	if valid := dto.IsValidCurrency(exchangeReq.CurrencyFrom); !valid {
		err := fmt.Errorf("invalid from_currency is given")
		e.log.Warn(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.ExchangeRes{}, err
	}
	if valid := dto.IsValidCurrency(exchangeReq.CurrencyTo); !valid {
		err := fmt.Errorf("invalid to_currency is given")
		e.log.Warn(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.ExchangeRes{}, err
	}
	// get exchange rate
	exR, exist, err := e.exchangeStorage.GetExchange(ctx, exchangeReq)
	if err != nil {
		return dto.ExchangeRes{}, err
	}
	if !exist {
		err := fmt.Errorf("unable to get conversion for %s to %s", exchangeReq.CurrencyFrom, exchangeReq.CurrencyTo)
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.ExchangeRes{}, err
	}
	return exR, nil
}

func (e *exchange) GetAvailableCurrencies(ctx context.Context, req dto.GetRequest) (dto.GetCurrencyReq, error) {
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset
	return e.exchangeStorage.GetAvailableCurrency(ctx, req)
}

func (e *exchange) GetCurrency(ctx context.Context, req dto.GetRequest) (dto.GetCurrencyReq, error) {
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset
	return e.exchangeStorage.GetCurrency(ctx, req)
}

func (e *exchange) CreateCurrency(ctx context.Context, req dto.Currency) (dto.Currency, error) {
	if yes := dto.IsValidCurrency(req.Name); !yes {
		err := fmt.Errorf("invalid currency")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.Currency{}, err
	}

	return e.exchangeStorage.CreateCurrency(ctx, req)
}
