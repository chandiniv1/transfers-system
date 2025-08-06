-- name: CreateAccount :one
INSERT INTO accounts (account_id, balance, currency)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetAccount :one
SELECT * FROM accounts
WHERE account_id = $1
LIMIT 1;

-- name: UpdateBalance :one
UPDATE accounts
SET balance = balance + sqlc.arg(amount)
WHERE account_id = sqlc.arg(account_id)
RETURNING *;

-- name: ListAccounts :many
SELECT * FROM accounts
ORDER BY account_id
LIMIT $1
OFFSET $2;


