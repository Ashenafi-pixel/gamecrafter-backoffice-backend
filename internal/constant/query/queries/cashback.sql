-- Cashback System Queries

-- name: CreateUserLevel :one
INSERT INTO user_levels (user_id, current_level, total_ggr, total_bets, total_wins, level_progress, current_tier_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetUserLevel :one
SELECT ul.*, ct.tier_name, ct.cashback_percentage, ct.bonus_multiplier, ct.special_benefits
FROM user_levels ul
LEFT JOIN cashback_tiers ct ON ul.current_tier_id = ct.id
WHERE ul.user_id = $1;

-- name: UpdateUserLevel :one
UPDATE user_levels 
SET 
    current_level = $2,
    total_ggr = $3,
    total_bets = $4,
    total_wins = $5,
    level_progress = $6,
    current_tier_id = $7,
    last_level_up = $8,
    updated_at = NOW()
WHERE user_id = $1
RETURNING *;

-- name: GetCashbackTiers :many
SELECT * FROM cashback_tiers 
WHERE is_active = true 
ORDER BY tier_level ASC;

-- name: GetCashbackTierByLevel :one
SELECT * FROM cashback_tiers 
WHERE tier_level = $1 AND is_active = true;

-- name: GetCashbackTierByGGR :one
SELECT * FROM cashback_tiers 
WHERE min_ggr_required <= $1 AND is_active = true
ORDER BY tier_level DESC
LIMIT 1;

-- name: CreateCashbackEarning :one
INSERT INTO cashback_earnings (
    user_id, tier_id, earning_type, source_bet_id, ggr_amount, 
    cashback_rate, earned_amount, available_amount, status, expires_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetUserCashbackEarnings :many
SELECT ce.*, ct.tier_name, ct.cashback_percentage
FROM cashback_earnings ce
JOIN cashback_tiers ct ON ce.tier_id = ct.id
WHERE ce.user_id = $1
ORDER BY ce.created_at DESC;

-- name: GetAvailableCashbackEarnings :many
SELECT ce.*, ct.tier_name, ct.cashback_percentage
FROM cashback_earnings ce
JOIN cashback_tiers ct ON ce.tier_id = ct.id
WHERE ce.user_id = $1 AND ce.status = 'available' AND ce.available_amount > 0
ORDER BY ce.created_at ASC;

-- name: UpdateCashbackEarningStatus :one
UPDATE cashback_earnings 
SET 
    status = $2,
    claimed_amount = $3,
    available_amount = $4,
    claimed_at = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CreateCashbackClaim :one
INSERT INTO cashback_claims (
    user_id, claim_amount, currency_code, status, processing_fee, 
    net_amount, claimed_earnings, admin_notes
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateCashbackClaimStatus :one
UPDATE cashback_claims 
SET 
    status = $2,
    transaction_id = $3,
    processed_at = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetUserCashbackClaims :many
SELECT * FROM cashback_claims 
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetCashbackClaim :one
SELECT * FROM cashback_claims 
WHERE id = $1;

-- name: GetGameHouseEdge :one
SELECT * FROM game_house_edges 
WHERE game_type = $1 
AND (game_variant = $2 OR game_variant IS NULL)
AND is_active = true
AND effective_from <= NOW()
AND (effective_until IS NULL OR effective_until > NOW())
ORDER BY effective_from DESC
LIMIT 1;

-- name: CreateGameHouseEdge :one
INSERT INTO game_house_edges (
    game_type, game_variant, house_edge, min_bet, max_bet, 
    is_active, effective_from, effective_until
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetActiveCashbackPromotions :many
SELECT * FROM cashback_promotions 
WHERE is_active = true 
AND starts_at <= NOW()
AND (ends_at IS NULL OR ends_at > NOW())
ORDER BY starts_at DESC;

-- name: GetCashbackPromotionForUser :one
SELECT cp.* FROM cashback_promotions cp
WHERE cp.is_active = true 
AND cp.starts_at <= NOW()
AND (cp.ends_at IS NULL OR cp.ends_at > NOW())
AND (
    cp.target_tiers IS NULL OR 
    $1 = ANY(cp.target_tiers)
)
AND (
    cp.target_games IS NULL OR 
    $2 = ANY(cp.target_games)
)
ORDER BY cp.starts_at DESC
LIMIT 1;

-- name: CreateCashbackPromotion :one
INSERT INTO cashback_promotions (
    promotion_name, description, promotion_type, boost_multiplier, 
    bonus_amount, min_bet_amount, max_bonus_amount, target_tiers, 
    target_games, is_active, starts_at, ends_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: UpdateCashbackPromotion :one
UPDATE cashback_promotions 
SET 
    promotion_name = $2,
    description = $3,
    promotion_type = $4,
    boost_multiplier = $5,
    bonus_amount = $6,
    min_bet_amount = $7,
    max_bonus_amount = $8,
    target_tiers = $9,
    target_games = $10,
    is_active = $11,
    starts_at = $12,
    ends_at = $13,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetUserCashbackSummary :one
SELECT 
    ul.user_id,
    ul.current_level,
    ul.total_ggr,
    ul.level_progress,
    ct.tier_name,
    ct.cashback_percentage,
    ct.bonus_multiplier,
    ct.daily_cashback_limit,
    ct.weekly_cashback_limit,
    ct.monthly_cashback_limit,
    ct.special_benefits,
    COALESCE(SUM(ce.available_amount), 0) as available_cashback,
    COALESCE(SUM(CASE WHEN ce.status = 'pending' THEN ce.earned_amount ELSE 0 END), 0) as pending_cashback,
    COALESCE(SUM(ce.claimed_amount), 0) as total_claimed,
    (
        SELECT min_ggr_required 
        FROM cashback_tiers 
        WHERE tier_level = ul.current_level + 1 
        AND is_active = true
    ) as next_tier_ggr
FROM user_levels ul
LEFT JOIN cashback_tiers ct ON ul.current_tier_id = ct.id
LEFT JOIN cashback_earnings ce ON ul.user_id = ce.user_id
WHERE ul.user_id = $1
GROUP BY ul.user_id, ul.current_level, ul.total_ggr, ul.level_progress, 
         ct.tier_name, ct.cashback_percentage, ct.bonus_multiplier,
         ct.daily_cashback_limit, ct.weekly_cashback_limit, ct.monthly_cashback_limit,
         ct.special_benefits;

-- name: GetCashbackStats :one
SELECT 
    COUNT(DISTINCT ul.user_id) as total_users_with_cashback,
    COALESCE(SUM(ce.earned_amount), 0) as total_cashback_earned,
    COALESCE(SUM(ce.claimed_amount), 0) as total_cashback_claimed,
    COALESCE(SUM(ce.available_amount), 0) as total_cashback_pending,
    COALESCE(AVG(ct.cashback_percentage), 0) as average_cashback_rate,
    COALESCE(SUM(CASE WHEN DATE(cc.created_at) = CURRENT_DATE THEN cc.claim_amount ELSE 0 END), 0) as daily_cashback_claims,
    COALESCE(SUM(CASE WHEN cc.created_at >= DATE_TRUNC('week', CURRENT_DATE) THEN cc.claim_amount ELSE 0 END), 0) as weekly_cashback_claims,
    COALESCE(SUM(CASE WHEN cc.created_at >= DATE_TRUNC('month', CURRENT_DATE) THEN cc.claim_amount ELSE 0 END), 0) as monthly_cashback_claims
FROM user_levels ul
LEFT JOIN cashback_tiers ct ON ul.current_tier_id = ct.id
LEFT JOIN cashback_earnings ce ON ul.user_id = ce.user_id
LEFT JOIN cashback_claims cc ON ul.user_id = cc.user_id;

-- name: GetTierDistribution :many
SELECT 
    ct.tier_name,
    COUNT(ul.user_id) as user_count
FROM cashback_tiers ct
LEFT JOIN user_levels ul ON ct.id = ul.current_tier_id
WHERE ct.is_active = true
GROUP BY ct.tier_name, ct.tier_level
ORDER BY ct.tier_level;

-- name: GetExpiredCashbackEarnings :many
SELECT * FROM cashback_earnings 
WHERE status = 'available' 
AND expires_at < NOW()
AND available_amount > 0;

-- name: MarkCashbackEarningsAsExpired :exec
UPDATE cashback_earnings 
SET 
    status = 'expired',
    updated_at = NOW()
WHERE status = 'available' 
AND expires_at < NOW()
AND available_amount > 0;

-- name: GetUserDailyCashbackLimit :one
SELECT 
    COALESCE(SUM(cc.claim_amount), 0) as daily_claimed
FROM cashback_claims cc
WHERE cc.user_id = $1 
AND DATE(cc.created_at) = CURRENT_DATE
AND cc.status IN ('completed', 'processing');

-- name: GetUserWeeklyCashbackLimit :one
SELECT 
    COALESCE(SUM(cc.claim_amount), 0) as weekly_claimed
FROM cashback_claims cc
WHERE cc.user_id = $1 
AND cc.created_at >= DATE_TRUNC('week', CURRENT_DATE)
AND cc.status IN ('completed', 'processing');

-- name: GetUserMonthlyCashbackLimit :one
SELECT 
    COALESCE(SUM(cc.claim_amount), 0) as monthly_claimed
FROM cashback_claims cc
WHERE cc.user_id = $1 
AND cc.created_at >= DATE_TRUNC('month', CURRENT_DATE)
AND cc.status IN ('completed', 'processing');