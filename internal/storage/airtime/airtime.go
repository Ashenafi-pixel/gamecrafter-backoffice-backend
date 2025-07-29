package airtime

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/joshjones612/egyptkingcrash/internal/constant"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/constant/model/db"
	"github.com/joshjones612/egyptkingcrash/internal/constant/persistencedb"
	"github.com/joshjones612/egyptkingcrash/internal/storage"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type airtime struct {
	log *zap.Logger
	db  *persistencedb.PersistenceDB
}

func Init(log *zap.Logger, db *persistencedb.PersistenceDB) storage.Airtime {
	return &airtime{
		log: log,
		db:  db,
	}
}

func (a *airtime) CreateUtility(ctx context.Context, req dto.AirtimeUtility) (dto.AirtimeUtility, error) {
	resp, err := a.db.CreateAirtimeUtiles(ctx, db.CreateAirtimeUtilesParams{
		ID:            int32(req.ID),
		Productname:   req.ProductName,
		Billername:    req.BillerName,
		Amount:        req.Amount,
		Isamountfixed: req.IsAmountFixed,
		Status:        req.Status,
		Timestamp:     req.Timestamp,
		Price:         decimal.NullDecimal{Decimal: req.Price, Valid: true},
	})

	if err != nil {
		a.log.Error(err.Error())
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.AirtimeUtility{}, err
	}
	return dto.AirtimeUtility{
		LocalID:       resp.LocalID,
		ID:            req.ID,
		ProductName:   resp.Productname,
		BillerName:    resp.Billername,
		Amount:        resp.Amount,
		IsAmountFixed: resp.Isamountfixed,
		Status:        resp.Status,
		Timestamp:     resp.Timestamp,
	}, err
}

func (a *airtime) GetAllAirtimeUtilities(ctx context.Context) ([]dto.AirtimeUtility, bool, error) {
	var AirtimeUtilities []dto.AirtimeUtility
	resp, err := a.db.Queries.GetAllUtilities(ctx)

	if err != nil && err.Error() != dto.ErrNoRows {
		a.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.AirtimeUtility{}, false, err
	}

	if err != nil || len(resp) == 0 {
		return []dto.AirtimeUtility{}, false, nil
	}

	for _, au := range resp {
		AirtimeUtilities = append(AirtimeUtilities, dto.AirtimeUtility{
			LocalID:       au.LocalID,
			ID:            int(au.ID),
			ProductName:   au.Productname,
			BillerName:    au.Billername,
			Amount:        au.Amount,
			IsAmountFixed: au.Isamountfixed,
			Status:        au.Status,
			Timestamp:     au.Timestamp,
			Price:         au.Price.Decimal,
		})
	}

	return AirtimeUtilities, true, nil
}

