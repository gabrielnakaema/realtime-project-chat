-- +goose Up
-- +goose StatementBegin
ALTER TABLE projects
  ADD COLUMN repository_url text,
  ADD COLUMN repository_owner text,
  ADD COLUMN repository_name text,
  ADD COLUMN default_branch text,
  ADD COLUMN branch_name_prefix text;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE projects
  DROP COLUMN branch_name_prefix,
  DROP COLUMN default_branch,
  DROP COLUMN repository_name,
  DROP COLUMN repository_owner,
  DROP COLUMN repository_url;
-- +goose StatementEnd
