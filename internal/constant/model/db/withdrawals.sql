-- name: GetAllWithdrawals :many
SELECT 
    w.id, w.user_id, w.admin_id, w.chain_id, w.currency_code, w.protocol,
    w.withdrawal_id, w.usd_amount_cents, w.crypto_amount, w.exchange_rate,
    w.fee_cents, w.source_wallet_address, w.to_address, w.tx_hash,
    w.status, w.requires_admin_review, w.admin_review_deadline,
    w.processed_by_system, w.amount_reserved_cents, w.reservation_released,
    w.reservation_released_at, w.metadata, w.error_message,
    w.created_at, w.updated_at,
    u.username, u.email, u.first_name, u.last_name
FROM withdrawals w
LEFT JOIN users u ON w.user_id = u.id
WHERE ($1::text IS NULL OR w.status = $1)
    AND ($2::uuid IS NULL OR w.user_id = $2)
    AND ($3::text IS NULL OR w.withdrawal_id ILIKE '%' || $3 || '%')
    AND ($4::text IS NULL OR u.username ILIKE '%' || $4 || '%')
    AND ($5::text IS NULL OR u.email ILIKE '%' || $5 || '%')
    AND ($6::timestamp IS NULL OR w.created_at >= $6)
    AND ($7::timestamp IS NULL OR w.created_at <= $7)
ORDER BY w.created_at DESC
LIMIT $8 OFFSET $9;

-- name: GetWithdrawalByID :one
SELECT 
    w.id, w.user_id, w.admin_id, w.chain_id, w.currency_code, w.protocol,
    w.withdrawal_id, w.usd_amount_cents, w.crypto_amount, w.exchange_rate,
    w.fee_cents, w.source_wallet_address, w.to_address, w.tx_hash,
    w.status, w.requires_admin_review, w.admin_review_deadline,
    w.processed_by_system, w.amount_reserved_cents, w.reservation_released,
    w.reservation_released_at, w.metadata, w.error_message,
    w.created_at, w.updated_at,
    u.username, u.email, u.first_name, u.last_name
FROM withdrawals w
LEFT JOIN users u ON w.user_id = u.id
WHERE w.id = $1;

-- name: GetWithdrawalByWithdrawalID :one
SELECT 
    w.id, w.user_id, w.admin_id, w.chain_id, w.currency_code, w.protocol,
    w.withdrawal_id, w.usd_amount_cents, w.crypto_amount, w.exchange_rate,
    w.fee_cents, w.source_wallet_address, w.to_address, w.tx_hash,
    w.status, w.requires_admin_review, w.admin_review_deadline,
    w.processed_by_system, w.amount_reserved_cents, w.reservation_released,
    w.reservation_released_at, w.metadata, w.error_message,
    w.created_at, w.updated_at,
    u.username, u.email, u.first_name, u.last_name
FROM withdrawals w
LEFT JOIN users u ON w.user_id = u.id
WHERE w.withdrawal_id = $1;

-- name: GetWithdrawalsByUserID :many
SELECT 
    w.id, w.user_id, w.admin_id, w.chain_id, w.currency_code, w.protocol,
    w.withdrawal_id, w.usd_amount_cents, w.crypto_amount, w.exchange_rate,
    w.fee_cents, w.source_wallet_address, w.to_address, w.tx_hash,
    w.status, w.requires_admin_review, w.admin_review_deadline,
    w.processed_by_system, w.amount_reserved_cents, w.reservation_released,
    w.reservation_released_at, w.metadata, w.error_message,
    w.created_at, w.updated_at,
    u.username, u.email, u.first_name, u.last_name
FROM withdrawals w
LEFT JOIN users u ON w.user_id = u.id
WHERE w.user_id = $1
ORDER BY w.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetWithdrawalStats :one
SELECT 
    COUNT(*) as total_withdrawals,
    COUNT(*) FILTER (WHERE status = 'pending') as pending_withdrawals,
    COUNT(*) FILTER (WHERE status = 'processing') as processing_withdrawals,
    COUNT(*) FILTER (WHERE status = 'completed') as completed_withdrawals,
    COUNT(*) FILTER (WHERE status = 'failed') as failed_withdrawals,
    COUNT(*) FILTER (WHERE status = 'cancelled') as cancelled_withdrawals,
    COUNT(*) FILTER (WHERE created_at >= CURRENT_DATE) as today_withdrawals,
    COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '1 hour') as hourly_withdrawals,
    COALESCE(SUM(usd_amount_cents) FILTER (WHERE status IN ('completed', 'processing')), 0) as total_amount_cents,
    COALESCE(SUM(usd_amount_cents) FILTER (WHERE created_at >= CURRENT_DATE AND status IN ('completed', 'processing')), 0) as today_amount_cents,
    COALESCE(SUM(usd_amount_cents) FILTER (WHERE created_at >= NOW() - INTERVAL '1 hour' AND status IN ('completed', 'processing')), 0) as hourly_amount_cents
FROM withdrawals;

-- name: GetWithdrawalsByDateRange :many
SELECT 
    w.id, w.user_id, w.admin_id, w.chain_id, w.currency_code, w.protocol,
    w.withdrawal_id, w.usd_amount_cents, w.crypto_amount, w.exchange_rate,
    w.fee_cents, w.source_wallet_address, w.to_address, w.tx_hash,
    w.status, w.requires_admin_review, w.admin_review_deadline,
    w.processed_by_system, w.amount_reserved_cents, w.reservation_released,
    w.reservation_released_at, w.metadata, w.error_message,
    w.created_at, w.updated_at,
    u.username, u.email, u.first_name, u.last_name
FROM withdrawals w
LEFT JOIN users u ON w.user_id = u.id
WHERE w.created_at >= $1 AND w.created_at <= $2
ORDER BY w.created_at DESC
LIMIT $3 OFFSET $4;
