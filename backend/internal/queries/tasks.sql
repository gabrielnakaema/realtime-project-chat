-- name: CreateTask :one
INSERT INTO tasks (project_id, title, description, status, author_id, responsible_id, priority, due_date, task_order) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id;

-- name: CreateTaskTag :exec
INSERT INTO task_tags (task_id, name) VALUES ($1, $2);

-- name: DeleteTaskTag :exec
DELETE FROM task_tags WHERE task_id = $1 AND name = $2;

-- name: DeleteAllTaskTags :exec
DELETE FROM task_tags WHERE task_id = $1;

-- name: GetTaskById :one
WITH task_tags_cte AS (
  SELECT
    tt.name as task_tag_name
  from task_tags tt
  WHERE tt.task_id = $1
), task_changes_cte AS (
  SELECT
    tc.id as task_change_id,
    tc.update_id as task_change_update_id,
    tc.subject_id as task_change_subject_id,
    tc.old_value_id as task_change_old_value_id,
    tc.new_value_id as task_change_new_value_id,
    tc.field as task_change_field,
    tc.old_value as task_change_old_value,
    tc.new_value as task_change_new_value,
    tc.old_display_value as task_change_old_display_value,
    tc.new_display_value as task_change_new_display_value,
    tc.created_at as task_change_created_at,
    s.name as task_change_subject_name,
    s.email as task_change_subject_email,
    s.created_at as task_change_subject_created_at,
    old_r.name as task_change_old_value_name,
    new_r.name as task_change_new_value_name
   FROM task_changes tc
   LEFT JOIN users s ON s.id = tc.subject_id
   LEFT JOIN users old_r ON tc.field = 'responsible_id' AND old_r.id::text = nullif(tc.old_value, '') AND tc.old_display_value IS NULL
   LEFT JOIN users new_r ON tc.field = 'responsible_id' AND new_r.id::text = nullif(tc.new_value, '') AND tc.new_display_value IS NULL
), task_updates_cte AS (
  SELECT
    tu.id as task_update_id,
    tu.task_id as task_update_task_id,
    tu.user_id as task_update_user_id,
    tu.update_type as task_update_update_type,
    tu.created_at as task_update_created_at,
    u.name as task_update_user_name,
    u.email as task_update_user_email,
    u.created_at as task_update_user_created_at,
    coalesce(jsonb_agg(
      jsonb_build_object(
        'id', tc.task_change_id,
        'update_id', tc.task_change_update_id,
        'subject_id', tc.task_change_subject_id,
        'old_value_id', tc.task_change_old_value_id,
        'new_value_id', tc.task_change_new_value_id,
        'field', tc.task_change_field,
        'old_value', tc.task_change_old_value,
        'new_value', tc.task_change_new_value,
        'old_display_value', CASE
          WHEN tc.task_change_old_display_value IS NOT NULL THEN tc.task_change_old_display_value
          WHEN tc.task_change_field = 'responsible_id' THEN tc.task_change_old_value_name
          ELSE NULL
        END,
        'new_display_value', CASE
          WHEN tc.task_change_new_display_value IS NOT NULL THEN tc.task_change_new_display_value
          WHEN tc.task_change_field = 'responsible_id' THEN tc.task_change_new_value_name
          ELSE NULL
        END,
        'created_at', tc.task_change_created_at,
        'subject', jsonb_build_object(
          'id', tc.task_change_subject_id,
          'name', tc.task_change_subject_name,
          'email', tc.task_change_subject_email,
          'created_at', tc.task_change_subject_created_at
        )
      )
      ORDER BY tc.task_change_created_at DESC, tc.task_change_id DESC
    ) filter (where tc.task_change_id is not null), '[]'::jsonb) as task_changes
   FROM task_updates tu
   LEFT JOIN users u ON u.id = tu.user_id
   LEFT JOIN task_changes_cte tc ON tc.task_change_update_id = tu.id
   WHERE tu.task_id = $1 
   GROUP BY tu.id, u.name, u.email, u.created_at
   ORDER BY tu.created_at DESC
)
SELECT
  t.id as task_id,
  t.project_id as task_project_id,
  t.title as task_title,
  t.description as task_description,
  t.status as task_status,
  t.priority as task_priority,
  t.responsible_id as task_responsible_id,
  t.due_date as task_due_date,
  t.done_at as task_done_at,
  t.task_order as task_order,
  t.created_at as task_created_at,
  t.updated_at as task_updated_at,
  t.author_id as task_author_id,
  a.name as task_author_name,
  a.email as task_author_email,
  a.created_at as task_author_created_at,
  r.name as task_responsible_name,
  r.email as task_responsible_email,
  r.created_at as task_responsible_created_at,
  (select coalesce(jsonb_agg(tt.task_tag_name) filter (where tt.task_tag_name is not null), '[]') from task_tags_cte tt) as tags,
  (select coalesce(jsonb_agg(
    jsonb_build_object(
      'id', tu.task_update_id,
      'task_id', tu.task_update_task_id,
      'user_id', tu.task_update_user_id,
      'update_type', tu.task_update_update_type,
      'created_at', tu.task_update_created_at,
      'user', jsonb_build_object(
        'id', tu.task_update_user_id,
        'name', tu.task_update_user_name,
        'email', tu.task_update_user_email,
        'created_at', tu.task_update_user_created_at
      ),
      'changes', tu.task_changes
    )
    ORDER BY tu.task_update_created_at DESC, tu.task_update_id DESC
  ) filter (where tu.task_update_id is not null), '[]'::jsonb) from task_updates_cte tu) as updates
