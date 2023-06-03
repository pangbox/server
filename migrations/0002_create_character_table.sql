-- +goose Up
CREATE TABLE character (
    character_id INTEGER PRIMARY KEY AUTOINCREMENT,
    player_id INTEGER REFERENCES player(player_id) ON DELETE CASCADE NOT NULL,
    character_type_id INTEGER NOT NULL,
    character_data BLOB NOT NULL
);

-- +goose Down
DROP TABLE character;
