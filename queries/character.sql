-- name: GetCharacter :one
SELECT
    character.*,
    inventory_character.item_type_id AS character_type_id,
    inventory_aux_part0.item_type_id AS inventory_aux_part0_type_id_FIXNULL,
    inventory_aux_part1.item_type_id AS inventory_aux_part1_type_id_FIXNULL,
    inventory_aux_part2.item_type_id AS inventory_aux_part2_type_id_FIXNULL,
    inventory_aux_part3.item_type_id AS inventory_aux_part3_type_id_FIXNULL,
    inventory_aux_part4.item_type_id AS inventory_aux_part4_type_id_FIXNULL
FROM character AS character
LEFT JOIN inventory AS inventory_character ON (character.item_id = inventory_character.item_id)
LEFT JOIN inventory AS inventory_aux_part0 ON (character.aux_part0_id = inventory_aux_part0.item_id)
LEFT JOIN inventory AS inventory_aux_part1 ON (character.aux_part1_id = inventory_aux_part1.item_id)
LEFT JOIN inventory AS inventory_aux_part2 ON (character.aux_part2_id = inventory_aux_part2.item_id)
LEFT JOIN inventory AS inventory_aux_part3 ON (character.aux_part3_id = inventory_aux_part3.item_id)
LEFT JOIN inventory AS inventory_aux_part4 ON (character.aux_part4_id = inventory_aux_part4.item_id)
WHERE character.character_id = ? LIMIT 1;

-- name: GetCharactersByPlayer :many
SELECT
    character.*,
    inventory_character.item_type_id AS character_type_id
FROM character AS character
LEFT JOIN inventory AS inventory_character ON (character.item_id = inventory_character.item_id)
WHERE character.player_id = ?;

-- name: PlayerHasCharacters :one
SELECT count(*) > 0 FROM character
WHERE character.player_id = ?;

-- name: CreateCharacter :one
INSERT INTO character (
    player_id,
    item_id,
    hair_color,
    shirt,
    mastery,
    part00_item_type_id,
    part01_item_type_id,
    part02_item_type_id,
    part03_item_type_id,
    part04_item_type_id,
    part05_item_type_id,
    part06_item_type_id,
    part07_item_type_id,
    part08_item_type_id,
    part09_item_type_id,
    part10_item_type_id,
    part11_item_type_id,
    part12_item_type_id,
    part13_item_type_id,
    part14_item_type_id,
    part15_item_type_id,
    part16_item_type_id,
    part17_item_type_id,
    part18_item_type_id,
    part19_item_type_id,
    part20_item_type_id,
    part21_item_type_id,
    part22_item_type_id,
    part23_item_type_id
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
)
RETURNING *;

-- name: SetCharacterParts :one
UPDATE character
SET
    part00_item_id = ?,
    part01_item_id = ?,
    part02_item_id = ?,
    part03_item_id = ?,
    part04_item_id = ?,
    part05_item_id = ?,
    part06_item_id = ?,
    part07_item_id = ?,
    part08_item_id = ?,
    part09_item_id = ?,
    part10_item_id = ?,
    part11_item_id = ?,
    part12_item_id = ?,
    part13_item_id = ?,
    part14_item_id = ?,
    part15_item_id = ?,
    part16_item_id = ?,
    part17_item_id = ?,
    part18_item_id = ?,
    part19_item_id = ?,
    part20_item_id = ?,
    part21_item_id = ?,
    part22_item_id = ?,
    part23_item_id = ?,
    part00_item_type_id = ?,
    part01_item_type_id = ?,
    part02_item_type_id = ?,
    part03_item_type_id = ?,
    part04_item_type_id = ?,
    part05_item_type_id = ?,
    part06_item_type_id = ?,
    part07_item_type_id = ?,
    part08_item_type_id = ?,
    part09_item_type_id = ?,
    part10_item_type_id = ?,
    part11_item_type_id = ?,
    part12_item_type_id = ?,
    part13_item_type_id = ?,
    part14_item_type_id = ?,
    part15_item_type_id = ?,
    part16_item_type_id = ?,
    part17_item_type_id = ?,
    part18_item_type_id = ?,
    part19_item_type_id = ?,
    part20_item_type_id = ?,
    part21_item_type_id = ?,
    part22_item_type_id = ?,
    part23_item_type_id = ?,
    cut_in_id = ?
WHERE character_id = ?
RETURNING *;
