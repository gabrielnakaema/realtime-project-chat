-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

ALTER TABLE tasks ADD COLUMN author_id uuid not null default gen_random_uuid();

ALTER TABLE tasks ADD CONSTRAINT fk_tasks_users FOREIGN KEY (author_id) REFERENCES users(id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

ALTER TABLE tasks DROP CONSTRAINT fk_tasks_users;
ALTER TABLE tasks DROP COLUMN author_id;

-- +goose StatementEnd
