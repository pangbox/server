-- +goose Up
CREATE TABLE player (
    player_id      INTEGER PRIMARY KEY AUTOINCREMENT,
    username       TEXT NOT NULL UNIQUE,
    nickname       TEXT UNIQUE,
    password_hash  TEXT NOT NULL,
    pang           INTEGER NOT NULL DEFAULT 10000,
    points         INTEGER NOT NULL DEFAULT 0,
    rank           INTEGER NOT NULL DEFAULT 0,
    ball_type_id   INTEGER NOT NULL DEFAULT 0,
    mascot_type_id INTEGER NOT NULL DEFAULT 0,
    slot0_type_id  INTEGER NOT NULL DEFAULT 0,
    slot1_type_id  INTEGER NOT NULL DEFAULT 0,
    slot2_type_id  INTEGER NOT NULL DEFAULT 0,
    slot3_type_id  INTEGER NOT NULL DEFAULT 0,
    slot4_type_id  INTEGER NOT NULL DEFAULT 0,
    slot5_type_id  INTEGER NOT NULL DEFAULT 0,
    slot6_type_id  INTEGER NOT NULL DEFAULT 0,
    slot7_type_id  INTEGER NOT NULL DEFAULT 0,
    slot8_type_id  INTEGER NOT NULL DEFAULT 0,
    slot9_type_id  INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE inventory (
    item_id INTEGER PRIMARY KEY AUTOINCREMENT,
    player_id INTEGER REFERENCES player(player_id) ON DELETE CASCADE NOT NULL,
    item_type_id INTEGER NOT NULL,
    quantity INTEGER
);

ALTER TABLE player ADD COLUMN caddie_id     INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE;
ALTER TABLE player ADD COLUMN club_id       INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE;
ALTER TABLE player ADD COLUMN background_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE;
ALTER TABLE player ADD COLUMN frame_id      INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE;
ALTER TABLE player ADD COLUMN sticker_id    INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE;
ALTER TABLE player ADD COLUMN slot_id       INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE;
ALTER TABLE player ADD COLUMN cut_in_id     INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE;
ALTER TABLE player ADD COLUMN title_id      INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE;
ALTER TABLE player ADD COLUMN poster0_id    INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE;
ALTER TABLE player ADD COLUMN poster1_id    INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE;

CREATE TABLE character (
    character_id   INTEGER PRIMARY KEY AUTOINCREMENT,
    player_id      INTEGER REFERENCES player(player_id) ON DELETE CASCADE NOT NULL,
    item_id        INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE NOT NULL,
    hair_color     INTEGER NOT NULL,
	shirt          INTEGER NOT NULL,
    mastery        INTEGER NOT NULL,

    -- Part IDs for equipped clothing/etc.
    part00_item_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE,
    part01_item_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE,
    part02_item_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE,
    part03_item_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE,
    part04_item_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE,
    part05_item_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE,
    part06_item_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE,
    part07_item_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE,
    part08_item_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE,
    part09_item_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE,
    part10_item_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE,
    part11_item_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE,
    part12_item_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE,
    part13_item_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE,
    part14_item_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE,
    part15_item_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE,
    part16_item_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE,
    part17_item_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE,
    part18_item_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE,
    part19_item_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE,
    part20_item_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE,
    part21_item_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE,
    part22_item_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE,
    part23_item_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE,

    -- Part IFF IDs for equipped clothing/etc.
    -- n.b.: This may be non-zero even though there is no underlying item.
    part00_item_type_id INTEGER NOT NULL DEFAULT 0,
    part01_item_type_id INTEGER NOT NULL DEFAULT 0,
    part02_item_type_id INTEGER NOT NULL DEFAULT 0,
    part03_item_type_id INTEGER NOT NULL DEFAULT 0,
    part04_item_type_id INTEGER NOT NULL DEFAULT 0,
    part05_item_type_id INTEGER NOT NULL DEFAULT 0,
    part06_item_type_id INTEGER NOT NULL DEFAULT 0,
    part07_item_type_id INTEGER NOT NULL DEFAULT 0,
    part08_item_type_id INTEGER NOT NULL DEFAULT 0,
    part09_item_type_id INTEGER NOT NULL DEFAULT 0,
    part10_item_type_id INTEGER NOT NULL DEFAULT 0,
    part11_item_type_id INTEGER NOT NULL DEFAULT 0,
    part12_item_type_id INTEGER NOT NULL DEFAULT 0,
    part13_item_type_id INTEGER NOT NULL DEFAULT 0,
    part14_item_type_id INTEGER NOT NULL DEFAULT 0,
    part15_item_type_id INTEGER NOT NULL DEFAULT 0,
    part16_item_type_id INTEGER NOT NULL DEFAULT 0,
    part17_item_type_id INTEGER NOT NULL DEFAULT 0,
    part18_item_type_id INTEGER NOT NULL DEFAULT 0,
    part19_item_type_id INTEGER NOT NULL DEFAULT 0,
    part20_item_type_id INTEGER NOT NULL DEFAULT 0,
    part21_item_type_id INTEGER NOT NULL DEFAULT 0,
    part22_item_type_id INTEGER NOT NULL DEFAULT 0,
    part23_item_type_id INTEGER NOT NULL DEFAULT 0,

    aux_part0_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE,
    aux_part1_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE,
    aux_part2_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE,
    aux_part3_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE,
    aux_part4_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE,

    cut_in_id INTEGER REFERENCES inventory(item_id) ON DELETE CASCADE
);

ALTER TABLE player ADD COLUMN character_id  INTEGER REFERENCES character(character_id) ON DELETE CASCADE;

CREATE TABLE session (
    session_id INTEGER PRIMARY KEY AUTOINCREMENT,
    player_id INTEGER REFERENCES player(player_id) ON DELETE CASCADE NOT NULL,
    session_key TEXT NOT NULL UNIQUE,
    session_address TEXT NOT NULL,
    session_expires_at INTEGER NOT NULL
);

CREATE UNIQUE INDEX `session_key_idx` ON session (session_key);

-- +goose Down
DROP INDEX session_key_idx;
DROP TABLE session;
DROP TABLE character;
DROP TABLE inventory;
DROP TABLE player;

