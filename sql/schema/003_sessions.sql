-- +goose Up
ALTER TABLE sessions 
ADD COLUMN paused_at TIMESTAMP WITH TIME ZONE;
-- +goose Down

ALTER TABLE sessions
DROP COLUMN paused_at;