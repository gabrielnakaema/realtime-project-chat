-- +goose Up
-- +goose StatementBegin
ALTER TABLE tasks ADD COLUMN version integer NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tasks DROP COLUMN IF EXISTS version;
-- +goose StatementEnd
