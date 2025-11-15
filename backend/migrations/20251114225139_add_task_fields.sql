-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE tasks ADD COLUMN priority text not null default 'medium';
ALTER TABLE tasks ADD COLUMN due_date timestamp with time zone;
ALTER TABLE tasks ADD COLUMN done_at timestamp with time zone;

ALTER TABLE tasks ADD COLUMN responsible_id uuid;
ALTER TABLE tasks ADD CONSTRAINT fk_tasks_users_responsible FOREIGN KEY (responsible_id) REFERENCES users(id);

ALTER TABLE tasks ADD COLUMN task_order int not null default 0;

CREATE TABLE IF NOT EXISTS task_tags (
  task_id uuid not null,
  name text not null,
  primary key (task_id, name),
  created_at timestamp with time zone default current_timestamp not null
);

ALTER TABLE task_tags ADD CONSTRAINT fk_task_tags_tasks FOREIGN KEY (task_id) REFERENCES tasks(id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE IF EXISTS task_tags;
ALTER TABLE tasks DROP CONSTRAINT IF EXISTS fk_tasks_users_responsible;
ALTER TABLE tasks DROP COLUMN IF EXISTS responsible_id;
ALTER TABLE tasks DROP COLUMN IF EXISTS done_at;
ALTER TABLE tasks DROP COLUMN IF EXISTS due_date;
ALTER TABLE tasks DROP COLUMN IF EXISTS priority;
ALTER TABLE tasks DROP COLUMN IF EXISTS task_order;
-- +goose StatementEnd
