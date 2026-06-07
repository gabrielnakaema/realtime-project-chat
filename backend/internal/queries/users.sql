-- name: ListUsers :many
SELECT DISTINCT u.id, u.name, u.email, u.created_at
FROM users u
JOIN project_members pm ON pm.user_id = u.id
WHERE pm.project_id IN (SELECT project_id FROM project_members pm2 WHERE pm2.user_id = $1)
AND u.id != $1
ORDER BY u.name ASC;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserById :one
SELECT * FROM users WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (name, email, password) VALUES ($1, $2, $3) returning id;

-- name: UpdateUserPassword :execrows
UPDATE users
SET password = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token, expires_at, active) VALUES ($1, $2, $3, $4) returning id;

-- name: GetRefreshTokenByToken :one
SELECT * FROM refresh_tokens WHERE token = $1;

-- name: UpdateRefreshToken :exec
UPDATE refresh_tokens SET active = $1 WHERE token = $2;
