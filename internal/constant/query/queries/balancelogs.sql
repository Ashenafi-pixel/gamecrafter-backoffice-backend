-- name: GetBalanceLogByID :one
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
    bl.id = $1;

-- name: GetBalanceLogByTransactionID :one
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
    bl.transaction_id = $1;

-- name: SaveBalanceLogs :one 
INSERT INTO balance_logs (user_id,component,currency,change_amount,operational_group_id,operational_type_id,description,TIMESTAMP,balance_after_update,transaction_id,status)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
RETURNING *;