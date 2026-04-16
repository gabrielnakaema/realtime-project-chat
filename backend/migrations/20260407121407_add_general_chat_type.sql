-- +goose Up
-- +goose StatementBegin
ALTER TABLE chats ADD COLUMN chat_type text NOT NULL DEFAULT 'project';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chats DROP COLUMN chat_type;
-- +goose StatementEnd
