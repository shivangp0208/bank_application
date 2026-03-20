-- name: GetAccounts :one
SELECT * FROM accounts
WHERE id = ? LIMIT 1;

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