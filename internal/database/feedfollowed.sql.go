// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: feedfollowed.sql

package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createFollow = `-- name: CreateFollow :one
INSERT INTO feed_followed (id, feed_id, user_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, feed_id, user_id, created_at, updated_at
`

type CreateFollowParams struct {
	ID        uuid.UUID
	FeedID    uuid.UUID
	UserID    uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (q *Queries) CreateFollow(ctx context.Context, arg CreateFollowParams) (FeedFollowed, error) {
	row := q.db.QueryRowContext(ctx, createFollow,
		arg.ID,
		arg.FeedID,
		arg.UserID,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	var i FeedFollowed
	err := row.Scan(
		&i.ID,
		&i.FeedID,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteFollow = `-- name: DeleteFollow :exec
DELETE FROM feed_followed
WHERE id = $1
`

func (q *Queries) DeleteFollow(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, deleteFollow, id)
	return err
}

const getFollowByAPIKey = `-- name: GetFollowByAPIKey :many
SELECT id, feed_id, user_id, created_at, updated_at FROM feed_followed
WHERE user_id = $1
`

func (q *Queries) GetFollowByAPIKey(ctx context.Context, userID uuid.UUID) ([]FeedFollowed, error) {
	rows, err := q.db.QueryContext(ctx, getFollowByAPIKey, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FeedFollowed
	for rows.Next() {
		var i FeedFollowed
		if err := rows.Scan(
			&i.ID,
			&i.FeedID,
			&i.UserID,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