FROM tasks t
LEFT JOIN users a ON a.id = t.author_id
LEFT JOIN users r ON r.id = t.responsible_id
WHERE t.id = $1;

-- name: ListTasksByProjectId :many
SELECT
  t.*,
  a.id as author_author_id,
  a.name as author_name,
  r.id as responsible_responsible_id,
  r.name as responsible_name,
  coalesce(jsonb_agg(tt.name) filter (where tt.name is not null), '[]') as tags
FROM tasks t
LEFT JOIN users a ON a.id = t.author_id
LEFT JOIN users r ON r.id = t.responsible_id
LEFT JOIN task_tags tt ON tt.task_id = t.id
WHERE project_id = $1
AND (
  cardinality(sqlc.slice('statuses')::text[]) = 0
  OR t.status = ANY(sqlc.slice('statuses')::text[])
)
AND (
  sqlc.narg('cursor_updated_at')::timestamptz IS NULL
  OR t.task_order > sqlc.narg('task_order')::integer
  OR (t.task_order = sqlc.narg('task_order')::integer AND t.updated_at < sqlc.narg('cursor_updated_at')::timestamptz)
)
GROUP BY t.id, a.name, a.id, r.name, r.id
ORDER BY t.task_order ASC, t.updated_at DESC
LIMIT $2;

-- name: ListUserDueTasks :many
SELECT
  t.*,
  p.id as project_project_id,
  p.name as project_name,
  p.description as project_description,
  p.created_at as project_created_at,
  p.updated_at as project_updated_at,
  p.user_id as project_user_id,
  r.id as responsible_responsible_id,
  r.name as responsible_name,
  coalesce(jsonb_agg(tt.name) filter (where tt.name is not null), '[]') as tags
FROM tasks t
LEFT JOIN users r ON r.id = t.responsible_id
LEFT JOIN task_tags tt ON tt.task_id = t.id
JOIN projects p ON p.id = t.project_id
WHERE t.responsible_id = $1
AND (
  cardinality(sqlc.slice('statuses')::text[]) = 0
  OR t.status = ANY(sqlc.slice('statuses')::text[])
)
AND (
  sqlc.narg('cursor_due_date')::timestamptz IS NULL
  OR sqlc.narg('cursor_updated_at')::timestamptz IS NULL
  OR t.due_date > sqlc.narg('cursor_due_date')::timestamptz
  OR (t.due_date = sqlc.narg('cursor_due_date')::timestamptz AND t.updated_at < sqlc.narg('cursor_updated_at')::timestamptz)
)
GROUP BY t.id, r.name, r.id, p.id, p.name, p.description, p.created_at, p.updated_at, p.user_id
ORDER BY t.due_date ASC, t.updated_at DESC
LIMIT $2;

