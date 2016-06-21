-- khan
-- https://github.com/topfreegames/khan
--
-- Licensed under the MIT license:
-- http://www.opensource.org/licenses/mit-license
-- Copyright © 2016 Top Free Games <backend@tfgco.com>

-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE hooks (
    id serial PRIMARY KEY,
    game_id varchar(36) NOT NULL REFERENCES games (public_id),
    public_id varchar(36) NOT NULL,
    event_type integer NOT NULL,
    url text NOT NULL,
    created_at bigint NOT NULL,
    updated_at bigint NULL,

    CONSTRAINT hookid_publicid UNIQUE(game_id, public_id)
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE hooks;
