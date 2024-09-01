-- +goose Up
ALTER TABLE sessions 
ADD COLUMN assign_to_day_before_start boolean;
-- +goose Down

ALTER TABLE sessions
DROP COLUMN assign_to_day_before_start;