-- +goose Up
-- +goose StatementBegin
ALTER TABLE project_columns
  ADD COLUMN description text NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE project_columns
  DROP COLUMN description;
-- +goose StatementEnd
