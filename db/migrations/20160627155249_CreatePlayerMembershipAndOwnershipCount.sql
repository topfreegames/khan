
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE players ADD COLUMN membership_count integer NOT NULL DEFAULT 0;
ALTER TABLE players ADD COLUMN ownership_count integer NOT NULL DEFAULT 0;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE players DROP COLUMN membership_count;
ALTER TABLE players DROP COLUMN ownership_count;
