-- name: CreateCourt :one
INSERT INTO "courts" (
  name
) VALUES (
  $1
) RETURNING *;

-- name: GetCourt :one
SELECT * FROM "courts" WHERE id = $1;

-- name: ListCourts :many
SELECT * FROM "courts";

-- name: UpdateCourt :one
UPDATE "courts"
SET
  name = $2
WHERE id = $1
RETURNING *;

-- name: DeleteCourt :exec
DELETE FROM "courts" WHERE id = $1;
