-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE memberships (
    id serial PRIMARY KEY,
    game_id varchar(10) NOT NULL,
    clan_id integer NOT NULL,
    player_id integer NOT NULL,
    membership_level integer NOT NULL,
    approved boolean NOT NULL DEFAULT false,
    denied boolean NOT NULL DEFAULT false,
    created_at bigint NOT NULL,
    updated_at bigint NULL,
    deleted_at bigint NULL,

    CONSTRAINT playerid_clanid UNIQUE(player_id, clan_id)
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE memberships;
