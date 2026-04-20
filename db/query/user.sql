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

-- name: UpdateUser :exec
UPDATE users 
SET full_name = COALESCE(sqlc.narg(full_name), full_name), hashed_password = COALESCE(sqlc.narg(hashed_password), hashed_password), email = COALESCE(sqlc.narg(email), email), password_changed_at = COALESCE(sqlc.narg(password_changed_at), password_changed_at)
WHERE username = sqlc.arg(username);