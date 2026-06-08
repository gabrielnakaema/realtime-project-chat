-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS task_dependencies (
  task_id uuid NOT NULL,
  depends_on_task_id uuid NOT NULL,
  created_at timestamp with time zone DEFAULT current_timestamp NOT NULL,
  PRIMARY KEY (task_id, depends_on_task_id),
  CONSTRAINT fk_task_dependencies_task FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
  CONSTRAINT fk_task_dependencies_depends_on FOREIGN KEY (depends_on_task_id) REFERENCES tasks(id) ON DELETE CASCADE,
  CONSTRAINT chk_task_dependencies_no_self CHECK (task_id <> depends_on_task_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS task_dependencies;
-- +goose StatementEnd
