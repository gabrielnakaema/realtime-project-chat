-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TABLE IF NOT EXISTS mcp_api_keys (
  id uuid primary key not null default gen_random_uuid(),
  user_id uuid not null,
  name text not null,
  key_prefix text not null,
  secret_hash text not null,
  created_at timestamp with time zone default current_timestamp not null,
  last_used_at timestamp with time zone,
  revoked_at timestamp with time zone
);

ALTER TABLE mcp_api_keys ADD CONSTRAINT fk_mcp_api_keys_users FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
CREATE UNIQUE INDEX IF NOT EXISTS idx_mcp_api_keys_key_prefix ON mcp_api_keys (key_prefix);
CREATE INDEX IF NOT EXISTS idx_mcp_api_keys_user_id ON mcp_api_keys (user_id);

CREATE TABLE IF NOT EXISTS mcp_api_key_scopes (
  api_key_id uuid not null,
  scope text not null,
  created_at timestamp with time zone default current_timestamp not null,
  primary key (api_key_id, scope)
);

ALTER TABLE mcp_api_key_scopes ADD CONSTRAINT fk_mcp_api_key_scopes_api_key FOREIGN KEY (api_key_id) REFERENCES mcp_api_keys(id) ON DELETE CASCADE;

ALTER TABLE task_comments ADD COLUMN IF NOT EXISTS action_origin text not null default 'user';
ALTER TABLE task_updates ADD COLUMN IF NOT EXISTS action_origin text not null default 'user';
ALTER TABLE project_activity_logs ADD COLUMN IF NOT EXISTS action_origin text not null default 'user';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

ALTER TABLE project_activity_logs DROP COLUMN IF EXISTS action_origin;
ALTER TABLE task_updates DROP COLUMN IF EXISTS action_origin;
ALTER TABLE task_comments DROP COLUMN IF EXISTS action_origin;

ALTER TABLE mcp_api_key_scopes DROP CONSTRAINT IF EXISTS fk_mcp_api_key_scopes_api_key;
DROP TABLE IF EXISTS mcp_api_key_scopes;

DROP INDEX IF EXISTS idx_mcp_api_keys_user_id;
DROP INDEX IF EXISTS idx_mcp_api_keys_key_prefix;
ALTER TABLE mcp_api_keys DROP CONSTRAINT IF EXISTS fk_mcp_api_keys_users;
DROP TABLE IF EXISTS mcp_api_keys;

-- +goose StatementEnd
