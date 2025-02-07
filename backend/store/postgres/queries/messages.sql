-- name: CreateMessage :exec
INSERT INTO messages (id, type, content, user_id, session_id, timestamp)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetMessageIDsBySessionID :many
SELECT id
FROM messages
WHERE session_id = $1
  AND timestamp < $2
ORDER BY timestamp DESC
LIMIT $3;

-- name: GetMessageByID :one
SELECT id, type, content, user_id, session_id, timestamp
FROM messages
WHERE id = $1;

-- name: DeleteMessage :exec
DELETE FROM messages
WHERE id = $1;

-- name: GetMessagesByIDs :many
SELECT id, type, content, user_id, session_id, timestamp
FROM messages
WHERE id = ANY($1); 