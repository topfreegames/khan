-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE players (
    id serial PRIMARY KEY,
    player_id varchar(255) NOT NULL,
    game_id varchar(10) NOT NULL,
    name varchar(2000) NOT NULL,
    created_at timestamp NOT NULL,
    updated_at timestamp NULL,
    deleted_at timestamp NULL,

    CONSTRAINT gameid_playerid UNIQUE(game_id, player_id)
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE players;
