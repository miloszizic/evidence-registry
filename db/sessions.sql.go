// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.1
// source: sessions.sql

package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createSession = `-- name: CreateSession :one
INSERT INTO sessions (user_id,refresh_payload_id,username, refresh_token, user_agent, client_ip, is_blocked, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, user_id, refresh_payload_id, username, refresh_token, user_agent, client_ip, is_blocked, expires_at, created_at
`

type CreateSessionParams struct {
	UserID           uuid.UUID `json:"user_id"`
	RefreshPayloadID uuid.UUID `json:"refresh_payload_id"`
	Username         string    `json:"username"`
	RefreshToken     string    `json:"refresh_token"`
	UserAgent        string    `json:"user_agent"`
	ClientIp         string    `json:"client_ip"`
	IsBlocked        bool      `json:"is_blocked"`
	ExpiresAt        time.Time `json:"expires_at"`
}

func (q *Queries) CreateSession(ctx context.Context, arg CreateSessionParams) (Session, error) {
	row := q.db.QueryRowContext(ctx, createSession,
		arg.UserID,
		arg.RefreshPayloadID,
		arg.Username,
		arg.RefreshToken,
		arg.UserAgent,
		arg.ClientIp,
		arg.IsBlocked,
		arg.ExpiresAt,
	)
	var i Session
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.RefreshPayloadID,
		&i.Username,
		&i.RefreshToken,
		&i.UserAgent,
		&i.ClientIp,
		&i.IsBlocked,
		&i.ExpiresAt,
		&i.CreatedAt,
	)
	return i, err
}

const getSession = `-- name: GetSession :one
SELECT id, user_id, refresh_payload_id, username, refresh_token, user_agent, client_ip, is_blocked, expires_at, created_at
FROM sessions
WHERE refresh_payload_id = $1
`

func (q *Queries) GetSession(ctx context.Context, refreshPayloadID uuid.UUID) (Session, error) {
	row := q.db.QueryRowContext(ctx, getSession, refreshPayloadID)
	var i Session
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.RefreshPayloadID,
		&i.Username,
		&i.RefreshToken,
		&i.UserAgent,
		&i.ClientIp,
		&i.IsBlocked,
		&i.ExpiresAt,
		&i.CreatedAt,
	)
	return i, err
}

const invalidateSession = `-- name: InvalidateSession :exec
UPDATE sessions
SET is_blocked = true
WHERE id = $1
`

func (q *Queries) InvalidateSession(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, invalidateSession, id)
	return err
}
