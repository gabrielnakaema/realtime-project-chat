-- name: CreateTaskComment :one
INSERT INTO task_comments (task_id, user_id, content, parent_comment_id) VALUES ($1, $2, $3, $4) returning id;

-- name: ListTaskComments :many
WITH recursive paginated_parent_comments AS (
  SELECT
    c.id,
    c.task_id,
    c.user_id,
    c.content,
    c.parent_comment_id,
    c.created_at,
    c.updated_at
  FROM task_comments c
  WHERE c.task_id = $1
    AND c.parent_comment_id IS NULL
    AND (c.created_at, c.id) < ($3::timestamptz, $4::uuid)
  ORDER BY c.created_at DESC, c.id DESC
  LIMIT $2
),
comment_tree AS (
  SELECT
    p.id,
    p.task_id,
    p.user_id,
    p.content,
    p.parent_comment_id,
    p.created_at,
    p.updated_at,
    0 as level,
    p.created_at as root_created_at,
    u.id as comment_user_id,
    u.name as comment_user_name,
    u.email as comment_user_email,
    u.created_at as comment_user_created_at,
    u.updated_at as comment_user_updated_at
  FROM paginated_parent_comments p
  JOIN users u ON p.user_id = u.id
  UNION ALL
  SELECT
    c.id,
    c.task_id,
    c.user_id,
    c.content,
    c.parent_comment_id,
    c.created_at,
    c.updated_at,
    ct.level + 1 as level,
    ct.root_created_at,
    u.id as comment_user_id,
    u.name as comment_user_name,
    u.email as comment_user_email,
    u.created_at as comment_user_created_at,
    u.updated_at as comment_user_updated_at
  FROM task_comments c
  JOIN users u ON c.user_id = u.id
  JOIN comment_tree ct ON c.parent_comment_id = ct.id
)
SELECT
  *
FROM comment_tree
ORDER BY root_created_at DESC, created_at ASC, id ASC;