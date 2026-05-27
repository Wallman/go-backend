-- name: GetUser :one
SELECT id, tokens_used FROM users WHERE id = $1;

-- name: AddTokensUsed :one
UPDATE users SET tokens_used = tokens_used + $1 WHERE id = $2 RETURNING id, tokens_used;

-- name: CreateUser :one
INSERT INTO users (id, tokens_used) VALUES ($1, 0) RETURNING id, tokens_used;
