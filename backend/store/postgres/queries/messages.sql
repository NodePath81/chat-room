-- name: CreateMessage :exec
INSERT INTO messages (id, type, content, user_id, session_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetMessagesBySessionID :many
SELECT id, type, content, user_id, session_id, created_at, updated_at
FROM messages
WHERE session_id = $1
  AND created_at < $2
ORDER BY created_at DESC
LIMIT $3;

-- name: DeleteMessage :exec
DELETE FROM messages
WHERE id = $1; 