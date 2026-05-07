-- name: GetEntries :one
SELECT * FROM entries
WHERE id = ?;

-- name: GetEntriesByAccount :one
SELECT * FROM entries
WHERE account_id = sqlc.arg(account_id) AND id = ?;

-- name: ListEntries :many
SELECT * FROM entries
ORDER BY id;

-- name: ListEntriesByAccountIdAndUsername :many
SELECT e.id, e.account_id, e.amount, e.created_at FROM entries e 
INNER JOIN accounts a ON e.account_id = a.id 
INNER JOIN users u ON u.username = a.owner WHERE u.username = sqlc.arg(username) AND a.id = sqlc.arg(account_id) ;

-- name: CreateEntries :execresult
INSERT INTO entries (
  account_id, amount
) VALUES (
  ?, ?
);

-- name: DeleteEntries :exec
DELETE FROM entries
WHERE id = ?;