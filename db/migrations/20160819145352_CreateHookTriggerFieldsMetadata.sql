-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE games ADD COLUMN clan_metadata_fields_whitelist varchar(2000) DEFAULT '';
ALTER TABLE games ADD COLUMN player_metadata_fields_whitelist varchar(2000) DEFAULT '';

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE games DROP COLUMN clan_metadata_fields_whitelist;
ALTER TABLE games DROP COLUMN player_metadata_fields_whitelist;
