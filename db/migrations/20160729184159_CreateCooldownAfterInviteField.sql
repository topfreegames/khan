
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE games ADD COLUMN cooldown_before_invite integer NOT NULL DEFAULT 0;
ALTER TABLE games ADD COLUMN cooldown_before_apply integer NOT NULL DEFAULT 3600;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE games DROP COLUMN cooldown_before_invite;
ALTER TABLE games DROP COLUMN cooldown_before_apply;
