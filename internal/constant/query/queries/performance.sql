-- name: GetFinancialMatrix :many 
SELECT 
    bl.currency,
    (SELECT COALESCE(SUM(bl_sub.change_amount), 0)::decimal
     FROM balance_logs bl_sub 
     JOIN operational_groups og_sub ON og_sub.id = bl_sub.operational_group_id 
     WHERE bl_sub.currency = bl.currency
       AND bl_sub.operational_type_id = (
           SELECT id 
           FROM operational_types 
           WHERE group_id = (SELECT id 
                              FROM operational_groups 
                              WHERE name = 'deposit' 
                              LIMIT 1) 
           LIMIT 1
       )) AS total_deposit_amount,

    (SELECT COALESCE(SUM(bl_sub.change_amount), 0)::decimal
     FROM balance_logs bl_sub 
     JOIN operational_groups og_sub ON og_sub.id = bl_sub.operational_group_id 
     WHERE bl_sub.currency = bl.currency
       AND bl_sub.operational_type_id = (
           SELECT id 
           FROM operational_types 
           WHERE group_id = (SELECT id 
                              FROM operational_groups 
                              WHERE name = 'withdrawal' 
                              LIMIT 1) 
           LIMIT 1
       )) AS total_withdrawal_amount,

    (SELECT COALESCE(COUNT(bl_sub.id), 0)::decimal
     FROM balance_logs bl_sub 
     JOIN operational_groups og_sub ON og_sub.id = bl_sub.operational_group_id 
     WHERE bl_sub.currency = bl.currency
       AND bl_sub.operational_type_id = (
           SELECT id 
           FROM operational_types 
           WHERE group_id = (SELECT id 
                              FROM operational_groups 
                              WHERE name = 'deposit' 
                              LIMIT 1) 
           LIMIT 1
       )) AS total_number_of_deposit,

    (SELECT COALESCE(COUNT(bl_sub.id), 0)::decimal
     FROM balance_logs bl_sub 
     JOIN operational_groups og_sub ON og_sub.id = bl_sub.operational_group_id 
     WHERE bl_sub.currency = bl.currency
       AND bl_sub.operational_type_id = (
           SELECT id 
           FROM operational_types 
           WHERE group_id = (SELECT id 
                              FROM operational_groups 
                              WHERE name = 'withdrawal' 
                              LIMIT 1) 
           LIMIT 1
       )) AS total_number_of_withdrawal
FROM balance_logs bl
GROUP BY bl.currency;

-- name: GetGameMatrics :one 
SELECT 
    COALESCE(COUNT(id), 0)::integer AS total_bets,
    COALESCE(SUM(amount), 0)::decimal AS total_bet_amount,
    COALESCE(AVG(amount), 0)::decimal AS average_bet_amount,
    COALESCE(MAX(amount), 0)::decimal AS highest_bet_amount,
    COALESCE(MIN(amount), 0)::decimal AS lowest_bet_amount,
    COALESCE(MAX(cash_out_multiplier), 0)::decimal AS max_multiplier,
    COALESCE(MIN(cash_out_multiplier), 0)::decimal AS min_multiplier,
    COALESCE(SUM(amount * cash_out_multiplier), 0)::decimal AS total_payout,
    COALESCE((SELECT COUNT(id) FROM bets WHERE cash_out_multiplier IS NOT NULL), 0)::integer AS total_wins,
    COALESCE((SELECT COUNT(id) FROM bets WHERE cash_out_multiplier IS NULL), 0)::integer AS total_losses,
    COALESCE(SUM(amount * cash_out_multiplier) / NULLIF(COUNT(id), 0), 0)::decimal AS avg_payout, 
    COALESCE(
        CASE 
            WHEN COUNT(id) = 0 THEN 0 
            ELSE (SELECT COUNT(id) FROM bets WHERE cash_out_multiplier IS NOT NULL) * 100.0 / COUNT(id) 
        END, 0)::decimal AS win_percentage,
    COALESCE(
        CASE 
            WHEN COUNT(id) = 0 THEN 0 
            ELSE (SELECT COUNT(id) FROM bets WHERE cash_out_multiplier IS NULL) * 100.0 / COUNT(id) 
        END, 0)::decimal AS loss_percentage
FROM bets;


