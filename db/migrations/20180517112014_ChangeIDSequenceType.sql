-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE players ALTER COLUMN id TYPE BIGINT;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE players ALTER COLUMN id TYPE INTEGER;