func (a *airtime) GetAvailableAirtime(ctx context.Context, req dto.GetRequest) (dto.GetAirtimeUtilitiesResp, bool, error) {
	var results []dto.AirtimeUtility
	resp, err := a.db.Queries.GetAvailableAirtime(ctx, db.GetAvailableAirtimeParams{
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		a.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetAirtimeUtilitiesResp{}, false, err
	}

	if err != nil || len(resp) == 0 {
		return dto.GetAirtimeUtilitiesResp{}, false, nil
	}

	totalPage := 1
	for i, r := range resp {
		results = append(results, dto.AirtimeUtility{
			LocalID:          r.LocalID,
			ID:               int(r.ID),
			ProductName:      r.Productname,
			BillerName:       r.Billername,
			Amount:           r.Amount,
			IsAmountFixed:    r.Isamountfixed,
			Status:           r.Status,
			Price:            r.Price.Decimal,
			Timestamp:        r.Timestamp,
			TotalRedemptions: r.TotalRedemptions,
			TotalBuckets:     r.TotalBucksSpent,
		})

		if i == 0 {
			totalPage = int(int(r.TotalRows) / req.PerPage)
			if int(r.TotalRows)%req.PerPage != 0 {
				totalPage++
			}
		}
	}

	return dto.GetAirtimeUtilitiesResp{
		Message: constant.SUCCESS,
		Data: dto.GetAirtimeUtilitiesData{
			TotalPages:       totalPage,
			AirtimeUtilities: results,
		},
	}, true, nil

}

func (a *airtime) UpdateAirtimeStatus(ctx context.Context, ID uuid.UUID, status string) (dto.UpdateAirtimeStatusResp, error) {
	resp, err := a.db.Queries.UpdateAirtimeUtilitiesStatus(ctx, db.UpdateAirtimeUtilitiesStatusParams{
		Status:  status,
		LocalID: ID,
	})

	if err != nil {
		a.log.Error(err.Error())
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.UpdateAirtimeStatusResp{}, err
	}

	return dto.UpdateAirtimeStatusResp{
		Message: constant.SUCCESS,
		Data: dto.AirtimeUtility{
			LocalID:       resp.LocalID,
			ID:            int(resp.ID),
			ProductName:   resp.Productname,
			BillerName:    resp.Billername,
			Amount:        resp.Amount,
			IsAmountFixed: resp.Isamountfixed,
			Status:        resp.Status,
			Price:         resp.Price.Decimal,
			Timestamp:     resp.Timestamp,
		},
	}, nil

}

func (a *airtime) GetAirtimeUtilityByLocalID(ctx context.Context, localID uuid.UUID) (dto.AirtimeUtility, bool, error) {
	resp, err := a.db.Queries.GetAirtimeUtilitiesByID(ctx, localID)
	if err != nil && err.Error() != dto.ErrNoRows {
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.AirtimeUtility{}, false, err
	}

	if err != nil {
		return dto.AirtimeUtility{}, false, nil
	}

	return dto.AirtimeUtility{
		LocalID:       resp.LocalID,
		ID:            int(resp.ID),
		ProductName:   resp.Productname,
		BillerName:    resp.Billername,
		Amount:        resp.Amount,
		IsAmountFixed: resp.Isamountfixed,
		Status:        resp.Status,
		Price:         resp.Price.Decimal,
		Timestamp:     resp.Timestamp,
	}, true, nil
}

func (a *airtime) UpdateAirtimeUtilityPrice(ctx context.Context, req dto.UpdateAirtimeUtilityPriceReq) (dto.AirtimeUtility, error) {
	resp, err := a.db.Queries.UpdateAirtimeUtilitiesPrice(ctx, db.UpdateAirtimeUtilitiesPriceParams{
		Price:   decimal.NullDecimal{Decimal: req.Price, Valid: true},
		LocalID: req.LocalID,
	})

	if err != nil {
		a.log.Error(err.Error())
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.AirtimeUtility{}, err
	}

	return dto.AirtimeUtility{
		LocalID:       resp.LocalID,
		ID:            int(resp.ID),
		ProductName:   resp.Productname,
		BillerName:    resp.Billername,
		Amount:        resp.Amount,
		IsAmountFixed: resp.Isamountfixed,
		Status:        resp.Status,
		Price:         resp.Price.Decimal,
		Timestamp:     resp.Timestamp,
	}, nil
}

func (a *airtime) GetActiveAvailableAirtime(ctx context.Context, req dto.GetRequest) (dto.GetAirtimeUtilitiesResp, error) {
	var utilitiesData []dto.AirtimeUtility
	resp, err := a.db.Queries.GetActiveAvailableAirtime(ctx, db.GetActiveAvailableAirtimeParams{
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		a.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetAirtimeUtilitiesResp{}, err
	}
	totalPage := 1
	for i, u := range resp {
		utilitiesData = append(utilitiesData, dto.AirtimeUtility{
			LocalID:       u.LocalID,
			ID:            int(u.ID),
			ProductName:   u.Productname,
			BillerName:    u.Billername,
			Amount:        u.Amount,
			IsAmountFixed: u.Isamountfixed,
			Status:        u.Status,
			Price:         u.Price.Decimal,
			Timestamp:     u.Timestamp,
		})

		if i == 0 {
			totalPage = int(int(u.TotalRows) / req.PerPage)
			if int(u.TotalRows)%req.PerPage != 0 {
				totalPage++
			}
		}
	}

	return dto.GetAirtimeUtilitiesResp{
		Message: constant.SUCCESS,
		Data: dto.GetAirtimeUtilitiesData{
			TotalPages:       totalPage,
			AirtimeUtilities: utilitiesData,
		},
	}, nil
}

func (a *airtime) SaveAirtimeTransactions(ctx context.Context, req dto.AirtimeTransactions) (dto.AirtimeTransactions, error) {
	resp, err := a.db.Queries.SaveAirtimeTransactions(ctx, db.SaveAirtimeTransactionsParams{
		UserID:           req.UserID,
		TransactionID:    req.TransactionID,
		Cashout:          req.Cashout,
		Billername:       req.BillerName,
		Utilitypackageid: int32(req.UtilityPackageId),
		Packagename:      req.PackageName,
		Amount:           req.Amount,
		Status:           req.Status,
		Timestamp:        time.Now(),
	})

	if err != nil {
		a.log.Error(err.Error())
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.AirtimeTransactions{}, err
	}
	return dto.AirtimeTransactions{
		ID:               resp.ID,
		UserID:           resp.UserID,
		TransactionID:    resp.TransactionID,
		Cashout:          resp.Cashout,
		BillerName:       resp.Billername,
		UtilityPackageId: int(resp.Utilitypackageid),
		PackageName:      resp.Packagename,
		Amount:           resp.Amount,
		Status:           resp.Status,
		Timestamp:        resp.Timestamp,
	}, nil
}

func (a *airtime) GetUserAirtimeTransactions(ctx context.Context, req dto.GetRequest, userID uuid.UUID) (dto.GetAirtimeTransactionsResp, error) {
	var transactions []dto.AirtimeTransactions
	resp, err := a.db.Queries.GetUserAitimeTransactions(ctx, db.GetUserAitimeTransactionsParams{
		UserID: userID,
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		a.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetAirtimeTransactionsResp{}, err
	}

	if err != nil {
		return dto.GetAirtimeTransactionsResp{}, nil
	}
	totalPage := 1
	for i, v := range resp {
		transactions = append(transactions, dto.AirtimeTransactions{
			ID:               v.ID,
			UserID:           v.UserID,
			TransactionID:    v.TransactionID,
			Cashout:          v.Cashout,
			BillerName:       v.Billername,
			UtilityPackageId: int(v.Utilitypackageid),
			PackageName:      v.Packagename,
			Amount:           v.Amount,
			Status:           v.Status,
			Timestamp:        v.Timestamp,
		})

		if i == 0 {
			totalPage = int(int(v.TotalRows) / req.PerPage)
			if int(v.TotalRows)%req.PerPage != 0 {
				totalPage++
			}
		}
	}

	return dto.GetAirtimeTransactionsResp{
		Message: constant.SUCCESS,
		Data: dto.GetAirtimeTransactionsRespData{
			TotalPages:   totalPage,
			Transactions: transactions,
		},
	}, nil
}

func (a *airtime) GetAllAirtimeUtilitiesTransactions(ctx context.Context, req dto.GetRequest) (dto.GetAirtimeTransactionsResp, error) {
	var transactions []dto.AirtimeTransactions
	resp, err := a.db.Queries.GetAllAitimeTransactions(ctx, db.GetAllAitimeTransactionsParams{
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		a.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetAirtimeTransactionsResp{}, err
	}

	if err != nil {
		return dto.GetAirtimeTransactionsResp{}, nil
	}
	totalPage := 1
	for i, v := range resp {
		transactions = append(transactions, dto.AirtimeTransactions{
			ID:               v.ID,
			UserID:           v.UserID,
			TransactionID:    v.TransactionID,
			Cashout:          v.Cashout,
			BillerName:       v.Billername,
			UtilityPackageId: int(v.Utilitypackageid),
			PackageName:      v.Packagename,
			Amount:           v.Amount,
			Status:           v.Status,
			Timestamp:        v.Timestamp,
		})

		if i == 0 {
			totalPage = int(int(v.TotalRows) / req.PerPage)
			if int(v.TotalRows)%req.PerPage != 0 {
				totalPage++
			}
		}
	}

	return dto.GetAirtimeTransactionsResp{
		Message: constant.SUCCESS,
		Data: dto.GetAirtimeTransactionsRespData{
			TotalPages:   totalPage,
			Transactions: transactions,
		},
	}, nil
}

func (a *airtime) UpdateAirtimeAmount(ctx context.Context, req dto.UpdateAirtimeAmountReq) (dto.AirtimeUtility, error) {
	resp, err := a.db.Queries.UpdateAirtimeUtilitiesAmount(ctx, db.UpdateAirtimeUtilitiesAmountParams{
		Amount:  req.Amount.String(),
		LocalID: req.LocalID,
	})

	if err != nil {
		a.log.Error(err.Error())
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.AirtimeUtility{}, err
	}

	return dto.AirtimeUtility{
		LocalID:       resp.LocalID,
		ID:            int(resp.ID),
		ProductName:   resp.Productname,
		BillerName:    resp.Billername,
		Amount:        resp.Amount,
		IsAmountFixed: resp.Isamountfixed,
		Status:        resp.Status,
		Price:         resp.Price.Decimal,
		Timestamp:     resp.Timestamp,
	}, nil
}

func (a *airtime) GetAirtimeUtilitiesStats(ctx context.Context) (dto.AirtimeUtilitiesStats, error) {
	resp, err := a.db.Queries.GetAirtimeUtilitiesStats(ctx)
	if err != nil {
		a.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.AirtimeUtilitiesStats{}, err
	}

	return dto.AirtimeUtilitiesStats{
		TotalRedemptions:       resp.TotalRedemptions,
		TotalBucksSpent:        resp.TotalSpendBucks,
		TotalActiveUtilities:   int(resp.ActiveUtilities),
		TotalInactiveUtilities: int(resp.InactiveUtilities),
		TotalUtilities:         int(resp.Total),
	}, nil
}
