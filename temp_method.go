func (p *PersistenceDB) GetBalanceLogByTransactionID(ctx context.Context, transactionID string) (db.BalanceLog, error) {
	query := `
		SELECT
			bl.id,
			bl.user_id,
			bl.component,
			bl.currency,
			bl.description,
			bl.change_amount,
			bl.operational_group_id,
			ops.name AS type,
			bl.operational_type_id,
			ot.name AS operational_type_name,
			bl.timestamp,
			bl.balance_after_update,
			bl.transaction_id,
			bl.status
		FROM
			balance_logs bl
		JOIN
			operational_groups ops ON ops.id = bl.operational_group_id
		JOIN
			operational_types ot ON ot.id = bl.operational_type_id
		WHERE
			bl.transaction_id = $1`

	var balanceLog db.BalanceLog
	err := p.pool.QueryRow(ctx, query, transactionID).Scan(
		&balanceLog.ID,
		&balanceLog.UserID,
		&balanceLog.Component,
		&balanceLog.Currency,
		&balanceLog.Description,
		&balanceLog.ChangeAmount,
		&balanceLog.OperationalGroupID,
		&balanceLog.Type,
		&balanceLog.OperationalTypeID,
		&balanceLog.OperationalTypeName,
		&balanceLog.Timestamp,
		&balanceLog.BalanceAfterUpdate,
		&balanceLog.TransactionID,
		&balanceLog.Status,
	)
	if err != nil {
		return db.BalanceLog{}, err
	}

	return balanceLog, nil
}