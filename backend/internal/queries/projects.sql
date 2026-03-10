-- name: CreateProject :one
INSERT INTO
  projects (user_id, name, description)
VALUES
  ($1, $2, $3) returning id;

-- name: CreateProjectMember :one
INSERT INTO
  project_members (user_id, project_id, role)
VALUES
  ($1, $2, $3) returning id;

-- name: GetProjectById :one
WITH project_members_cte AS (
  SELECT
    pm.id as project_member_id,
    pm.user_id as project_member_user_id,
    pm.project_id,
    pm.role,
    u.id as user_id,
    u.name,
    u.email,
    u.created_at,
    u.updated_at
  FROM
    project_members pm
    JOIN users u ON u.id = pm.user_id
  WHERE
    pm.project_id = $1
)
SELECT
  p.*,
  coalesce(
    jsonb_agg(
      jsonb_build_object(
        'id',
        pm.project_member_id,
        'user_id',
        pm.project_member_user_id,
        'project_id',
        pm.project_id,
        'role',
        pm.role,
        'user',
        jsonb_build_object(
          'id',
          pm.user_id,
          'name',
          pm.name,
          'email',
          pm.email,
          'created_at',
          pm.created_at,
          'updated_at',
          pm.updated_at
        )
      )
    ) filter (
      where
        pm.project_member_id is not null
    ),
    '[]' :: jsonb
  ) as members
FROM
  projects p
  LEFT JOIN project_members_cte pm ON pm.project_id = p.id
WHERE
  p.id = $1
GROUP BY
  p.id;

-- name: ListProjectsByUserId :many
WITH project_members_cte AS (
  SELECT
    pm.id as project_member_id,
    pm.user_id as project_member_user_id,
    pm.project_id,
    pm.role,
    u.id as user_id,
    u.name,
    u.email,
    u.created_at,
    u.updated_at
  FROM
    project_members pm
    JOIN users u ON u.id = pm.user_id
)
SELECT
  p.*,
  coalesce(
    jsonb_agg(
      jsonb_build_object(
        'id',
        pm.project_member_id,
        'user_id',
        pm.project_member_user_id,
        'project_id',
        pm.project_id,
        'role',
        pm.role,
        'user',
        jsonb_build_object(
          'id',
          pm.user_id,
          'name',
          pm.name,
          'email',
          pm.email,
          'created_at',
          pm.created_at,
          'updated_at',
          pm.updated_at
        )
      )
    ) filter (
      where
        pm.project_member_id is not null
    ),
    '[]' :: jsonb
  ) as members
FROM
  projects p
  INNER JOIN project_members_cte pm ON pm.project_id = p.id
WHERE
  p.id IN (
    SELECT DISTINCT project_id
    FROM project_members
    WHERE project_members.user_id = $1
    AND (
      sqlc.narg('role')::text is null
      or role = sqlc.narg('role')::text
    )
  )
  AND (sqlc.narg('query')::text IS NULL OR (p.name ILIKE '%' || sqlc.narg('query')::text || '%' OR p.description ILIKE '%' || sqlc.narg('query')::text || '%'))
GROUP BY
  p.id
ORDER BY p.updated_at DESC;

-- name: UpdateProject :exec
UPDATE
  projects
SET
  name = $1,
  description = $2,
  updated_at = CURRENT_TIMESTAMP
WHERE
  id = $3;

-- name: RemoveProjectMember :exec
DELETE FROM project_members
WHERE user_id = $1
  AND project_id = $2;

-- name: GetProjectMemberByUserIdAndProjectId :one
SELECT * FROM project_members
WHERE user_id = $1
  AND project_id = $2;

-- name: MarkProjectUpdatedAt :exec
UPDATE projects
SET updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: ListProjectMembersByProjectId :many
SELECT
  pm.id,
  pm.user_id as project_member_user_id,
  pm.project_id as project_member_project_id,
  pm.role as project_member_role,
  u.id as user_id,
  u.name as user_name,
  u.email as user_email,
  u.created_at as user_created_at,
  u.updated_at as user_updated_at
FROM project_members pm
JOIN users u ON u.id = pm.user_id
WHERE pm.project_id = $1
ORDER BY u.name ASC;

-- name: RemoveProjectTasksResponsibleByUserId :exec
UPDATE tasks
  SET responsible_id = NULL
  WHERE responsible_id = $1
  AND project_id = $2;