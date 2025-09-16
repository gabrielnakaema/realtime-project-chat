-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS "users" (
	id uuid primary key not null default gen_random_uuid(),
	name text not null,
	email text not null,
	password text not null,
	created_at timestamp with time zone default current_timestamp not null,
	updated_at timestamp with time zone default current_timestamp not null
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
	id uuid primary key not null default gen_random_uuid(),
	active bool not null,
	token text not null,
	user_id uuid not null,
	created_at timestamp with time zone default current_timestamp not null,
	expires_at timestamp with time zone not null
);

CREATE TABLE IF NOT EXISTS projects (
	id uuid primary key not null default gen_random_uuid(),
	user_id uuid not null,
	name text not null,
	description text not null,
	created_at timestamp with time zone default current_timestamp not null,
	updated_at timestamp with time zone default current_timestamp not null
);

CREATE TABLE IF NOT EXISTS project_members (
	id uuid primary key not null default gen_random_uuid(),
	user_id uuid not null,
	project_id uuid not null,
	role text not null
);

CREATE TABLE IF NOT EXISTS tasks (
	id uuid primary key not null default gen_random_uuid(),
	project_id uuid not null,
	title text not null,
	description text not null,
	status text not null,
	created_at timestamp with time zone default current_timestamp not null,
	updated_at timestamp with time zone default current_timestamp not null
);

CREATE TABLE IF NOT EXISTS task_changes(
	id uuid primary key not null default gen_random_uuid(),
	task_id uuid not null,
	user_id uuid,
	description text not null,
	created_at timestamp with time zone default current_timestamp not null
);

CREATE TABLE IF NOT EXISTS chats (
	id uuid primary key not null default gen_random_uuid(),
	project_id uuid,
	created_at timestamp with time zone default current_timestamp not null,
	updated_at timestamp with time zone default current_timestamp not null
);

CREATE TABLE IF NOT EXISTS chat_members (
	id uuid primary key not null default gen_random_uuid(),
	user_id uuid not null,
	chat_id uuid not null,
	last_seen_at timestamp with time zone default current_timestamp not null,
	joined_at timestamp with time zone default current_timestamp not null
);

CREATE TABLE IF NOT EXISTS chat_messages (
	id uuid primary key not null default gen_random_uuid(),
	chat_id uuid not null,
	member_id uuid not null,
	content text not null,
	created_at timestamp with time zone default current_timestamp not null,
	updated_at timestamp with time zone default current_timestamp not null
);

ALTER TABLE "users" ADD CONSTRAINT users_email_key UNIQUE (email);
ALTER TABLE refresh_tokens ADD CONSTRAINT fk_refresh_tokens_users FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE projects ADD CONSTRAINT fk_projects_users FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE project_members ADD CONSTRAINT fk_project_members_projects FOREIGN KEY (project_id) REFERENCES projects(id);
ALTER TABLE project_members ADD CONSTRAINT fk_project_members_users FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE tasks ADD CONSTRAINT fk_tasks_projects FOREIGN KEY (project_id) REFERENCES projects(id);
ALTER TABLE task_changes ADD CONSTRAINT fk_task_changes_tasks FOREIGN KEY (task_id) REFERENCES tasks(id);
ALTER TABLE task_changes ADD CONSTRAINT fk_task_changes_users FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE chats ADD CONSTRAINT fk_chats_projects FOREIGN KEY (project_id) REFERENCES projects(id);
ALTER TABLE chat_members ADD CONSTRAINT chat_members_user_id_chat_id_key UNIQUE (user_id, chat_id);
ALTER TABLE chat_members ADD CONSTRAINT fk_chat_members_chats FOREIGN KEY (chat_id) REFERENCES chats(id);
ALTER TABLE chat_members ADD CONSTRAINT fk_chat_members_users FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE chat_messages ADD CONSTRAINT fk_chat_messages_chat_members FOREIGN KEY (member_id) REFERENCES chat_members(id);
ALTER TABLE chat_messages ADD CONSTRAINT fk_chat_messages_chats FOREIGN KEY (chat_id) REFERENCES chats(id);

CREATE INDEX IF NOT EXISTS idx_chat_messages_created_at_desc_chat_id ON chat_messages (created_at DESC, chat_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

DROP TABLE IF EXISTS chat_messages;
DROP TABLE IF EXISTS chat_members;
DROP TABLE IF EXISTS chats;
DROP TABLE IF EXISTS task_changes;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS "users";

-- +goose StatementEnd
