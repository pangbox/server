-- name: GetCharacter :one
SELECT * FROM character
WHERE character_id = ? LIMIT 1;

-- name: GetCharactersByPlayer :many
SELECT * FROM character
WHERE player_id = ?;

-- name: PlayerHasCharacters :one
SELECT count(*) > 0 FROM character
WHERE player_id = ?;

-- name: CreateCharacter :one
INSERT INTO character (
    player_id,
    character_type_id,
    character_data
) VALUES (
    ?, ?, ?
)
RETURNING *;
