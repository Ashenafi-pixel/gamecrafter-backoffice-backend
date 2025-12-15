package report

import (
	"context"
	"encoding/json"
	"time"

	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/storage"
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

func (r *report) GetDuplicateIPAccounts(ctx context.Context) ([]dto.DuplicateIPAccountsReport, error) {
	// Query to find IP addresses that have created more than one account
	// We use the first session (earliest created_at) for each user as the registration session
	query := `
		WITH first_sessions AS (
			SELECT DISTINCT ON (us.user_id)
				us.user_id,
				us.ip_address,
				us.user_agent,
				us.created_at as session_date,
				u.username,
				u.email,
				u.created_at
			FROM user_sessions us
			INNER JOIN users u ON us.user_id = u.id
			WHERE us.ip_address IS NOT NULL 
				AND us.ip_address != ''
				AND u.is_admin = false
			ORDER BY us.user_id, us.created_at ASC
		),
		ip_counts AS (
			SELECT 
				ip_address,
				COUNT(*) as account_count
			FROM first_sessions
			GROUP BY ip_address
			HAVING COUNT(*) > 1
		)
		SELECT 
			fs.ip_address,
			ic.account_count as count,
			json_agg(
				json_build_object(
					'user_id', fs.user_id::text,
					'username', fs.username,
					'email', fs.email,
					'user_agent', fs.user_agent,
					'created_at', fs.created_at::text,
					'session_date', fs.session_date::text
				) ORDER BY fs.created_at
			) as accounts
		FROM first_sessions fs
		INNER JOIN ip_counts ic ON fs.ip_address = ic.ip_address
		GROUP BY fs.ip_address, ic.account_count
		ORDER BY ic.account_count DESC, fs.ip_address
	`

	rows, err := r.db.GetPool().Query(ctx, query)
	if err != nil {
		r.log.Error("failed to get duplicate IP accounts", zap.Error(err))
		return nil, errors.ErrUnableToGet.Wrap(err, "failed to get duplicate IP accounts")
	}
	defer rows.Close()

	var reports []dto.DuplicateIPAccountsReport
	for rows.Next() {
		var report dto.DuplicateIPAccountsReport
		var accountsJSON []byte

		if err := rows.Scan(&report.IPAddress, &report.Count, &accountsJSON); err != nil {
			r.log.Error("failed to scan duplicate IP accounts row", zap.Error(err))
			continue
		}

		// Parse the JSON array of accounts
		if err := json.Unmarshal(accountsJSON, &report.Accounts); err != nil {
			r.log.Error("failed to unmarshal accounts JSON", zap.Error(err))
			continue
		}

		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("error iterating duplicate IP accounts rows", zap.Error(err))
		return nil, errors.ErrUnableToGet.Wrap(err, "error iterating duplicate IP accounts rows")
	}

	return reports, nil
}

func (r *report) SuspendAccountsByIP(ctx context.Context, ipAddress string) ([]uuid.UUID, error) {
	// Get all user IDs that have the first session from this IP address
	query := `
		WITH first_sessions AS (
			SELECT DISTINCT ON (us.user_id)
				us.user_id
			FROM user_sessions us
			INNER JOIN users u ON us.user_id = u.id
			WHERE us.ip_address = $1
				AND us.ip_address IS NOT NULL 
				AND us.ip_address != ''
				AND u.is_admin = false
			ORDER BY us.user_id, us.created_at ASC
		)
		SELECT user_id FROM first_sessions
	`

	rows, err := r.db.GetPool().Query(ctx, query, ipAddress)
	if err != nil {
		r.log.Error("failed to get user IDs from IP", zap.Error(err), zap.String("ip_address", ipAddress))
		return nil, errors.ErrUnableToGet.Wrap(err, "failed to get user IDs from IP")
	}
	defer rows.Close()

	var userIDs []uuid.UUID
	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			r.log.Error("failed to scan user ID", zap.Error(err))
			continue
		}
		userIDs = append(userIDs, userID)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("error iterating user IDs", zap.Error(err))
		return nil, errors.ErrUnableToGet.Wrap(err, "error iterating user IDs")
	}

	return userIDs, nil
}
