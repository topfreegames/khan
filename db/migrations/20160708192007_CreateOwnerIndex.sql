
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE INDEX clans_owner_id ON clans(owner_id);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP INDEX IF EXISTS clans_owner_id;
