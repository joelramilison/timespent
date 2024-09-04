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
    user_id,
    activity_id
)
VALUES (
    $1, NOW(), NOW(), NOW(), $2, $3
);

-- name: StopSession :exec
UPDATE sessions
SET updated_at = NOW(), ended_at = NOW(), pause_seconds = $1, paused_at = NULL, assign_to_day_before_start = $2
WHERE id = $3;

-- name: PauseSession :exec
UPDATE sessions
SET updated_at = NOW(), paused_at = $1
WHERE id = $2;

-- name: ResumeSession :exec
UPDATE sessions
SET updated_at = NOW(), paused_at = NULL, pause_seconds = $1
WHERE id = $2;

-- name: UpdateDayReassign :exec
UPDATE sessions
SET assign_to_day_before_start = $1, created_at = NOW()
WHERE id = $2;