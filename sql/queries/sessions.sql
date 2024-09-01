-- name: GetNewestSession :one
SELECT * FROM sessions
WHERE user_id = $1
ORDER BY started_at DESC
LIMIT 1;

-- name: StartSession :exec
INSERT INTO sessions(
    id,
    created_at,
    updated_at,
    started_at,
    user_id
)
VALUES (
    $1, NOW(), NOW(), NOW(), $2
);