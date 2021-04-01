
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE encrypted_players ALTER COLUMN player_id TYPE BIGINT;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE encrypted_players ALTER COLUMN player_id TYPE INTEGER;

