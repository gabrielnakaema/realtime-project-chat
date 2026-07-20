-- name: CreateProject :one
INSERT INTO
  projects (
    user_id,
    name,
    description,
    repository_url,
    repository_owner,
    repository_name,
    default_branch,
    branch_name_prefix
  )
VALUES
  ($1, $2, $3, $4, $5, $6, $7, $8) returning id;

-- name: CreateProjectColumn :one
INSERT INTO
  project_columns (project_id, name, description, color, position, is_done_column)
VALUES
  ($1, $2, $3, $4, $5, $6) returning id;

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
), project_columns_cte AS (
  SELECT
    ps.id,
    ps.project_id,
    ps.name,
    ps.description,
    ps.color,
    ps.position,
    ps.is_done_column,
    ps.created_at,
    ps.updated_at
  FROM project_columns ps
  WHERE ps.project_id = $1
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
  ) as members,
  (
    SELECT coalesce(jsonb_agg(
      jsonb_build_object(
        'id', ps.id,
        'project_id', ps.project_id,
        'name', ps.name,
        'description', ps.description,
        'color', ps.color,
        'position', ps.position,
        'is_done_column', ps.is_done_column,
        'created_at', ps.created_at,
        'updated_at', ps.updated_at
      )
      ORDER BY ps.position ASC, ps.created_at ASC
    ), '[]'::jsonb)
    FROM project_columns_cte ps
  ) as columns
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
), project_columns_cte AS (
  SELECT
    ps.id,
    ps.project_id,
    ps.name,
    ps.description,
    ps.color,
    ps.position,
    ps.is_done_column,
    ps.created_at,
    ps.updated_at
  FROM project_columns ps
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
  ) as members,
  (
    SELECT coalesce(jsonb_agg(
      jsonb_build_object(
        'id', ps.id,
        'project_id', ps.project_id,
        'name', ps.name,
        'description', ps.description,
        'color', ps.color,
        'position', ps.position,
        'is_done_column', ps.is_done_column,
        'created_at', ps.created_at,
        'updated_at', ps.updated_at
      )
      ORDER BY ps.position ASC, ps.created_at ASC
    ), '[]'::jsonb)
    FROM project_columns_cte ps
    WHERE ps.project_id = p.id
  ) as columns
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
  repository_url = $3,
  repository_owner = $4,
  repository_name = $5,
  default_branch = $6,
  branch_name_prefix = $7,
  updated_at = CURRENT_TIMESTAMP
WHERE
  id = $8;

-- name: UpdateProjectColumn :exec
UPDATE
  project_columns
SET
  name = $1,
  description = $2,
  color = $3,
  position = $4,
  is_done_column = $5,
  updated_at = CURRENT_TIMESTAMP
WHERE
  id = $6
  AND project_id = $7;

-- name: DeleteProjectColumn :exec
DELETE FROM project_columns
WHERE id = $1
  AND project_id = $2;

-- name: ReassignTasksToProjectColumn :exec
UPDATE tasks
SET
  project_column_id = $1,
  done_at = CASE
    WHEN $2::boolean = true AND archived_at IS NULL AND done_at IS NULL THEN CURRENT_TIMESTAMP
    WHEN $2::boolean = false THEN NULL
    ELSE done_at
  END,
  updated_at = CURRENT_TIMESTAMP
WHERE project_column_id = $3;

-- name: GetProjectColumnById :one
SELECT id, project_id, name, description, color, position, is_done_column, created_at, updated_at
FROM project_columns
WHERE id = $1;

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
