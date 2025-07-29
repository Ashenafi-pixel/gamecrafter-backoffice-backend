-- name: GetExchangesFromTo :one
SELECT * FROM exchange_rates where currency_from = $1 and currency_to = $2;
