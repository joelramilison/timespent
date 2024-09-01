-- +goose Up
ALTER TABLE users 
ADD COLUMN await_assign_decision_until TIMESTAMP WITH TIME ZONE;
-- +goose Down

ALTER TABLE users
DROP COLUMN await_assign_decision_until;