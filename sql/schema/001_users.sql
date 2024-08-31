-- +goose Up
CREATE TABLE users (
    id uuid PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    username TEXT NOT NULL UNIQUE,
    password_hash BYTEA NOT NULL,
    time_zone TEXT NOT NULL,
    session_id_hash BYTEA,
    session_expires_at TIMESTAMP WITH TIME ZONE
);

-- +goose Down

DROP TABLE users;