// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.1
// source: court.sql

package db

import (
	"context"

	"github.com/google/uuid"
)

const createCourt = `-- name: CreateCourt :one
INSERT INTO "courts" (
  name
) VALUES (
  $1
) RETURNING id, code, name, short_name
`

func (q *Queries) CreateCourt(ctx context.Context, name string) (Court, error) {
	row := q.db.QueryRowContext(ctx, createCourt, name)
	var i Court
	err := row.Scan(
		&i.ID,
		&i.Code,
		&i.Name,
		&i.ShortName,
	)
	return i, err
}

const deleteCourt = `-- name: DeleteCourt :exec
DELETE FROM "courts" WHERE id = $1
`

func (q *Queries) DeleteCourt(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, deleteCourt, id)
	return err
}

const getCourt = `-- name: GetCourt :one
SELECT id, code, name, short_name FROM "courts" WHERE id = $1
`

func (q *Queries) GetCourt(ctx context.Context, id uuid.UUID) (Court, error) {
	row := q.db.QueryRowContext(ctx, getCourt, id)
	var i Court
	err := row.Scan(
		&i.ID,
		&i.Code,
		&i.Name,
		&i.ShortName,
	)
	return i, err
}

const listCourts = `-- name: ListCourts :many
SELECT id, code, name, short_name FROM "courts"
`

func (q *Queries) ListCourts(ctx context.Context) ([]Court, error) {
	rows, err := q.db.QueryContext(ctx, listCourts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Court{}
	for rows.Next() {
		var i Court
		if err := rows.Scan(
			&i.ID,
			&i.Code,
			&i.Name,
			&i.ShortName,
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

const updateCourt = `-- name: UpdateCourt :one
UPDATE "courts"
SET
  name = $2
WHERE id = $1
RETURNING id, code, name, short_name
`

type UpdateCourtParams struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func (q *Queries) UpdateCourt(ctx context.Context, arg UpdateCourtParams) (Court, error) {
	row := q.db.QueryRowContext(ctx, updateCourt, arg.ID, arg.Name)
	var i Court
	err := row.Scan(
		&i.ID,
		&i.Code,
		&i.Name,
		&i.ShortName,
	)
	return i, err
}