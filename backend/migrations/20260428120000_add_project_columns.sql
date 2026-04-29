-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS project_columns (
  id uuid primary key not null default gen_random_uuid(),
  project_id uuid not null,
  name text not null,
  color text not null,
  position integer not null,
  is_done_column boolean not null default false,
  created_at timestamp with time zone default current_timestamp not null,
  updated_at timestamp with time zone default current_timestamp not null
);

ALTER TABLE project_columns
  ADD CONSTRAINT fk_project_columns_projects
  FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;

CREATE UNIQUE INDEX IF NOT EXISTS idx_project_columns_project_position
  ON project_columns (project_id, position);

CREATE UNIQUE INDEX IF NOT EXISTS idx_project_columns_single_done
  ON project_columns (project_id, is_done_column)
  WHERE is_done_column = true;

ALTER TABLE tasks ADD COLUMN project_column_id uuid;
ALTER TABLE tasks ADD COLUMN archived_at timestamp with time zone;

ALTER TABLE tasks
  ADD CONSTRAINT fk_tasks_project_columns
  FOREIGN KEY (project_column_id) REFERENCES project_columns(id);

WITH inserted_statuses AS (
  INSERT INTO project_columns (project_id, name, color, position, is_done_column)
  SELECT p.id, s.name, s.color, s.position, s.is_done_column
  FROM projects p
  CROSS JOIN (
    VALUES
      ('Pending', '#64748B', 0, false),
      ('Doing', '#2563EB', 1, false),
      ('Done', '#059669', 2, true)
  ) AS s(name, color, position, is_done_column)
  RETURNING id, project_id, name, position
), archived_recovery AS (
  SELECT DISTINCT ON (tu.task_id)
    tu.task_id,
    tc.old_value AS recovered_status
  FROM task_updates tu
  JOIN task_changes tc ON tc.update_id = tu.id
  WHERE tc.field = 'status'
    AND tc.new_value = 'archived'
    AND tc.old_value <> ''
  ORDER BY tu.task_id, tu.created_at DESC, tu.id DESC
)
UPDATE tasks t
SET
  project_column_id = ps.id,
  archived_at = CASE WHEN lower(t.status) = 'archived' THEN t.updated_at ELSE NULL END
FROM project_columns ps
WHERE ps.project_id = t.project_id
  AND ps.name = CASE
    WHEN lower(t.status) = 'pending' THEN 'Pending'
    WHEN lower(t.status) = 'doing' THEN 'Doing'
    WHEN lower(t.status) = 'done' THEN 'Done'
    WHEN lower(t.status) = 'archived' THEN CASE
      WHEN COALESCE((SELECT recovered_status FROM archived_recovery WHERE task_id = t.id LIMIT 1), 'pending') = 'doing' THEN 'Doing'
      WHEN COALESCE((SELECT recovered_status FROM archived_recovery WHERE task_id = t.id LIMIT 1), 'pending') = 'done' THEN 'Done'
      ELSE 'Pending'
    END
    ELSE 'Pending'
  END;

UPDATE tasks t
SET project_column_id = ps.id
FROM project_columns ps
WHERE t.project_column_id IS NULL
  AND ps.project_id = t.project_id
  AND ps.position = 0;

ALTER TABLE tasks ALTER COLUMN project_column_id SET NOT NULL;
ALTER TABLE tasks DROP COLUMN status;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
