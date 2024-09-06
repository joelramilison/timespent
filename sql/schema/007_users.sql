-- +goose Up
ALTER TABLE users
DROP COLUMN time_zone;

-- +goose Down
ALTER TABLE users
ADD COLUMN time_zone TEXT NOT NULL;

