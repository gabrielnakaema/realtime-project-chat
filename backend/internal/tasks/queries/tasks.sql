-- name: CreateTask :one
INSERT INTO tasks (project_id, title, description, code, project_column_id, author_id, responsible_id, priority, due_date, done_at, task_order)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) returning id;

-- name: CreateTaskTag :exec
INSERT INTO task_tags (task_id, name) VALUES ($1, $2);

-- name: DeleteTaskTag :exec
DELETE FROM task_tags WHERE task_id = $1 AND name = $2;

-- name: DeleteAllTaskTags :exec
DELETE FROM task_tags WHERE task_id = $1;

-- name: CreateTaskDependency :exec
INSERT INTO task_dependencies (task_id, depends_on_task_id) VALUES ($1, $2);

-- name: DeleteAllTaskDependencies :exec
DELETE FROM task_dependencies WHERE task_id = $1;

-- name: CountTasksInProjectByIds :one
SELECT COUNT(*)::int AS count
FROM tasks
WHERE project_id = $1 AND id = ANY($2::uuid[]);

-- name: ListTaskDependenciesByProjectId :many
SELECT td.task_id, td.depends_on_task_id
FROM task_dependencies td
JOIN tasks t ON t.id = td.task_id
WHERE t.project_id = $1;

-- name: GetTaskById :one
WITH task_tags_cte AS (
  SELECT
    tt.name as task_tag_name
  from task_tags tt
  WHERE tt.task_id = $1
), task_dependencies_cte AS (
  SELECT
    td.depends_on_task_id,
    dep.title as depends_on_title,
    dep.code as depends_on_code
  FROM task_dependencies td
  JOIN tasks dep ON dep.id = td.depends_on_task_id
  WHERE td.task_id = $1
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
    tu.action_origin as task_update_action_origin,
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
  t.code as task_code,
  t.project_column_id as task_project_column_id,
  t.priority as task_priority,
  t.responsible_id as task_responsible_id,
  t.due_date as task_due_date,
  t.done_at as task_done_at,
  t.archived_at as task_archived_at,
  t.task_order as task_order,
  t.version as task_version,
  t.created_at as task_created_at,
  t.updated_at as task_updated_at,
  t.author_id as task_author_id,
  a.name as task_author_name,
  a.email as task_author_email,
  a.created_at as task_author_created_at,
  r.name as task_responsible_name,
  r.email as task_responsible_email,
  r.created_at as task_responsible_created_at,
  jsonb_build_object(
    'id', ps.id,
    'project_id', ps.project_id,
    'name', ps.name,
    'color', ps.color,
    'position', ps.position,
    'is_done_column', ps.is_done_column,
    'created_at', ps.created_at,
    'updated_at', ps.updated_at
  ) as project_column,
  (select coalesce(jsonb_agg(tt.task_tag_name) filter (where tt.task_tag_name is not null), '[]') from task_tags_cte tt) as tags,
  (select coalesce(jsonb_agg(
    jsonb_build_object(
      'id', tdc.depends_on_task_id,
      'title', tdc.depends_on_title,
      'code', coalesce(tdc.depends_on_code, '')
    )
    ORDER BY tdc.depends_on_title, tdc.depends_on_task_id
  ) filter (where tdc.depends_on_task_id is not null), '[]') from task_dependencies_cte tdc) as depends_on_tasks,
  (select coalesce(jsonb_agg(
    jsonb_build_object(
      'id', tu.task_update_id,
      'task_id', tu.task_update_task_id,
      'user_id', tu.task_update_user_id,
      'update_type', tu.task_update_update_type,
      'action_origin', tu.task_update_action_origin,
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
JOIN project_columns ps ON ps.id = t.project_column_id
WHERE t.id = $1;

-- name: ListTasksByProjectId :many
SELECT
  t.*,
  ps.id as project_column_id_2,
  ps.project_id as project_column_project_id,
  ps.name as project_column_name,
  ps.color as project_column_color,
  ps.position as project_column_position,
  ps.is_done_column as project_column_is_done_column,
  ps.created_at as project_column_created_at,
  ps.updated_at as project_column_updated_at,
  a.id as author_author_id,
  a.name as author_name,
  r.id as responsible_responsible_id,
  r.name as responsible_name,
  coalesce(jsonb_agg(DISTINCT tt.name) filter (where tt.name is not null), '[]') as tags,
  coalesce(jsonb_agg(DISTINCT td.depends_on_task_id) filter (where td.depends_on_task_id is not null), '[]') as depends_on_task_ids
FROM tasks t
JOIN project_columns ps ON ps.id = t.project_column_id
LEFT JOIN users a ON a.id = t.author_id
LEFT JOIN users r ON r.id = t.responsible_id
LEFT JOIN task_tags tt ON tt.task_id = t.id
LEFT JOIN task_dependencies td ON td.task_id = t.id
WHERE t.project_id = $1
AND (
  cardinality(sqlc.slice('project_column_ids')::uuid[]) = 0
  OR t.project_column_id = ANY(sqlc.slice('project_column_ids')::uuid[])
)
AND (
  (sqlc.arg('archived')::boolean = true AND t.archived_at IS NOT NULL)
  OR (sqlc.arg('archived')::boolean = false AND t.archived_at IS NULL)
)
AND (
  sqlc.narg('cursor_updated_at')::timestamptz IS NULL
  OR t.task_order > sqlc.narg('task_order')::text
  OR (t.task_order = sqlc.narg('task_order')::text AND t.updated_at < sqlc.narg('cursor_updated_at')::timestamptz)
)
GROUP BY t.id, ps.id, a.name, a.id, r.name, r.id
ORDER BY t.task_order ASC, t.updated_at DESC
LIMIT $2;

-- name: ListUserDueTasks :many
SELECT
  t.*,
  ps.id as project_column_id_2,
  ps.project_id as project_column_project_id,
  ps.name as project_column_name,
  ps.color as project_column_color,
  ps.position as project_column_position,
  ps.is_done_column as project_column_is_done_column,
  ps.created_at as project_column_created_at,
  ps.updated_at as project_column_updated_at,
  p.id as project_project_id,
  p.name as project_name,
  p.description as project_description,
  p.created_at as project_created_at,
  p.updated_at as project_updated_at,
  p.user_id as project_user_id,
  r.id as responsible_responsible_id,
  r.name as responsible_name,
  coalesce(jsonb_agg(DISTINCT tt.name) filter (where tt.name is not null), '[]') as tags,
  coalesce(jsonb_agg(DISTINCT td.depends_on_task_id) filter (where td.depends_on_task_id is not null), '[]') as depends_on_task_ids
FROM tasks t
JOIN project_columns ps ON ps.id = t.project_column_id
LEFT JOIN users r ON r.id = t.responsible_id
LEFT JOIN task_tags tt ON tt.task_id = t.id
LEFT JOIN task_dependencies td ON td.task_id = t.id
JOIN projects p ON p.id = t.project_id
WHERE t.responsible_id = $1
AND t.due_date IS NOT NULL
AND t.archived_at IS NULL
AND ps.is_done_column = false
AND (
  sqlc.narg('cursor_due_date')::timestamptz IS NULL
  OR sqlc.narg('cursor_updated_at')::timestamptz IS NULL
  OR t.due_date > sqlc.narg('cursor_due_date')::timestamptz
  OR (t.due_date = sqlc.narg('cursor_due_date')::timestamptz AND t.updated_at < sqlc.narg('cursor_updated_at')::timestamptz)
)
GROUP BY t.id, ps.id, r.name, r.id, p.id, p.name, p.description, p.created_at, p.updated_at, p.user_id
ORDER BY t.due_date ASC, t.updated_at DESC
LIMIT $2;

-- name: UpdateTask :one
UPDATE tasks SET title = $1, description = $2, code = $3, project_column_id = $4, task_order = $5, priority = $6, due_date = $7, responsible_id = $8, done_at = $9, archived_at = $10, version = version + 1, updated_at = CURRENT_TIMESTAMP WHERE id = $11 RETURNING version, updated_at;

-- name: CreateTaskUpdate :one
INSERT INTO task_updates (task_id, user_id, update_type, action_origin) VALUES ($1, $2, $3, $4) returning id;

-- name: CreateTaskChange :one
INSERT INTO task_changes (update_id, field, old_value, new_value, subject_id, old_value_id, new_value_id, old_display_value, new_display_value) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id;

-- name: GetFirstTaskInColumn :one
SELECT id, project_id, title, description, code, project_column_id, created_at, updated_at, author_id, priority, due_date, done_at, archived_at, responsible_id, task_order
FROM tasks
WHERE project_id = $1
  AND project_column_id = $2
  AND archived_at IS NULL
ORDER BY task_order ASC, updated_at DESC
LIMIT 1;

-- name: GetProjectTaskAfterId :one
SELECT t.* 
FROM tasks t
JOIN tasks current ON current.id = $1
WHERE t.project_id = $2
  AND t.project_column_id = current.project_column_id
  AND t.id != current.id
  AND t.archived_at IS NULL
  AND (
    t.task_order > current.task_order
    OR (t.task_order = current.task_order AND t.updated_at < current.updated_at)
    OR (t.task_order = current.task_order AND t.updated_at = current.updated_at AND t.id > current.id)
  )
ORDER BY t.task_order ASC, t.updated_at DESC, t.id ASC
LIMIT 1;

-- name: MoveTask :one
UPDATE tasks t
SET task_order = $1,
    project_column_id = $2,
    done_at = $3,
    version = t.version + 1,
    updated_at = CURRENT_TIMESTAMP
FROM projects p
JOIN project_members pm ON pm.project_id = p.id AND pm.user_id = $4
WHERE t.id = $5
  AND p.id = t.project_id
RETURNING t.id, t.task_order, t.project_column_id, t.version;

-- name: CountTasksByProjectIdAndColumn :many
SELECT t.project_column_id, COUNT(*) AS count
FROM tasks t
WHERE t.project_id = $1
  AND t.archived_at IS NULL
  AND ($2::uuid[] IS NULL OR t.project_column_id = ANY($2::uuid[]))
GROUP BY t.project_column_id;

-- name: SearchTasksForUser :many
WITH project_ids_cte AS (
  SELECT DISTINCT pm.project_id
  FROM project_members pm
  WHERE pm.user_id = $1
)
SELECT t.*,
  ps.id as project_column_id_2,
  ps.project_id as project_column_project_id,
  ps.name as project_column_name,
  ps.color as project_column_color,
  ps.position as project_column_position,
  ps.is_done_column as project_column_is_done_column,
  ps.created_at as project_column_created_at,
  ps.updated_at as project_column_updated_at,
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
  coalesce(jsonb_agg(DISTINCT tt.name) filter (where tt.name is not null), '[]') as tags,
  coalesce(jsonb_agg(DISTINCT td.depends_on_task_id) filter (where td.depends_on_task_id is not null), '[]') as depends_on_task_ids
FROM tasks t
JOIN project_columns ps ON ps.id = t.project_column_id
LEFT JOIN users r ON r.id = t.responsible_id
LEFT JOIN users a ON a.id = t.author_id
JOIN projects p ON p.id = t.project_id
JOIN project_ids_cte pi ON pi.project_id = t.project_id
LEFT JOIN task_tags tt ON tt.task_id = t.id
LEFT JOIN task_dependencies td ON td.task_id = t.id
WHERE (sqlc.narg('query')::text IS NULL OR (t.title ILIKE '%' || sqlc.narg('query')::text || '%' OR t.description ILIKE '%' || sqlc.narg('query')::text || '%' OR t.code ILIKE '%' || sqlc.narg('query')::text || '%'))
AND t.archived_at IS NULL
AND ps.is_done_column = false
AND (
  sqlc.narg('cursor_due_date')::timestamptz IS NULL
  OR sqlc.narg('cursor_updated_at')::timestamptz IS NULL
  OR t.due_date > sqlc.narg('cursor_due_date')::timestamptz
  OR (t.due_date = sqlc.narg('cursor_due_date')::timestamptz AND t.updated_at < sqlc.narg('cursor_updated_at')::timestamptz)
)
GROUP BY t.id, ps.id, p.id, p.name, p.description, p.created_at, p.updated_at, p.user_id, r.id, r.name, r.email, r.created_at, a.id, a.name, a.email, a.created_at
ORDER BY t.due_date ASC, t.updated_at DESC
LIMIT $2;

-- name: SearchProjectTasksForDependencies :many
SELECT
  t.id,
  t.title,
  coalesce(t.code, '') as code
FROM tasks t
WHERE t.project_id = sqlc.arg('project_id')
  AND t.archived_at IS NULL
  AND (
    sqlc.narg('exclude_task_id')::uuid IS NULL
    OR t.id <> sqlc.narg('exclude_task_id')::uuid
  )
  AND (
    t.title ILIKE '%' || sqlc.arg('query')::text || '%'
    OR coalesce(t.code, '') ILIKE '%' || sqlc.arg('query')::text || '%'
    OR t.id::text ILIKE '%' || sqlc.arg('query')::text || '%'
    OR lower(regexp_replace(regexp_replace(t.description, '<[^>]+>', ' ', 'g'), '\s+', ' ', 'g')) LIKE '%' || lower(sqlc.arg('query')::text) || '%'
  )
ORDER BY t.title ASC, t.id ASC
LIMIT sqlc.arg('limit');

-- name: SuggestTaskCodesByProjectPrefix :many
WITH matches AS (
  SELECT
    btrim(t.code) AS code,
    substring(btrim(t.code) from length(sqlc.arg('sequence_base')::text) + 1) AS suffix,
    substring(btrim(t.code) from 1 for length(sqlc.arg('sequence_base')::text)) AS matched_base
  FROM tasks t
  WHERE t.project_id = sqlc.arg('project_id')
    AND t.code IS NOT NULL
    AND btrim(t.code) <> ''
    AND lower(btrim(t.code)) LIKE sqlc.arg('sequence_base_pattern')::text ESCAPE '\'
),
existing_matches AS (
  SELECT
    btrim(t.code) AS code
  FROM tasks t
  WHERE t.project_id = sqlc.arg('project_id')
    AND t.code IS NOT NULL
    AND btrim(t.code) <> ''
    AND lower(btrim(t.code)) LIKE sqlc.arg('prefix_pattern')::text ESCAPE '\'
),
numeric_matches AS (
  SELECT
    code,
    matched_base,
    suffix::bigint AS number,
    length(suffix) AS suffix_width
  FROM matches
  WHERE suffix ~ '^[0-9]{1,18}$'
),
next_number AS (
  SELECT
    coalesce(max(number), 0) + 1 AS number,
    coalesce(max(suffix_width), 1) AS suffix_width,
    coalesce((array_agg(matched_base ORDER BY number DESC, code DESC))[1], sqlc.arg('sequence_base')::text) AS base
  FROM numeric_matches
),
next_code AS (
  SELECT
    base || lpad(number::text, greatest(length(number::text), suffix_width), '0') AS code,
    'next'::text AS kind,
    0 AS sort_rank,
    NULL::bigint AS sort_number
  FROM next_number
),
existing_codes AS (
  SELECT
    code,
    kind,
    sort_rank,
    sort_number
  FROM (
    SELECT DISTINCT
      em.code,
      'existing'::text AS kind,
      1 AS sort_rank,
      nm.number AS sort_number
    FROM existing_matches em
    LEFT JOIN numeric_matches nm ON nm.code = em.code
  ) deduped
  ORDER BY sort_number DESC NULLS LAST, code DESC
  LIMIT greatest(sqlc.arg('limit')::int - 1, 0)
)
SELECT code::text AS code, kind::text AS kind
FROM (
  SELECT code, kind, sort_rank, sort_number FROM next_code
  UNION ALL
  SELECT code, kind, sort_rank, sort_number FROM existing_codes
) suggestions
ORDER BY sort_rank ASC, sort_number DESC NULLS LAST, code DESC
LIMIT sqlc.arg('limit');

-- name: FindTaskRefsByProjectAndCode :many
SELECT
  t.id,
  t.title,
  coalesce(t.code, '') as code
FROM tasks t
WHERE t.project_id = sqlc.arg('project_id')
  AND t.archived_at IS NULL
  AND lower(btrim(coalesce(t.code, ''))) = lower(btrim(sqlc.arg('code')::text))
ORDER BY t.updated_at DESC, t.created_at DESC, t.id DESC;

-- name: ClearTasksResponsibleForProjectMember :exec
UPDATE tasks
  SET responsible_id = NULL
  WHERE responsible_id = $1
  AND project_id = $2;
