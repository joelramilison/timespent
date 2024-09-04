-- name: CreateActivity :one
INSERT INTO activities(
    id, created_at, updated_at, name, user_id, color_code
)
VALUES (
    $1, NOW(), NOW(), $2, $3, '000000'
)
RETURNING *;

-- name: GetActivity :one
SELECT * FROM activities
WHERE id = $1;

-- name: GetUserActivities :many
SELECT * FROM activities
WHERE user_id = $1;

-- name: DeleteActivity :exec
DELETE FROM activities
WHERE id = $1;