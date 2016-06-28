
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE clans ADD COLUMN membership_count integer NOT NULL DEFAULT 1;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE clans DROP COLUMN membership_count;
