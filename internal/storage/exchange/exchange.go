package exchange

import (
	"context"
	"database/sql"

	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/db"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
)

type exchange struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

func Init(db *persistencedb.PersistenceDB, log *zap.Logger) storage.Exchage {
	return &exchange{
		db:  db,
		log: log,
	}
}

func (e *exchange) GetExchange(ctx context.Context, exchangeReq dto.ExchangeReq) (dto.ExchangeRes, bool, error) {
	ex, err := e.db.Queries.GetExchangesFromTo(ctx, db.GetExchangesFromToParams{
		CurrencyFrom: sql.NullString{String: exchangeReq.CurrencyFrom, Valid: true},
		CurrencyTo:   sql.NullString{String: exchangeReq.CurrencyTo, Valid: true},
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		e.log.Error(err.Error(), zap.Any("exchangeReq", exchangeReq))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.ExchangeRes{}, false, err
	}
	if err != nil {
		return dto.ExchangeRes{}, false, nil
	}
	return dto.ExchangeRes{
		ID:           ex.ID,
		CurrencyFrom: ex.CurrencyFrom.String,
		CurrencyTo:   ex.CurrencyTo.String,
		Rate:         ex.Rate.Decimal,
		UpdatedAt:    ex.UpdatedAt.Time,
	}, true, nil
}

func (e *exchange) CreateCurrency(ctx context.Context, req dto.Currency) (dto.Currency, error) {
	resp, err := e.db.Queries.CreateCurrency(ctx, req.Name)
	if err != nil {
		e.log.Error(err.Error())
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.Currency{}, err
	}

	return dto.Currency{ID: resp.ID, Name: resp.Name, Status: resp.Status, Timestamp: resp.Timestamp}, nil
}

func (e *exchange) GetAvailableCurrency(ctx context.Context, req dto.GetRequest) (dto.GetCurrencyReq, error) {
	var availableCurrencies []dto.Currency
	resp, err := e.db.Queries.GetAvailableCurrencies(ctx, db.GetAvailableCurrenciesParams{
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		e.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetCurrencyReq{}, err
	}

	if err != nil {
		return dto.GetCurrencyReq{}, nil
	}
	totalPage := 1
	for i, currency := range resp {
		availableCurrencies = append(availableCurrencies, dto.Currency{
			ID:        currency.ID,
			Name:      currency.Name,
			Status:    currency.Status,
			Timestamp: currency.Timestamp,
		})

		if i == 0 {
			totalPage := int(int(currency.TotalRows) / req.PerPage)
			if int(currency.TotalRows)%req.PerPage != 0 {
				totalPage++
			}
		}

	}
	return dto.GetCurrencyReq{
		Message: constant.SUCCESS,
		Data: dto.GetCurrencyReqData{
			TotalPages: totalPage,
			Currencies: availableCurrencies,
		},
	}, nil
}

func (e *exchange) UpdateCurrency(ctx context.Context, req dto.Currency) (dto.Currency, error) {
	resp, err := e.db.Queries.UpdateCurrencyStatus(ctx, db.UpdateCurrencyStatusParams{
		Status: req.Status,
		ID:     req.ID,
	})

	if err != nil {
		e.log.Error(err.Error())
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.Currency{}, err
	}

	return dto.Currency{
		ID:        resp.ID,
		Name:      resp.Name,
		Status:    resp.Status,
		Timestamp: resp.Timestamp,
	}, nil
}

func (e *exchange) GetCurrency(ctx context.Context, req dto.GetRequest) (dto.GetCurrencyReq, error) {
	var availableCurrencies []dto.Currency
	resp, err := e.db.Queries.GetCurrency(ctx, db.GetCurrencyParams{
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		e.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetCurrencyReq{}, err
	}

	if err != nil {
		return dto.GetCurrencyReq{}, nil
	}
	totalPage := 1
	for i, currency := range resp {
		availableCurrencies = append(availableCurrencies, dto.Currency{
			ID:        currency.ID,
			Name:      currency.Name,
			Status:    currency.Status,
			Timestamp: currency.Timestamp,
		})

		if i == 0 {
			totalPage := int(int(currency.TotalRows) / req.PerPage)
			if int(currency.TotalRows)%req.PerPage != 0 {
				totalPage++
			}
		}

	}
	return dto.GetCurrencyReq{
		Message: constant.SUCCESS,
		Data: dto.GetCurrencyReqData{
			TotalPages: totalPage,
			Currencies: availableCurrencies,
		},
	}, nil
}
