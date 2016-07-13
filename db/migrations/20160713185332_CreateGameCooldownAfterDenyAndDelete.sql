
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE games ADD COLUMN cooldown_after_deny integer NOT NULL DEFAULT 0;
ALTER TABLE games ADD COLUMN cooldown_after_delete integer NOT NULL DEFAULT 0;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE games DROP COLUMN cooldown_after_deny;
ALTER TABLE games DROP COLUMN cooldown_after_delete;
