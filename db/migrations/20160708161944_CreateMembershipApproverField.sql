
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE memberships ADD COLUMN approver_id integer NULL REFERENCES players (id);
ALTER TABLE memberships ADD COLUMN approved_at bigint NULL;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE memberships DROP COLUMN approver_id;
ALTER TABLE memberships DROP COLUMN approved_at;
