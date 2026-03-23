-- name: GetEntries :one
SELECT * FROM entries
WHERE id = ?;

-- name: ListEntries :many
SELECT * FROM entries
ORDER BY id;

-- name: CreateEntries :execresult
INSERT INTO entries (
  account_id, amount
) VALUES (
  ?, ?
);

-- name: DeleteEntries :exec
DELETE FROM entries
WHERE id = ?;