package report

import (
	"context"
	"time"

	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/constant/persistencedb"
	"github.com/joshjones612/egyptkingcrash/internal/storage"
	"go.uber.org/zap"
)

type report struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

func Init(db *persistencedb.PersistenceDB, log *zap.Logger) storage.Report {
	return &report{
		db:  db,
		log: log,
	}
}

func (r *report) DailyReport(ctx context.Context, req dto.DailyReportReq) (dto.DailyReportRes, error) {
	var res dto.DailyReportRes

	date, _ := time.Parse("2006-01-02", req.Date)

	playerCounts, err := r.db.GetPlayerCounts(ctx, date)
	if err != nil {
		r.log.Error("failed to get player counts", zap.String("id", date.String()), zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return res, err
	}
	res.TotalPlayers = playerCounts.TotalPlayers
	res.NewPlayers = playerCounts.NewPlayers

	bucksSpent, err := r.db.GetBucksSpent(ctx, date)
	if err != nil {
		r.log.Error("failed to get bucks spent", zap.String("id", date.String()), zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return res, err
	}

	res.BucksSpent = bucksSpent

	// Default the rest to zero for now
	res.BucksEarned = 0
	res.NetBucksFlow = res.BucksEarned - res.BucksSpent
	res.Revenue = dto.RevenueStream{}
	res.Store = dto.StoreTransaction{}
	res.Airtime = dto.AirtimeConversion{}

	return res, nil
}
