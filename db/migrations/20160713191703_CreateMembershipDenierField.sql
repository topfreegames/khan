
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE memberships ADD COLUMN denier_id integer NULL REFERENCES players (id);
ALTER TABLE memberships ADD COLUMN denied_at bigint NULL;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE memberships DROP COLUMN denier_id;
ALTER TABLE memberships DROP COLUMN denied_at;
