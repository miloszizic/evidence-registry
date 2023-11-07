-- name: SetCurrentUser :exec
-- Sets the current user in the session_data table.
INSERT INTO session_data (key, value)
VALUES ('current_user', $1)
ON CONFLICT (key)
DO UPDATE SET value = $1;

