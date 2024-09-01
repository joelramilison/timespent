-- name: CreateUser :exec
INSERT INTO users (id, username, created_at, updated_at, password_hash, time_zone, session_id_hash, session_expires_at)
VALUES (
    $1, $2, NOW(), NOW(), $3, $4, $5, $6
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserbyName :one
SELECT * FROM users
WHERE username = $1;

-- name: UpdateLoginSession :exec
UPDATE users
SET updated_at = NOW(), session_expires_at = $1, session_id_hash = $2
WHERE id = $3;

-- name: LogUserOut :exec
UPDATE users
SET updated_at = NOW(), session_expires_at = NULL, session_id_hash = NULL
WHERE id = $1;

-- name: UpdateAssignAwait :exec
UPDATE users
SET updated_at = NOW(), await_assign_decision_until = $1
WHERE id = $2;