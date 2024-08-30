-- +goose Up
CREATE TABLE users (
    id uuid PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    password_hash BYTEA NOT NULL,
    time_zone TEXT NOT NULL
);

CREATE TABLE refresh_tokens (
    user_id uuid UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    refresh_token BYTEA,
    expires_at TIMESTAMP WITH TIME ZONE
);

-- +goose Down
DROP TABLE users;
DROP TABLE refresh_tokens;