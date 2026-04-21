-- name: CreateChat :one
INSERT INTO chats (project_id) VALUES ($1) returning id;

-- name: CreateChatMember :exec
INSERT INTO chat_members (user_id, chat_id, last_seen_at, joined_at) VALUES ($1, $2, $3, $4);

-- name: UpdateChatMemberLastSeenAt :exec
UPDATE chat_members
SET last_seen_at = GREATEST(last_seen_at, $1)
WHERE user_id = $2 AND chat_id = $3;

-- name: CreateChatMessage :one
INSERT INTO chat_messages (chat_id, user_id, content, created_at, updated_at, message_type) VALUES ($1, $2, $3, $4, $5, $6) returning id;

-- name: UpdateChatUpdatedAt :exec
UPDATE chats SET updated_at = $1 WHERE id = $2;

-- name: GetChatById :one
with chat_members_cte as (
	select 
		cm.chat_id as member_chat_id,
		cm.user_id as member_user_id,
		cm.joined_at as member_joined_at,
		cm.last_seen_at as member_last_seen_at,
		u."name" as member_name
	from chat_members cm
	left join users u on u.id = cm.user_id
)
select 
	c.*,
	coalesce(
		jsonb_agg(
			jsonb_build_object(
				'chat_id', cm.member_chat_id,
				'user_id', cm.member_user_id,
				'last_seen_at', cm.member_last_seen_at,
				'joined_at', cm.member_joined_at,
				'user',
				jsonb_build_object(
					'id', cm.member_user_id,
					'name', cm.member_name
				)
			)
		) filter (where cm.member_user_id is not null and cm.member_chat_id is not null)
	, '[]'::jsonb) as members
from chats c
left join chat_members_cte cm on cm.member_chat_id = c.id
where c.id = $1
group by c.id;

-- name: GetChatByProjectId :one
with chat_members_cte as (
	select 
		cm.chat_id as member_chat_id,
		cm.user_id as member_user_id,
		cm.joined_at as member_joined_at,
		cm.last_seen_at as member_last_seen_at,
		u."name" as member_name
	from chat_members cm
	left join users u on u.id = cm.user_id
)
select 
	c.*,
	coalesce(
		jsonb_agg(
			jsonb_build_object(
				'chat_id', cm.member_chat_id,
				'user_id', cm.member_user_id,
				'last_seen_at', cm.member_last_seen_at,
				'joined_at', cm.member_joined_at,
				'user',
				jsonb_build_object(
					'id', cm.member_user_id,
					'name', cm.member_name
				)
			)
		) filter (where cm.member_user_id is not null and cm.member_chat_id is not null)
	, '[]'::jsonb) as members
from chats c
left join chat_members_cte cm on cm.member_chat_id = c.id
where c.project_id = $1
group by c.id;

-- name: CreateGeneralChat :one
INSERT INTO chats (chat_type) VALUES ('general') RETURNING id;

-- name: FindGeneralChatByExactMembers :one
SELECT c.id FROM chats c
WHERE c.chat_type = 'general'
  AND (SELECT COUNT(*) FROM chat_members WHERE chat_id = c.id) = cardinality($1::uuid[])
  AND NOT EXISTS (
    SELECT 1 FROM chat_members cm
    WHERE cm.chat_id = c.id AND cm.user_id != ALL($1::uuid[])
  )
LIMIT 1;

-- name: ListGeneralChatsByUserId :many
WITH chat_members_cte AS (
	SELECT
		cm.chat_id AS member_chat_id,
		u.id AS member_user_id,
		cm.joined_at AS member_joined_at,
		cm.last_seen_at AS member_last_seen_at,
		u.name AS member_name,
		u.email AS member_email
	FROM chat_members cm
	JOIN users u ON u.id = cm.user_id
)
SELECT
	c.id,
	c.project_id,
	c.chat_type,
	c.created_at,
	c.updated_at,
	coalesce(unread_stats.unread_count, 0)::int as unread_count,
	coalesce(unread_stats.has_more_unread, false) as has_more_unread,
	coalesce(
		jsonb_agg(
			jsonb_build_object(
				'chat_id', cm.member_chat_id,
				'user_id', cm.member_user_id,
				'last_seen_at', cm.member_last_seen_at,
				'joined_at', cm.member_joined_at,
				'user',
				jsonb_build_object(
					'id', cm.member_user_id,
					'name', cm.member_name,
					'email', cm.member_email
				)
			)
		) FILTER (WHERE cm.member_user_id IS NOT NULL)
	, '[]'::jsonb) AS members
