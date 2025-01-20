-- name: CreateUser :exec
INSERT INTO users (id, username, password, nickname, avatar_url, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetUserByID :one
SELECT id, username, password, nickname, avatar_url, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT id, username, password, nickname, avatar_url, created_at, updated_at
FROM users
WHERE username = $1;

-- name: UpdateUser :exec
UPDATE users
SET username = $2,
    nickname = $3,
    avatar_url = $4,
    updated_at = $5
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: CheckUsernameExists :one
SELECT EXISTS(
    SELECT 1
    FROM users
    WHERE username = $1
) AS exists;

-- name: CheckNicknameExists :one
SELECT EXISTS(
    SELECT 1
    FROM users
    WHERE nickname = $1
) AS exists; 