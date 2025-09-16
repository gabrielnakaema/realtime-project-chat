-- name: CreateTask :one
INSERT INTO tasks (project_id, title, description, status, author_id) VALUES ($1, $2, $3, $4, $5) returning id;

-- name: GetTaskById :one
WITH task_changes_cte AS (
  SELECT 
    tc.id as task_change_id,
    tc.task_id as task_change_task_id,
    tc.user_id as task_change_user_id,
    tc.description as task_change_description,
    tc.created_at as task_change_created_at,
    a.id as task_change_author_id,
    a.name as task_change_author_name,
    a.email as task_change_author_email,
    a.created_at as task_change_author_created_at
   FROM task_changes tc
   JOIN users a ON a.id = tc.user_id
   WHERE tc.task_id = $1
  ORDER BY tc.created_at ASC
)
SELECT 
  t.id as task_id,
  t.project_id as task_project_id,
  t.title as task_title,
  t.description as task_description,
  t.status as task_status,
  t.created_at as task_created_at,
  t.updated_at as task_updated_at,
  t.author_id as task_author_id,
  a.name as task_author_name,
  a.email as task_author_email,
  a.created_at as task_author_created_at,
  coalesce(jsonb_agg(
    jsonb_build_object(
      'id',
      tc.task_change_id,
      'task_id',
      tc.task_change_task_id,
      'author_id',
      tc.task_change_author_id,
      'change_description',
      tc.task_change_description,
      'created_at',
      tc.task_change_created_at,
      'author',
      jsonb_build_object(
        'id',
        tc.task_change_author_id,
        'name',
        tc.task_change_author_name,
        'email',
        tc.task_change_author_email,
        'created_at',
        tc.task_change_author_created_at
      )
    ) filter (where tc.task_change_id is not null)
  ), '[]'::jsonb) as task_changes
FROM tasks t
LEFT JOIN task_changes_cte tc ON tc.task_change_task_id = t.id
LEFT JOIN users a ON a.id = t.author_id
WHERE t.id = $1;

-- name: ListTasksByProjectId :many
SELECT * FROM tasks WHERE project_id = $1;

-- name: UpdateTask :exec
UPDATE tasks SET title = $1, description = $2, status = $3, updated_at = CURRENT_TIMESTAMP WHERE id = $4;

-- name: CreateTaskChange :one
INSERT INTO task_changes (task_id, user_id, description) VALUES ($1, $2, $3) returning id;