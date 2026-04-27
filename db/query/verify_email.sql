-- name: CreateVerifiyEmail :execresult
INSERT INTO verify_emails (
  username, email, secret_code
) VALUES (
  ?, ?, ?
);

-- name: GetVerifiyEmail :one
SELECT * FROM verify_emails 
WHERE id = ?;

-- name: GetVerifiyEmailByUsername :one
SELECT * FROM verify_emails 
WHERE username = sqlc.arg(username);

-- name: UpdateVerifyEmail :execresult
UPDATE verify_emails
SET is_used = true
WHERE username=sqlc.arg(username) AND secret_code=sqlc.arg(secret_code) AND expired_at>=NOW() AND created_at<=NOW() AND is_used=false;
