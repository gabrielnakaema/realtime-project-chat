-- name: CreateChat :one
INSERT INTO chats (project_id) VALUES ($1) returning id;

-- name: CreateChatMember :exec
INSERT INTO chat_members (user_id, chat_id, last_seen_at, joined_at) VALUES ($1, $2, $3, $4);

-- name: UpdateChatMemberLastSeenAt :exec
UPDATE chat_members SET last_seen_at = $1 WHERE user_id = $2 AND chat_id = $3;

-- name: CreateChatMessage :one
INSERT INTO chat_messages (chat_id, user_id, content, created_at, updated_at, message_type) VALUES ($1, $2, $3, $4, $5, $6) returning id;

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

-- name: ListChatMessages :many
select 
	cm.id,
	cm.chat_id,
	cm.content,
	cm.created_at,
	cm.updated_at,
	cm.user_id,
	cm.message_type,
	u.name as user_name
from chat_messages cm
left join users u on u.id = cm.user_id
where cm.chat_id = $1
and (cm.created_at, cm.id) < ($2, $3::uuid)
order by cm.created_at desc, cm.id desc
limit $4;