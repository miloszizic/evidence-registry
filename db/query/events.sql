-- Calendar Events

-- name: CreateEvent :one
INSERT INTO "calendar_events" (
  user_id,
  case_id,
  event_date,
  notes
) VALUES (
  $1, $2, $3, $4
) RETURNING *;

-- name: GetEvent :one
SELECT * FROM "calendar_events" WHERE id = $1;

-- name: ListEvents :many
SELECT * FROM "calendar_events";

-- name: UpdateEvent :one
UPDATE "calendar_events"
SET
  user_id = $2,
  case_id = $3,
  event_date = $4,
  notes = $5
WHERE id = $1
RETURNING *;

-- name: DeleteEvent :exec
DELETE FROM "calendar_events" WHERE id = $1;

-- name: EventExists :one
SELECT EXISTS(SELECT 1 FROM "calendar_events" WHERE id = $1);
