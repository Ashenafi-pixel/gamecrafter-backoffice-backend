package balancelogs

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/module"
	"github.com/joshjones612/egyptkingcrash/internal/storage"
	"github.com/joshjones612/egyptkingcrash/platform/utils"
	"go.uber.org/zap"
)

type balance_logs struct {
	balanceLogStorage storage.BalanceLogs
	log               *zap.Logger
}

var components = []string{"real_money", "bonus_money"}

func Init(balanceLogStorage storage.BalanceLogs, log *zap.Logger) module.BalanceLogs {
	return &balance_logs{
		balanceLogStorage: balanceLogStorage,
		log:               log,
	}
}

func (bl *balance_logs) GetBalanceLogs(ctx context.Context, balanceLogsReq dto.GetBalanceLogReq) (dto.GetBalanceLogRes, error) {
	// Set default values for pagination
	if balanceLogsReq.PerPage <= 0 {
		balanceLogsReq.PerPage = 10
	}
	if balanceLogsReq.Page <= 0 {
		balanceLogsReq.Page = 1
	}

	offset := (balanceLogsReq.Page - 1) * balanceLogsReq.PerPage
	balanceLogsReq.Offset = offset
	// validate if it the value has non component
	if balanceLogsReq.Component != "" {
		validComponent := false
		for _, component := range components {
			if balanceLogsReq.Component == component {
				validComponent = true
				break
			}
		}
		if !validComponent {
			err := fmt.Errorf("invalid component is given")
			bl.log.Error(err.Error(), zap.Any("balanceLogsReq", balanceLogsReq))
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			return dto.GetBalanceLogRes{}, err
		}
	}

	return bl.balanceLogStorage.GetBalanceLog(ctx, balanceLogsReq)
}

func (bl *balance_logs) GetBalanceLogByID(ctx context.Context, id uuid.UUID ) (dto.BalanceLogsRes, error) {
	if id == uuid.Nil {
		err := errors.ErrInvalidUserInput.New("invalid blance logs UUID")
		return dto.BalanceLogsRes{}, err
	}

	balanceLog, err :=  bl.balanceLogStorage.GetBalanceLogByID(ctx, id)
	if err != nil {
		bl.log.Error(err.Error(), zap.Any("balanceLogID", id.String()))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.BalanceLogsRes{}, err
	}

	return balanceLog, nil
}

func (bl *balance_logs) GetBalanceLogsForAdmin(ctx context.Context, req dto.AdminGetBalanceLogsReq) (dto.AdminGetBalanceLogsRes, error) {
	// Set default values for Pagination
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	// check the filters
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset
	// validate sort if exist

	if req.Sort.Amount != "" {
		if err := utils.ValidateSortOptions("amount", req.Sort.Amount); err != nil {
			return dto.AdminGetBalanceLogsRes{}, err
		}
	}

	if req.Sort.Date != "" {
		if err := utils.ValidateSortOptions("date", req.Sort.Date); err != nil {
			return dto.AdminGetBalanceLogsRes{}, err
		}
	}

	return bl.balanceLogStorage.GetBalanceLogsForAdmin(ctx, req)
}
