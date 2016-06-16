-- khan
-- https://github.com/topfreegames/khan
--
-- Licensed under the MIT license:
-- http://www.opensource.org/licenses/mit-license
-- Copyright © 2016 Top Free Games <backend@tfgco.com>

-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE clans (
    id serial PRIMARY KEY,
    public_id varchar(255) NOT NULL,
    game_id varchar(36) NOT NULL REFERENCES games (public_id),
    name varchar(2000) NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::JSONB,
    allow_application boolean NOT NULL DEFAULT true,
    auto_join boolean NOT NULL DEFAULT false,
    created_at bigint NOT NULL,
    updated_at bigint NULL,
    deleted_at bigint NULL,
    owner_id integer NOT NULL REFERENCES players (id),

    CONSTRAINT gameid_clanid UNIQUE(game_id, public_id)
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE clans;
