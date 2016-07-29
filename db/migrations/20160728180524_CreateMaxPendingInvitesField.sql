
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE games ADD COLUMN max_pending_invites integer DEFAULT -1 NOT NULL;
UPDATE games SET max_pending_invites=-1;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE games DROP COLUMN max_pending_invites;
