package balance

import (
	"context"

	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/platform/utils"
)

func (b *balance) SaveBalanceLogs(ctx context.Context, saveLogsReq dto.SaveBalanceLogsReq) (dto.BalanceLogs, error) {
	transactionID := utils.GenerateTransactionId()
	return b.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
		UserID:             saveLogsReq.UpdateRes.Data.UserID,
		Component:          saveLogsReq.UpdateReq.Component,
		Currency:           saveLogsReq.UpdateRes.Data.Currency,
		Description:        saveLogsReq.UpdateReq.Description,
		ChangeAmount:       saveLogsReq.UpdateReq.Amount,
		OperationalGroupID: saveLogsReq.OperationalGroupID,
		OperationalTypeID:  saveLogsReq.OperationalGroupType,
		BalanceAfterUpdate: &saveLogsReq.UpdateRes.Data.NewBalance,
		TransactionID:      &transactionID,
	})
}
