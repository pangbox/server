-- +goose Up
CREATE TABLE player (
    player_id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    nickname TEXT UNIQUE,
    password_hash TEXT NOT NULL,
    pang INTEGER NOT NULL DEFAULT 10000,
    rank INTEGER NOT NULL DEFAULT 0
);

-- +goose Down
DROP TABLE player;
