-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TABLE IF NOT EXISTS task_comments (
  id uuid primary key not null default gen_random_uuid(),
  task_id uuid not null,
  user_id uuid not null,
  content text not null,
  parent_comment_id uuid,
  created_at timestamp with time zone default current_timestamp not null,
  updated_at timestamp with time zone default current_timestamp not null
);

ALTER TABLE task_comments ADD CONSTRAINT fk_task_comments_tasks FOREIGN KEY (task_id) REFERENCES tasks(id);
ALTER TABLE task_comments ADD CONSTRAINT fk_task_comments_users FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE task_comments ADD CONSTRAINT fk_task_comments_parent_comments FOREIGN KEY (parent_comment_id) REFERENCES task_comments(id);

CREATE INDEX IF NOT EXISTS idx_task_comments_created_at_desc_task_id ON task_comments (created_at DESC, task_id);
CREATE INDEX IF NOT EXISTS idx_task_comments_created_at_desc_parent_comment_id ON task_comments (created_at DESC, parent_comment_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE IF EXISTS task_comments;
ALTER TABLE task_comments DROP CONSTRAINT fk_task_comments_tasks;
ALTER TABLE task_comments DROP CONSTRAINT fk_task_comments_users;
ALTER TABLE task_comments DROP CONSTRAINT fk_task_comments_parent_comments;

DROP INDEX IF EXISTS idx_task_comments_created_at_desc_task_id;
DROP INDEX IF EXISTS idx_task_comments_created_at_desc_parent_comment_id;
-- +goose StatementEnd
