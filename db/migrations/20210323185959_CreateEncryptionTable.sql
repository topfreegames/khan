
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE players_encrypteds (
    id integer PRIMARY KEY REFERENCES players (id)
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE players_encrypteds;
