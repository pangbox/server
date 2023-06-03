-- name: GetPlayer :one
SELECT * FROM player
WHERE player_id = ? LIMIT 1;

-- name: GetPlayerByUsername :one
SELECT * FROM player
WHERE username = ? LIMIT 1;

-- name: CreatePlayer :one
INSERT INTO player (
    username,
    nickname,
    password_hash
) VALUES (
    ?, ?, ?
)
RETURNING *;

-- name: SetPlayerNickname :one
UPDATE player SET nickname = ? WHERE player_id = ? RETURNING *;
