-- name: CreateNotification :one
INSERT INTO notifications (
  user_id,
  actor_id,
  project_id,
  task_id,
  task_comment_id,
  type,
  read_at,
  created_at,
  updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id;

-- name: ListNotificationsByUserId :many
WITH notification_rows AS (
  SELECT
    n.id,
    n.user_id,
    n.actor_id,
    n.project_id,
    n.task_id,
    n.task_comment_id,
    n.type,
    n.read_at,
    n.created_at,
    n.updated_at
  FROM notifications n
  WHERE n.user_id = $1
    AND (
      sqlc.narg('before_created_at')::timestamptz IS NULL
      OR (n.created_at, n.id) < (sqlc.narg('before_created_at')::timestamptz, sqlc.narg('before_id')::uuid)
    )
  ORDER BY n.created_at DESC, n.id DESC
  LIMIT (COALESCE(sqlc.narg('limit')::integer, 0) + 1)
)
SELECT
  nr.id,
  nr.user_id,
  nr.actor_id,
  nr.project_id,
  nr.task_id,
  nr.task_comment_id,
  nr.type,
  nr.read_at,
  nr.created_at,
  nr.updated_at,
  jsonb_build_object(
    'id', actor.id,
    'name', actor.name,
    'email', actor.email,
    'created_at', actor.created_at,
    'updated_at', actor.updated_at
  ) AS actor,
  jsonb_build_object(
    'id', p.id,
    'user_id', p.user_id,
    'name', p.name,
    'description', p.description,
    'created_at', p.created_at,
    'updated_at', p.updated_at
  ) AS project,
  CASE
    WHEN t.id IS NULL THEN NULL
    ELSE jsonb_build_object(
      'id', t.id,
      'project_id', t.project_id,
      'author_id', t.author_id,
      'title', t.title,
      'description', t.description,
      'status', t.status,
      'priority', t.priority,
      'order', t.task_order,
      'responsible_id', t.responsible_id,
      'due_date', t.due_date,
      'done_at', t.done_at,
      'created_at', t.created_at,
      'updated_at', t.updated_at,
      'tags', coalesce(task_tags.tags, '[]'::jsonb),
      'author', CASE
        WHEN author.id IS NULL THEN NULL
        ELSE jsonb_build_object(
          'id', author.id,
          'name', author.name,
          'email', author.email,
          'created_at', author.created_at,
          'updated_at', author.updated_at
        )
      END,
      'responsible', CASE
        WHEN responsible.id IS NULL THEN NULL
        ELSE jsonb_build_object(
          'id', responsible.id,
          'name', responsible.name,
          'email', responsible.email,
          'created_at', responsible.created_at,
          'updated_at', responsible.updated_at
        )
      END
    )
  END AS task,
  CASE
    WHEN tc.id IS NULL THEN NULL
    ELSE jsonb_build_object(
      'id', tc.id,
      'content', tc.content,
      'parent_comment_id', tc.parent_comment_id,
      'created_at', tc.created_at,
      'updated_at', tc.updated_at
    )
  END AS task_comment
FROM notification_rows nr
JOIN users actor ON actor.id = nr.actor_id
JOIN projects p ON p.id = nr.project_id
LEFT JOIN tasks t ON t.id = nr.task_id
LEFT JOIN users author ON author.id = t.author_id
LEFT JOIN users responsible ON responsible.id = t.responsible_id
LEFT JOIN LATERAL (
  SELECT jsonb_agg(tt.name ORDER BY tt.name) AS tags
  FROM task_tags tt
  WHERE tt.task_id = t.id
) task_tags ON true
LEFT JOIN task_comments tc ON tc.id = nr.task_comment_id
ORDER BY nr.created_at DESC, nr.id DESC;

-- name: ListNotificationsByIds :many
WITH notification_rows AS (
  SELECT
    n.id,
    n.user_id,
    n.actor_id,
    n.project_id,
    n.task_id,
    n.task_comment_id,
    n.type,
    n.read_at,
    n.created_at,
    n.updated_at
  FROM notifications n
  WHERE n.id = ANY($1::uuid[])
)
SELECT
  nr.id,
  nr.user_id,
  nr.actor_id,
  nr.project_id,
  nr.task_id,
  nr.task_comment_id,
  nr.type,
  nr.read_at,
  nr.created_at,
  nr.updated_at,
  jsonb_build_object(
    'id', actor.id,
    'name', actor.name,
    'email', actor.email,
    'created_at', actor.created_at,
    'updated_at', actor.updated_at
  ) AS actor,
  jsonb_build_object(
    'id', p.id,
    'user_id', p.user_id,
    'name', p.name,
    'description', p.description,
    'created_at', p.created_at,
    'updated_at', p.updated_at
  ) AS project,
  CASE
    WHEN t.id IS NULL THEN NULL
    ELSE jsonb_build_object(
      'id', t.id,
      'project_id', t.project_id,
      'author_id', t.author_id,
      'title', t.title,
      'description', t.description,
      'status', t.status,
      'priority', t.priority,
      'order', t.task_order,
      'responsible_id', t.responsible_id,
      'due_date', t.due_date,
      'done_at', t.done_at,
      'created_at', t.created_at,
      'updated_at', t.updated_at,
      'tags', coalesce(task_tags.tags, '[]'::jsonb),
      'author', CASE
        WHEN author.id IS NULL THEN NULL
        ELSE jsonb_build_object(
          'id', author.id,
          'name', author.name,
          'email', author.email,
          'created_at', author.created_at,
          'updated_at', author.updated_at
        )
      END,
      'responsible', CASE
        WHEN responsible.id IS NULL THEN NULL
        ELSE jsonb_build_object(
          'id', responsible.id,
          'name', responsible.name,
          'email', responsible.email,
          'created_at', responsible.created_at,
          'updated_at', responsible.updated_at
        )
      END
    )
  END AS task,
  CASE
    WHEN tc.id IS NULL THEN NULL
    ELSE jsonb_build_object(
      'id', tc.id,
      'content', tc.content,
      'parent_comment_id', tc.parent_comment_id,
      'created_at', tc.created_at,
      'updated_at', tc.updated_at
    )
  END AS task_comment
FROM notification_rows nr
JOIN users actor ON actor.id = nr.actor_id
JOIN projects p ON p.id = nr.project_id
LEFT JOIN tasks t ON t.id = nr.task_id
LEFT JOIN users author ON author.id = t.author_id
LEFT JOIN users responsible ON responsible.id = t.responsible_id
LEFT JOIN LATERAL (
  SELECT jsonb_agg(tt.name ORDER BY tt.name) AS tags
  FROM task_tags tt
  WHERE tt.task_id = t.id
) task_tags ON true
LEFT JOIN task_comments tc ON tc.id = nr.task_comment_id
ORDER BY nr.created_at DESC, nr.id DESC;

-- name: CountUnreadNotificationsByUserId :one
SELECT COUNT(*)::int AS unread_count
FROM notifications
WHERE user_id = $1
  AND read_at IS NULL;

-- name: MarkNotificationRead :execrows
UPDATE notifications
SET read_at = COALESCE(read_at, $3), updated_at = $3
WHERE id = $1
  AND user_id = $2;

-- name: MarkAllNotificationsRead :exec
UPDATE notifications
SET read_at = COALESCE(read_at, $2), updated_at = $2
WHERE user_id = $1
  AND read_at IS NULL;
