-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

ALTER TABLE task_changes DROP CONSTRAINT fk_task_changes_tasks;
ALTER TABLE task_changes DROP CONSTRAINT fk_task_changes_users;

ALTER TABLE task_changes DROP COLUMN task_id;
ALTER TABLE task_changes DROP COLUMN user_id;
ALTER TABLE task_changes DROP COLUMN description;

CREATE TABLE IF NOT EXISTS task_updates (
  id uuid primary key not null default gen_random_uuid(),
  task_id uuid not null,
  user_id uuid not null,
  update_type text not null,
  created_at timestamp with time zone default current_timestamp not null
);

ALTER TABLE task_updates ADD CONSTRAINT fk_task_updates_tasks FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE;
ALTER TABLE task_updates ADD CONSTRAINT fk_task_updates_users FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE task_changes ADD COLUMN update_id uuid;
ALTER TABLE task_changes ADD CONSTRAINT fk_task_changes_task_updates FOREIGN KEY (update_id) REFERENCES task_updates(id);

ALTER TABLE task_changes ADD COLUMN field text not null default '';
ALTER TABLE task_changes ADD COLUMN old_value text not null default '';
ALTER TABLE task_changes ADD COLUMN new_value text not null default '';

ALTER TABLE task_changes ADD COLUMN subject_id uuid;
ALTER TABLE task_changes ADD CONSTRAINT fk_task_changes_users FOREIGN KEY (subject_id) REFERENCES users(id);

CREATE INDEX IF NOT EXISTS idx_task_updates_created_at_desc_task_id ON task_updates (created_at DESC, task_id);
CREATE INDEX IF NOT EXISTS idx_task_changes_created_at_desc_update_id ON task_changes (created_at DESC, update_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';


-- +goose StatementEnd
