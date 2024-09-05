-- +goose Up
ALTER TABLE sessions 
ADD COLUMN corresponding_date date;

ALTER TABLE sessions
ADD COLUMN started_at_local_date date NOT NULL;
-- +goose Down

ALTER TABLE sessions
DROP COLUMN corresponding_date;

ALTER TABLE sessions
DROP COLUMN started_at_local_date;