-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_tasks_project_code_prefix
ON tasks (project_id, (lower(btrim(code))) text_pattern_ops)
WHERE code IS NOT NULL AND btrim(code) <> '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_tasks_project_code_prefix;
-- +goose StatementEnd
