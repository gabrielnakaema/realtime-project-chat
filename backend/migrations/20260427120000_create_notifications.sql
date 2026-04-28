-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TABLE IF NOT EXISTS notifications (
  id uuid primary key not null default gen_random_uuid(),
  user_id uuid not null,
  actor_id uuid not null,
  project_id uuid not null,
  task_id uuid,
  task_comment_id uuid,
  type text not null,
  read_at timestamp with time zone,
  created_at timestamp with time zone default current_timestamp not null,
  updated_at timestamp with time zone default current_timestamp not null
);

ALTER TABLE notifications ADD CONSTRAINT fk_notifications_users FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE notifications ADD CONSTRAINT fk_notifications_actors FOREIGN KEY (actor_id) REFERENCES users(id);
ALTER TABLE notifications ADD CONSTRAINT fk_notifications_projects FOREIGN KEY (project_id) REFERENCES projects(id);
ALTER TABLE notifications ADD CONSTRAINT fk_notifications_tasks FOREIGN KEY (task_id) REFERENCES tasks(id);
ALTER TABLE notifications ADD CONSTRAINT fk_notifications_task_comments FOREIGN KEY (task_comment_id) REFERENCES task_comments(id);

CREATE INDEX IF NOT EXISTS idx_notifications_user_created_at_desc ON notifications (user_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_user_read_at ON notifications (user_id, read_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

DROP INDEX IF EXISTS idx_notifications_user_read_at;
DROP INDEX IF EXISTS idx_notifications_user_created_at_desc;
DROP TABLE IF EXISTS notifications;
-- +goose StatementEnd
