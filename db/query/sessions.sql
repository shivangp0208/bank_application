-- name: GetSession :one
SELECT * FROM sessions
WHERE id = ?;

-- name: CreateSession :execresult
INSERT INTO sessions (
  id, username, refresh_token, user_agent, client_ip, is_blocked, expires_at
) VALUES (
  ?, ?, ?, ?, ?, ?, ?
);

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE id = ?;