
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE EXTENSION IF NOT EXISTS unaccent;

ALTER TABLE clans
ADD COLUMN IF NOT EXISTS searchable tsvector GENERATED ALWAYS AS (
    setweight(to_tsvector('portuguese', coalesce(name, '')), 'A')) STORED;

CREATE INDEX IF NOT EXISTS IDX_CLAN_TS_VECTOR ON clans USING GIN(searchable) ;