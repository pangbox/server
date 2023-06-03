-- name: GetSession :one
SELECT * FROM session
WHERE session_id = ? LIMIT 1;

-- name: GetSessionByKey :one
SELECT * FROM session
WHERE session_key = ? LIMIT 1;

-- name: GetSessionsByPlayer :many
SELECT * FROM session
WHERE player_id = ?;

-- name: CreateSession :one
INSERT INTO session (
    player_id,
    session_key,
    session_address,
    session_expires_at
) VALUES (
    ?, ?, ?, ?
)
RETURNING *;

-- name: UpdateSessionExpiry :one
UPDATE session SET session_expires_at = ? WHERE session_id = ? RETURNING *;

-- name: DeleteExpiredSessions :exec
DELETE FROM session WHERE session_expires_at < ?;
