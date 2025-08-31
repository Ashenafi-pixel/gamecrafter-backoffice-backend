package performance

import (
	"context"

	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/storage"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type performance struct {
	log *zap.Logger
	db  *persistencedb.PersistenceDB
}

func Init(db *persistencedb.PersistenceDB, log *zap.Logger) storage.Performance {
	return &performance{
		log: log,
		db:  db,
	}
}

func (p *performance) GetFinancialMatrix(ctx context.Context) ([]dto.FinancialMatrix, error) {
	var financialMatrix []dto.FinancialMatrix
	resp, err := p.db.Queries.GetFinancialMatrix(ctx)
	if err != nil && err.Error() != dto.ErrNoRows {
		p.log.Error(err.Error())
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return []dto.FinancialMatrix{}, err
	}
	if err != nil {
		return []dto.FinancialMatrix{}, nil
	}
	for _, matrix := range resp {
		totalDepositsWithdrawals := matrix.TotalNumberOfDeposit.Add(matrix.TotalNumberOfWithdrawal)
		if totalDepositsWithdrawals.Cmp(decimal.Zero) == 0 {
			totalDepositsWithdrawals = decimal.NewFromInt(1)
		}

		avg := matrix.TotalDepositAmount.Add(matrix.TotalWithdrawalAmount).
			Div(totalDepositsWithdrawals)

		financialMatrix = append(financialMatrix, dto.FinancialMatrix{
			Currency:                 matrix.Currency.String,
			TotalDepositAmount:       matrix.TotalDepositAmount,
			TotalWithdrawalAmount:    matrix.TotalWithdrawalAmount,
			NumberOfDeposites:        matrix.TotalNumberOfDeposit.IntPart(),
			NumberOfWithdrawals:      matrix.TotalNumberOfWithdrawal.IntPart(),
			AverageTransactionValues: avg,
		})
	}
	return financialMatrix, nil
}

func (p *performance) GetGameMatrics(ctx context.Context) (dto.GameMatricsRes, error) {

	gm, err := p.db.Queries.GetGameMatrics(ctx)
	if err != nil && err.Error() != dto.ErrNoRows {
		p.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GameMatricsRes{}, err
	}
	if err != nil {
		return dto.GameMatricsRes{}, nil
	}
	return dto.GameMatricsRes{
		TotalBets:      gm.TotalBets,
		TotalBetAmount: gm.TotalBetAmount,
		GGR:            gm.TotalBetAmount.Sub(gm.TotalPayout),
		BettingPatterns: dto.GamePattern{
			TotalWins:        int(gm.TotalWins),
			TotalLosses:      int(gm.TotalLosses),
			AvgBetAmount:     gm.AverageBetAmount,
			AvgPayout:        gm.AvgPayout,
			WinPercentage:    gm.WinPercentage,
			LossPercentage:   gm.LossPercentage,
			HighestBetAmount: gm.HighestBetAmount,
			LowestBetAmount:  gm.LowestBetAmount,
			MaxMultiplier:    gm.MaxMultiplier,
			MinMultiplier:    gm.MinMultiplier,
			TotalPayouts:     gm.TotalPayout,
		},
	}, nil
}
