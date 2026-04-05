-- +goose Up
-- +goose StatementBegin
ALTER TABLE tasks ADD COLUMN task_order_next text;

WITH ranked_tasks AS (
  SELECT
    id,
    lpad((row_number() OVER (PARTITION BY project_id, status ORDER BY task_order ASC, updated_at DESC) * 1000)::text, 12, '0') AS next_task_order
  FROM tasks
)
UPDATE tasks t
SET task_order_next = ranked_tasks.next_task_order
FROM ranked_tasks
WHERE ranked_tasks.id = t.id;

ALTER TABLE tasks ALTER COLUMN task_order_next SET NOT NULL;
ALTER TABLE tasks DROP COLUMN task_order;
ALTER TABLE tasks RENAME COLUMN task_order_next TO task_order;
ALTER TABLE tasks ALTER COLUMN task_order SET DEFAULT '500000000000';

DROP INDEX IF EXISTS idx_tasks_task_order;
CREATE INDEX IF NOT EXISTS idx_tasks_project_status_order_updated_at
  ON tasks (project_id, status, task_order ASC, updated_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tasks ADD COLUMN task_order_prev integer;

WITH ranked_tasks AS (
  SELECT
    id,
    row_number() OVER (PARTITION BY project_id, status ORDER BY task_order ASC, updated_at DESC) * 1000 AS prev_task_order
  FROM tasks
)
UPDATE tasks t
SET task_order_prev = ranked_tasks.prev_task_order
FROM ranked_tasks
WHERE ranked_tasks.id = t.id;

ALTER TABLE tasks ALTER COLUMN task_order_prev SET NOT NULL;
ALTER TABLE tasks DROP COLUMN task_order;
ALTER TABLE tasks RENAME COLUMN task_order_prev TO task_order;
ALTER TABLE tasks ALTER COLUMN task_order SET DEFAULT 0;

DROP INDEX IF EXISTS idx_tasks_project_status_order_updated_at;
-- +goose StatementEnd
