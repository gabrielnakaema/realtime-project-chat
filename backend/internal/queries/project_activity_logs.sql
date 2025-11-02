-- name: CreateProjectActivityLog :one
INSERT INTO project_activity_logs (
    project_id,
    actor_id,
    activity_type,
    activity_data,
    entity_type,
    entity_id
  )
VALUES ($1, $2, $3, $4, $5, $6)
returning id;

-- name: ListProjectActivityLogs :many
WITH user_projects AS (
  SELECT DISTINCT pm.project_id
  FROM project_members pm
  WHERE (pm.user_id = sqlc.narg('user_id')::uuid OR sqlc.narg('user_id')::uuid IS NULL)
),
activity_logs AS (
  SELECT 
    pal.id,
    pal.project_id,
    pal.actor_id,
    pal.activity_type,
    COALESCE(pal.activity_data, '{}'::jsonb) as activity_data,
    pal.entity_type,
    pal.entity_id,
    pal.created_at,
    pal.updated_at
  FROM project_activity_logs pal
  WHERE 
    sqlc.arg('filter_by_user')::boolean = false 
    OR pal.project_id IN (SELECT project_id FROM user_projects)
),
task_entities AS (
  SELECT 
    t.id,
    jsonb_build_object(
      'id', t.id,
      'project_id', t.project_id,
      'title', t.title,
      'description', t.description,
      'status', t.status,
      'created_at', t.created_at,
      'updated_at', t.updated_at
    ) as entity_data
  FROM tasks t
),
project_member_entities AS (
  SELECT 
    pm.id,
    jsonb_build_object(
      'id', pm.id,
      'user_id', pm.user_id,
      'project_id', pm.project_id,
      'role', pm.role,
      'user', jsonb_build_object(
        'id', u.id,
        'name', u.name,
        'email', u.email,
        'created_at', u.created_at,
        'updated_at', u.updated_at
      )
    ) as entity_data
  FROM project_members pm
  JOIN users u ON pm.user_id = u.id
)
SELECT
  al.id,
  al.project_id,
  al.actor_id,
  al.activity_type,
  al.activity_data,
  al.entity_type,
  al.entity_id,
  al.created_at,
  al.updated_at,
  p.name as project_name,
  p.description as project_description,
  p.user_id as project_user_id,
  p.created_at as project_created_at,
  p.updated_at as project_updated_at,
  actor.name as actor_name,
  actor.email as actor_email,
  CASE 
    WHEN al.entity_type = 'task' THEN te.entity_data
    WHEN al.entity_type = 'project.member' THEN pme.entity_data
    WHEN al.entity_type = 'project' THEN '{}'::jsonb
    ELSE '{}'::jsonb
  END as entity
FROM activity_logs al
JOIN projects p ON al.project_id = p.id
JOIN users actor ON al.actor_id = actor.id
LEFT JOIN task_entities te ON al.entity_type = 'task' AND al.entity_id = te.id
LEFT JOIN project_member_entities pme ON al.entity_type = 'project.member' AND al.entity_id = pme.id
WHERE (al.project_id = sqlc.narg('project_id')::uuid OR sqlc.narg('project_id')::uuid IS NULL)
AND (sqlc.narg('before_created_at')::timestamptz IS NULL OR al.created_at < sqlc.narg('before_created_at')::timestamptz)
AND (sqlc.narg('before_id')::uuid IS NULL OR al.id != sqlc.narg('before_id')::uuid)
ORDER BY al.created_at DESC
LIMIT (COALESCE(sqlc.narg('limit')::integer, 0) + 1);
