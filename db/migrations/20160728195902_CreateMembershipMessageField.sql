
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE memberships ADD COLUMN message varchar(255) NULL;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE memberships DROP COLUMN message;
