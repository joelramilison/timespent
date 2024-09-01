-- +goose Up
CREATE TABLE sessions (
    id uuid PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    ended_at TIMESTAMP WITH TIME ZONE,
    pause_seconds INTEGER NOT NULL DEFAULT 0,
    user_id uuid NOT NULL
);

-- +goose Down

DROP TABLE sessions;