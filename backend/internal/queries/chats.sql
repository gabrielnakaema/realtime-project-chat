-- name: CreateChat :one
INSERT INTO chats (project_id) VALUES ($1) returning id;

-- name: CreateChatMember :exec
INSERT INTO chat_members (user_id, chat_id, last_seen_at, joined_at) VALUES ($1, $2, $3, $4);

-- name: UpdateChatMemberLastSeenAt :exec
UPDATE chat_members SET last_seen_at = $1 WHERE user_id = $2 AND chat_id = $3;

-- name: CreateChatMessage :exec
INSERT INTO chat_messages (chat_id, user_id, content, created_at, updated_at, message_type) VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetChatById :one
SELECT * FROM chats WHERE id = $1;

-- name: GetChatByProjectId :one
SELECT * FROM chats WHERE project_id = $1;