-- name: GetAccountsForUpdate :many
SELECT * FROM accounts
WHERE id IN (?,?)
FOR UPDATE;

-- name: GetAccount :one
SELECT * FROM accounts
WHERE id = ?;

-- name: ListAccounts :many
SELECT * FROM accounts
ORDER BY id;

-- name: CreateAccounts :execresult
INSERT INTO accounts (
  owner, balance, currency
) VALUES (
  ?, ?, ?
);

-- name: DeleteAccounts :exec
DELETE FROM accounts
WHERE id = ?;

-- name: UpdateAccount :exec
UPDATE accounts SET balance = ?
WHERE id = ?;

-- name: AddAccountBalance :exec
UPDATE accounts SET balance = balance+sqlc.arg(amount)
WHERE id = sqlc.arg(id);