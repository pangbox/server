-- name: GetPlayer :one
SELECT
    player.*,
    character.*,
    inventory_character.item_type_id  AS character_type_id_FIXNULL,
    inventory_caddie.item_type_id     AS caddie_type_id_FIXNULL,
    inventory_club.item_type_id       AS club_type_id_FIXNULL,
    inventory_background.item_type_id AS background_type_id_FIXNULL,
    inventory_frame.item_type_id      AS frame_type_id_FIXNULL,
    inventory_sticker.item_type_id    AS sticker_type_id_FIXNULL,
    inventory_slot.item_type_id       AS slot_type_id_FIXNULL,
    inventory_cut_in.item_type_id     AS cut_in_type_id_FIXNULL,
    inventory_title.item_type_id      AS title_type_id_FIXNULL,
    inventory_poster0.item_type_id    AS poster0_type_id_FIXNULL,
    inventory_poster1.item_type_id    AS poster1_type_id_FIXNULL
FROM player AS player
LEFT JOIN character USING (character_id)
LEFT JOIN inventory AS inventory_character  ON (character.item_id    = inventory_character.item_id)
LEFT JOIN inventory AS inventory_caddie     ON (player.caddie_id     = inventory_caddie.item_id)
LEFT JOIN inventory AS inventory_club       ON (player.club_id       = inventory_club.item_id)
LEFT JOIN inventory AS inventory_background ON (player.background_id = inventory_background.item_id)
LEFT JOIN inventory AS inventory_frame      ON (player.frame_id      = inventory_frame.item_id)
LEFT JOIN inventory AS inventory_sticker    ON (player.sticker_id    = inventory_sticker.item_id)
LEFT JOIN inventory AS inventory_slot       ON (player.slot_id       = inventory_slot.item_id)
LEFT JOIN inventory AS inventory_cut_in     ON (player.cut_in_id     = inventory_cut_in.item_id)
LEFT JOIN inventory AS inventory_title      ON (player.title_id      = inventory_title.item_id)
LEFT JOIN inventory AS inventory_poster0    ON (player.poster0_id    = inventory_poster0.item_id)
LEFT JOIN inventory AS inventory_poster1    ON (player.poster1_id    = inventory_poster1.item_id)
WHERE player.player_id = ?
LIMIT 1;

-- name: GetPlayerByUsername :one
SELECT * FROM player
WHERE username = ?
LIMIT 1;

-- name: CreatePlayer :one
INSERT INTO player (
    username,
    nickname,
    password_hash,
    pang
) VALUES (
    ?, ?, ?, ?
)
RETURNING *;

-- name: SetPlayerNickname :one
UPDATE player SET nickname = ? WHERE player_id = ? RETURNING *;

-- name: SetPlayerCharacter :one
UPDATE player SET character_id = ? WHERE player_id = ? RETURNING *;

-- name: SetPlayerClubSet :one
UPDATE player SET club_id = ? WHERE player_id = ? RETURNING *;

-- name: SetPlayerCaddie :one
UPDATE player SET caddie_id = ? WHERE player_id = ? RETURNING *;

-- name: GetPlayerConsumables :one
SELECT
    slot0_type_id,
    slot1_type_id,
    slot2_type_id,
    slot3_type_id,
    slot4_type_id,
    slot5_type_id,
    slot6_type_id,
    slot7_type_id,
    slot8_type_id,
    slot9_type_id
FROM player
WHERE player_id = ?;

-- name: SetPlayerConsumables :one
UPDATE player
SET
    slot0_type_id = ?,
    slot1_type_id = ?,
    slot2_type_id = ?,
    slot3_type_id = ?,
    slot4_type_id = ?,
    slot5_type_id = ?,
    slot6_type_id = ?,
    slot7_type_id = ?,
    slot8_type_id = ?,
    slot9_type_id = ?
WHERE player_id = ?
RETURNING *;

-- name: SetPlayerComet :one
UPDATE player SET ball_type_id = ? WHERE player_id = ? RETURNING *;

-- name: SetPlayerDecoration :one
UPDATE player
SET
    background_id = ?,
    frame_id = ?,
    sticker_id = ?,
    slot_id = ?,
    cut_in_id = ?,
    title_id = ?
WHERE player_id = ?
RETURNING *;

-- name: GetPlayerCurrency :one
SELECT pang, points FROM player WHERE player_id = ?;

-- name: SetPlayerCurrency :one
UPDATE player SET pang = ?, points = ? WHERE player_id = ? RETURNING pang, points;

-- name: GetPlayerRank :one
SELECT rank, exp FROM player WHERE player_id = ?;

-- name: SetPlayerRank :one
UPDATE player SET rank = ?, exp = ? WHERE player_id = ? RETURNING rank, exp;
