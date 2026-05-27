-- name: GetUser :one
SELECT id, tokens_used, version FROM users WHERE id = $1;

-- name: AddTokensUsed :one
UPDATE users SET tokens_used = tokens_used + $1, version = version + 1 WHERE id = $2 AND version = $3 RETURNING id, tokens_used, version;

-- name: CreateUser :one
INSERT INTO users (id, tokens_used, version) VALUES ($1, 0, 0) RETURNING id, tokens_used, version;
