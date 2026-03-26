-- name: GetUser :one
SELECT * FROM users
WHERE username = ?;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY username;

-- name: ListPagedUsers :many
SELECT * FROM users
ORDER BY username
LIMIT ?
OFFSET ?;

-- name: CreateUser :execresult
INSERT INTO users (
  username, hashed_password, full_name, email
) VALUES (
  ?, ?, ?, ?
);

-- name: DeleteUser :exec
DELETE FROM users
WHERE username = ?;