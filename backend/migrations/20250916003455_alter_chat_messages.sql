-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

ALTER TABLE chat_messages DROP COLUMN member_id;
ALTER TABLE chat_messages ADD COLUMN user_id uuid;
ALTER TABLE chat_messages ADD COLUMN message_type text NOT NULL DEFAULT 'text';
ALTER TABLE chat_messages ADD CONSTRAINT fk_chat_messages_users FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE chat_members DROP CONSTRAINT chat_members_pkey;
ALTER TABLE chat_members DROP COLUMN id;
ALTER TABLE chat_members ADD PRIMARY KEY (user_id, chat_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

ALTER TABLE chat_messages DROP CONSTRAINT fk_chat_messages_users;
ALTER TABLE chat_messages DROP COLUMN user_id;
ALTER TABLE chat_messages ADD COLUMN member_id uuid NOT NULL;
ALTER TABLE chat_messages DROP COLUMN message_type;

ALTER TABLE chat_members DROP CONSTRAINT chat_members_pkey;
ALTER TABLE chat_members ADD COLUMN id uuid NOT NULL DEFAULT gen_random_uuid();

-- +goose StatementEnd
