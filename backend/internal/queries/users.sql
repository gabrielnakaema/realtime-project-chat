-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserById :one
SELECT * FROM users WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (name, email, password) VALUES ($1, $2, $3) returning id;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token, expires_at, active) VALUES ($1, $2, $3, $4) returning id;

-- name: GetRefreshTokenByToken :one
SELECT * FROM refresh_tokens WHERE token = $1;

-- name: UpdateRefreshToken :exec
UPDATE refresh_tokens SET active = $1 WHERE token = $2;
