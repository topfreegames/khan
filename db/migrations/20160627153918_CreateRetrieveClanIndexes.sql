
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE INDEX memberships_approved ON memberships(approved);
CREATE INDEX memberships_banned ON memberships(banned);
CREATE INDEX memberships_denied ON memberships(denied);
CREATE INDEX memberships_pending ON memberships(approved, banned, denied);
CREATE INDEX memberships_clan ON memberships(clan_id);


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP INDEX IF EXISTS memberships_approved;
DROP INDEX IF EXISTS memberships_banned;
DROP INDEX IF EXISTS memberships_denied;
DROP INDEX IF EXISTS memberships_pending;
DROP INDEX IF EXISTS memberships_clan;
