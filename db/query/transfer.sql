-- name: GetTransfers :one
SELECT * FROM transfers
WHERE id = ?;

-- name: ListTransfers :many
SELECT * FROM transfers
ORDER BY id;

-- name: CreateTransfers :execresult
INSERT INTO transfers (
  from_account_id, to_account_id, amount
) VALUES (
  ?, ?, ?
);

-- name: DeleteTransfers :exec
DELETE FROM transfers
WHERE id = ?;