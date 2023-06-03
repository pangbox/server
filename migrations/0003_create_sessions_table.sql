-- +goose Up
CREATE TABLE session (
    session_id INTEGER PRIMARY KEY AUTOINCREMENT,
    player_id INTEGER REFERENCES player(player_id) ON DELETE CASCADE NOT NULL,
    session_key TEXT NOT NULL UNIQUE,
    session_address TEXT NOT NULL,
    session_expires_at INTEGER NOT NULL
);

CREATE UNIQUE INDEX session_key_idx ON session (session_key);

-- +goose Down
DROP INDEX session_key_idx;

DROP TABLE session;