FROM chats c
JOIN chat_members cm_self ON cm_self.chat_id = c.id AND cm_self.user_id = $1
LEFT JOIN chat_members_cte cm ON cm.member_chat_id = c.id
LEFT JOIN LATERAL (
	SELECT
		LEAST(COUNT(*), $2 - 1)::int AS unread_count,
		COUNT(*) = $2 AS has_more_unread
	FROM (
		SELECT 1
		FROM chat_messages chat_message
		WHERE chat_message.chat_id = c.id
			AND chat_message.user_id IS DISTINCT FROM $1
			AND (chat_message.created_at, chat_message.id) > (
				cm_self.last_seen_at,
				'00000000-0000-0000-0000-000000000000'::uuid
			)
		ORDER BY chat_message.created_at DESC, chat_message.id DESC
		LIMIT $2
	) AS unread_messages
) unread_stats ON true
WHERE c.chat_type = 'general'
GROUP BY c.id, unread_stats.unread_count, unread_stats.has_more_unread
ORDER BY c.updated_at DESC;

-- name: ListChatMessages :many
select
	cm.id,
	cm.chat_id,
	cm.content,
	cm.created_at,
	cm.updated_at,
	cm.user_id,
	cm.message_type,
	u.name as user_name,
	coalesce(count(cmr.user_id), 0)::int as reads_count
from chat_messages cm
left join users u on u.id = cm.user_id
left join chat_message_reads cmr on cmr.chat_message_id = cm.id
where cm.chat_id = $1
and (cm.created_at, cm.id) < ($2, $3::uuid)
group by cm.id, u.name
order by cm.created_at desc, cm.id desc
limit $4;

-- name: GetUnreadCountByChatId :one
WITH unread_messages AS (
	SELECT 1
	FROM chat_messages cm
	JOIN chat_members cmm ON cmm.chat_id = cm.chat_id AND cmm.user_id = $2
	WHERE cm.chat_id = $1
		AND cm.user_id IS DISTINCT FROM $2
		AND (cm.created_at, cm.id) > (cmm.last_seen_at, '00000000-0000-0000-0000-000000000000'::uuid)
	ORDER BY cm.created_at DESC, cm.id DESC
	LIMIT $3
)
SELECT
	LEAST(COUNT(*), $3 - 1)::int AS unread_count,
	COUNT(*) = $3 AS has_more_unread
FROM unread_messages;

-- name: GetChatMessageById :one
SELECT id, chat_id, content, created_at, updated_at, user_id, message_type
FROM chat_messages
WHERE id = $1;

-- name: UpsertChatMessageReadsUpTo :exec
INSERT INTO chat_message_reads (chat_message_id, user_id, read_at)
SELECT cm.id, $2, $3
FROM chat_messages cm
WHERE cm.chat_id = $1
	AND (cm.created_at, cm.id) <= ($4, $5::uuid)
	AND cm.user_id IS DISTINCT FROM $2
ON CONFLICT (chat_message_id, user_id)
DO UPDATE SET read_at = GREATEST(chat_message_reads.read_at, EXCLUDED.read_at);

-- name: ListChatMessageReads :many
SELECT
	cmr.chat_message_id,
	cmr.user_id,
	cmr.read_at,
	u.name,
	u.email
FROM chat_message_reads cmr
JOIN users u ON u.id = cmr.user_id
JOIN chat_messages cm ON cm.id = cmr.chat_message_id
WHERE cmr.chat_message_id = $1
	AND cm.user_id IS DISTINCT FROM cmr.user_id
ORDER BY cmr.read_at ASC, u.name ASC;
