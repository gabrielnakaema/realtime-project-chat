-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

ALTER TABLE task_changes ADD COLUMN IF NOT EXISTS old_value_id uuid;
ALTER TABLE task_changes ADD COLUMN IF NOT EXISTS new_value_id uuid;
ALTER TABLE task_changes ADD COLUMN IF NOT EXISTS old_display_value text;
ALTER TABLE task_changes ADD COLUMN IF NOT EXISTS new_display_value text;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

ALTER TABLE task_changes DROP COLUMN IF EXISTS new_display_value;
ALTER TABLE task_changes DROP COLUMN IF EXISTS old_display_value;
ALTER TABLE task_changes DROP COLUMN IF EXISTS new_value_id;
ALTER TABLE task_changes DROP COLUMN IF EXISTS old_value_id;

-- +goose StatementEnd
