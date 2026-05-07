-- name: GetAccountsForUpdate :many
SELECT * FROM accounts
WHERE id IN (?,?)
FOR UPDATE;

-- name: GetAccount :one
SELECT * FROM accounts
WHERE id = ?;

-- name: ListAllAccountIdByUsername :many
SELECT a.id FROM accounts a 
INNER JOIN users u ON a.owner = u.username 
WHERE u.username = sqlc.arg(username);

-- name: ListAccounts :many
SELECT * FROM accounts
ORDER BY id;

-- name: ListPagedAccounts :many
SELECT * FROM accounts
WHERE owner = ?
ORDER BY id
LIMIT ?
OFFSET ?;

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
UPDATE accounts SET owner = ?, balance = ?, currency = ?
WHERE id = ?;

-- name: AddAccountBalance :exec
UPDATE accounts SET balance = balance+sqlc.arg(amount)
WHERE id = sqlc.arg(id);