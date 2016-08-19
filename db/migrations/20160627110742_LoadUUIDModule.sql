
-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";


-- +goose Down
DROP EXTENSION IF EXISTS "uuid-ossp";
