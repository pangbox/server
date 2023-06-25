-- name: AddItemToInventory :one
INSERT INTO inventory (
    player_id,
    item_type_id,
    quantity
) VALUES (
    ?,
    ?,
    ?
)
RETURNING *;

-- name: RemoveItemFromInventory :exec
DELETE FROM inventory WHERE player_id = ? AND item_id = ?;

-- name: SetItemQuantity :one
UPDATE inventory SET quantity = ? WHERE player_id = ? AND item_id = ? RETURNING *;

-- name: GetPlayerInventory :many
SELECT * FROM inventory WHERE player_id = ?;

-- name: GetItemsByTypeID :many
SELECT * FROM inventory WHERE player_id = ? AND item_type_id = ?;

-- name: GetItem :one
SELECT * FROM inventory WHERE player_id = ? AND item_id = ?;
