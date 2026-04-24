-- name: CreateVerifiyEmail :execresult
INSERT INTO verify_emails (
  username, email, secret_code
) VALUES (
  ?, ?, ?
);

-- name: GetVerifiyEmail :one
SELECT * FROM verify_emails 
WHERE id = ?;
