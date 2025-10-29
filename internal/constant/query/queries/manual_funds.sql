-- name: SaveManualFund :one 
INSERT INTO manual_funds (user_id,admin_id,transaction_id,type,amount_cents,reason,currency_code,note,created_at) VALUES(
  $1,$2,$3,$4,$5,$6,$7,$8,$9
) RETURNING *;

