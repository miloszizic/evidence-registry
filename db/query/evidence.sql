-- name: CreateEvidence :one
INSERT INTO "evidence" (
  case_id,
  app_user_id,
  name,
  description,
  hash,
  evidence_type_id
) VALUES (
  $1, $2, $3, $4, $5, $6
) RETURNING *;


-- name: GetEvidence :one
SELECT * FROM "evidence" WHERE id = $1;

-- name: ListEvidence :many
SELECT * FROM "evidence";

-- name: ListEvidenceTypes :many
SELECT * FROM "evidence_types";

-- name: UpdateEvidenceDescription :exec
UPDATE "evidence" SET description = $1 WHERE id = $2;


-- name: DeleteEvidence :exec
DELETE FROM "evidence" WHERE id = $1;

-- name: EvidenceExists :one
SELECT EXISTS (SELECT 1 FROM "evidence" WHERE name = $1 AND case_id = $2);

-- name: GetEvidencesByCaseID :many
SELECT * FROM "evidence" WHERE case_id = $1;

-- name: GetEvidenceIDByType :one
SELECT id FROM "evidence_types" WHERE name = $1;
