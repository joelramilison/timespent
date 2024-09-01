// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: sessions.sql

package database

import (
	"context"

	"github.com/google/uuid"
)

const getNewestSession = `-- name: GetNewestSession :one
SELECT id, created_at, updated_at, started_at, ended_at, pause_seconds, user_id FROM sessions
WHERE user_id = $1
ORDER BY started_at DESC
LIMIT 1
`

func (q *Queries) GetNewestSession(ctx context.Context, userID uuid.UUID) (Session, error) {
	row := q.db.QueryRowContext(ctx, getNewestSession, userID)
	var i Session
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.StartedAt,
		&i.EndedAt,
		&i.PauseSeconds,
		&i.UserID,
	)
	return i, err
}

const startSession = `-- name: StartSession :exec
INSERT INTO sessions(
    id,
    created_at,
    updated_at,
    started_at,
    user_id
)
VALUES (
    $1, NOW(), NOW(), NOW(), $2
)
`

type StartSessionParams struct {
	ID     uuid.UUID
	UserID uuid.UUID
}

func (q *Queries) StartSession(ctx context.Context, arg StartSessionParams) error {
	_, err := q.db.ExecContext(ctx, startSession, arg.ID, arg.UserID)
	return err
}

const stopSession = `-- name: StopSession :exec
UPDATE sessions
SET updated_at = NOW(), ended_at = NOW(), pause_seconds = $1
WHERE id = $2
`

type StopSessionParams struct {
	PauseSeconds int32
	ID           uuid.UUID
}

func (q *Queries) StopSession(ctx context.Context, arg StopSessionParams) error {
	_, err := q.db.ExecContext(ctx, stopSession, arg.PauseSeconds, arg.ID)
	return err
}
