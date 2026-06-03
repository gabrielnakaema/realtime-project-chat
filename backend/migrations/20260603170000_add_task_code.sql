-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE tasks ADD COLUMN code text;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE tasks DROP COLUMN IF EXISTS code;
-- +goose StatementEnd
