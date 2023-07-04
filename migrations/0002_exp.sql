-- +goose Up
ALTER TABLE player ADD COLUMN exp INTEGER NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE player DROP COLUMN exp;
