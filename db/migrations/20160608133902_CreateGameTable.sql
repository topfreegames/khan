-- khan
-- https://github.com/topfreegames/khan
--
-- Licensed under the MIT license:
-- http://www.opensource.org/licenses/mit-license
-- Copyright © 2016 Top Free Games <backend@tfgco.com>

-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE games (
    id serial PRIMARY KEY,
    public_id varchar(36) NOT NULL,
    name varchar(2000) NOT NULL,
    min_membership_level integer NOT NULL,
    max_membership_level integer NOT NULL,
    min_level_to_accept_application integer NOT NULL,
    min_level_to_create_invitation integer NOT NULL,
    min_level_offset_to_promote_member integer NOT NULL,
    min_level_offset_to_demote_member integer NOT NULL,
    max_members integer NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at bigint NOT NULL,
    updated_at bigint NULL,

    CONSTRAINT public_id UNIQUE(public_id)
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE games;
