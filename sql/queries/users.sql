-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUser
SELECT id, created_at, updated_at, name FROM users
WHERE api_key = $1;