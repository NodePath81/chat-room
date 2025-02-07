-- name: CreateSession :exec
INSERT INTO sessions (id, name, creator_id, created_at)
VALUES ($1, $2, $3, $4);

-- name: GetSessionByID :one
SELECT id, name, creator_id, created_at
FROM sessions
WHERE id = $1;

-- name: GetUserSessions :many
SELECT s.id, s.name, s.creator_id, s.created_at,
       us.role, us.joined_at
FROM sessions s
JOIN user_sessions us ON s.id = us.session_id
WHERE us.user_id = $1;

-- name: UpdateSession :exec
UPDATE sessions
SET name = $2
WHERE id = $1;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE id = $1;

-- name: AddUserToSession :exec
INSERT INTO user_sessions (user_id, session_id, role, joined_at)
VALUES ($1, $2, $3, $4);

-- name: RemoveUserFromSession :exec
DELETE FROM user_sessions
WHERE user_id = $1 AND session_id = $2;

-- name: GetSessionUsers :many
SELECT u.id, u.username, u.nickname, u.avatar_url, u.created_at
FROM users u
JOIN user_sessions us ON u.id = us.user_id
WHERE us.session_id = $1;

-- name: GetUserSessionRole :one
SELECT role
FROM user_sessions
WHERE user_id = $1 AND session_id = $2;

-- name: GetSessionIDsByUserID :many
SELECT session_id
FROM user_sessions
WHERE user_id = $1;

-- name: GetUserIDsBySessionID :many
SELECT user_id
FROM user_sessions
WHERE session_id = $1;

-- name: GetSessionsByIDs :many
SELECT id, name, creator_id, created_at
FROM sessions
WHERE id = ANY($1); 