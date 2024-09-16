-- name: CreateFollow :one
INSERT INTO feed_followed (id, feed_id, user_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetFollowByAPIKey :many
SELECT * FROM feed_followed
WHERE user_id = $1;

-- name: DeleteFollow :exec
DELETE FROM feed_followed
WHERE id = $1;