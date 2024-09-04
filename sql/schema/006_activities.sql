-- +goose Up
CREATE TABLE activities (
    id uuid PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    name TEXT NOT NULL,
    user_id uuid NOT NULL,
    color_code VARCHAR(6) NOT NULL
);

ALTER TABLE sessions
ADD COLUMN activity_id uuid REFERENCES activities(id);


-- +goose Down

DROP TABLE activities;

ALTER TABLE sessions
DROP COLUMN activity_id;