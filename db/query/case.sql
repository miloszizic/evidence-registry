-- name: CreateCase :one
INSERT INTO "cases" (
  name,
  tags,
  case_year,
  case_type_id,
  case_number,
  case_court_id
) VALUES (
  $1, $2, $3, $4, $5, $6
) RETURNING *;


-- name: GetCase :one
SELECT * FROM "cases" WHERE id = $1;

-- name: ListCases :many
SELECT * FROM "cases";

-- name: UpdateCase :one
UPDATE "cases"
SET
  name = $2,
  tags = $3,
  case_year = $4,
  case_type_id = $5,
  case_number = $6,
  case_court_id = $7
WHERE id = $1
RETURNING *;


-- name: DeleteCase :exec
DELETE FROM "cases" WHERE id = $1;

-- name: DeleteCaseByName :exec
DELETE FROM "cases" WHERE name = $1;


-- name: CaseExists :one
SELECT EXISTS(SELECT 1 FROM "cases" WHERE name = $1);

-- name: GetCaseType :one
SELECT * FROM "case_types" WHERE id = $1;

-- name: CreateCaseType :one
INSERT INTO "case_types" (
  name,
  description
) VALUES (
  $1, $2
) RETURNING *;

-- name: UpdateCaseType :one
UPDATE "case_types"
SET
  name = $2,
  description = $3
WHERE id = $1
RETURNING *;

-- name: DeleteCaseType :exec
DELETE FROM "case_types" WHERE id = $1;

-- name: ListCaseTypes :many
SELECT * FROM "case_types";

-- name: CaseTypeExists :one
SELECT EXISTS(SELECT 1 FROM "case_types" WHERE name = $1);

-- name: CaseTypeExistsByID :one
SELECT EXISTS(SELECT 1 FROM "case_types" WHERE id = $1);

-- name: GetCaseTypeIDByName :one
SELECT id FROM "case_types" WHERE name = $1;

-- name: GetCaseIDTypes :many
SELECT * FROM "case_types";


-- name: GetCaseByName :one
SELECT * FROM "cases" WHERE name = $1;

-- name: GetCourtShortName :one
SELECT * FROM "courts" WHERE id = $1;

-- name: GetCourtIDByCode :one
SELECT id FROM "courts" WHERE code = $1;

-- name: GetCourtIDByShortName :one
SELECT id FROM "courts" WHERE short_name = $1;