-- name: UpdateTask :exec
UPDATE tasks SET title = $1, description = $2, status = $3, task_order = $4, priority = $5, due_date = $6, responsible_id = $7, done_at = $8, updated_at = CURRENT_TIMESTAMP WHERE id = $9;

-- name: UpdateTaskOrder :exec
UPDATE tasks SET task_order = $1 WHERE id = $2;

-- name: CreateTaskUpdate :one
INSERT INTO task_updates (task_id, user_id, update_type) VALUES ($1, $2, $3) returning id;

-- name: CreateTaskChange :one
INSERT INTO task_changes (update_id, field, old_value, new_value, subject_id, old_value_id, new_value_id, old_display_value, new_display_value) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id;

-- name: GetSmallestOrderProjectTask :one
SELECT id, project_id, title, description, status, created_at, updated_at, author_id, priority, due_date, done_at, responsible_id, task_order FROM tasks WHERE project_id = $1 ORDER BY task_order ASC, updated_at DESC LIMIT 1;

-- name: GetProjectTaskAfterId :one
SELECT t.* 
FROM tasks t
WHERE t.task_order >= (SELECT task_order FROM tasks t2 WHERE t2.id = $1)
  AND t.project_id = $2
  AND t.id != $1
ORDER BY t.task_order ASC, t.updated_at DESC
LIMIT 1;

-- name: MoveTask :one
UPDATE tasks t
SET task_order = $1,
    status = $2,
    updated_at = CURRENT_TIMESTAMP
FROM projects p
JOIN project_members pm ON pm.project_id = p.id AND pm.user_id = $3
WHERE t.id = $4
  AND p.id = t.project_id
RETURNING t.id, t.task_order, t.status;

-- name: CountTasksByProjectIdAndStatus :many
SELECT t.status, COUNT(*) AS count
FROM tasks t
WHERE t.project_id = $1
  AND ($2::text[] IS NULL OR t.status = ANY($2::text[]))
GROUP BY t.status;

-- name: SearchTasksForUser :many
WITH project_ids_cte AS (
  SELECT DISTINCT pm.project_id
  FROM project_members pm
  WHERE pm.user_id = $1
)
SELECT t.*,
  p.id as project_project_id,
  p.name as project_name,
  p.description as project_description,
  p.created_at as project_created_at,
  p.updated_at as project_updated_at,
  p.user_id as project_user_id,
  r.id as responsible_responsible_id,
  r.name as responsible_name,
  r.email as responsible_email,
  r.created_at as responsible_created_at,
  a.id as author_author_id,
  a.name as author_name,
  a.email as author_email,
  a.created_at as author_created_at,
  coalesce(jsonb_agg(tt.name) filter (where tt.name is not null), '[]') as tags
FROM tasks t
LEFT JOIN users r ON r.id = t.responsible_id
LEFT JOIN users a ON a.id = t.author_id
JOIN projects p ON p.id = t.project_id
JOIN project_ids_cte pi ON pi.project_id = t.project_id
LEFT JOIN task_tags tt ON tt.task_id = t.id
WHERE (sqlc.narg('query')::text IS NULL OR (t.title ILIKE '%' || sqlc.narg('query')::text || '%' OR t.description ILIKE '%' || sqlc.narg('query')::text || '%'))
AND (
  sqlc.narg('cursor_due_date')::timestamptz IS NULL
  OR sqlc.narg('cursor_updated_at')::timestamptz IS NULL
  OR t.due_date > sqlc.narg('cursor_due_date')::timestamptz
  OR (t.due_date = sqlc.narg('cursor_due_date')::timestamptz AND t.updated_at < sqlc.narg('cursor_updated_at')::timestamptz)
) AND (
  cardinality(sqlc.slice('statuses')::text[]) = 0
  OR t.status = ANY(sqlc.slice('statuses')::text[])
)
GROUP BY t.id, p.id, p.name, p.description, p.created_at, p.updated_at, p.user_id, r.id, r.name, r.email, r.created_at, a.id, a.name, a.email, a.created_at
ORDER BY t.due_date ASC, t.updated_at DESC
LIMIT $2;
