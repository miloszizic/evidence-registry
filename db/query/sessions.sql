-- name: CreateSession :one
INSERT INTO sessions (user_id,refresh_payload_id,username, refresh_token, user_agent, client_ip, is_blocked, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;


-- name: GetSession :one
SELECT *
FROM sessions
WHERE refresh_payload_id = $1;

-- name: InvalidateSession :exec
UPDATE sessions
SET is_blocked = true
WHERE id = $1;

