-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TABLE IF NOT EXISTS project_activity_logs (
  id uuid primary key not null default gen_random_uuid(),
  project_id uuid not null,
  actor_id uuid not null,
  activity_type text not null,
  activity_data jsonb not null,
  entity_type text,
  entity_id uuid,
  created_at timestamp with time zone default current_timestamp not null,
  updated_at timestamp with time zone default current_timestamp not null
);

ALTER TABLE project_activity_logs ADD CONSTRAINT fk_project_activity_logs_projects FOREIGN KEY (project_id) REFERENCES projects(id);
ALTER TABLE project_activity_logs ADD CONSTRAINT fk_project_activity_logs_users FOREIGN KEY (actor_id) REFERENCES users(id);

CREATE INDEX IF NOT EXISTS idx_project_activity_logs_created_at_desc_project_id ON project_activity_logs (created_at DESC, project_id);
CREATE INDEX IF NOT EXISTS idx_project_activity_logs_created_at_desc_actor_id ON project_activity_logs (created_at DESC, actor_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

ALTER TABLE project_activity_logs DROP CONSTRAINT fk_project_activity_logs_projects;
ALTER TABLE project_activity_logs DROP CONSTRAINT fk_project_activity_logs_users;

DROP INDEX IF EXISTS idx_project_activity_logs_created_at_desc_project_id;
DROP INDEX IF EXISTS idx_project_activity_logs_created_at_desc_actor_id;

DROP TABLE IF EXISTS project_activity_logs;
-- +goose StatementEnd
