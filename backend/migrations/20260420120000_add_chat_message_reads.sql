-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS chat_message_reads (
	chat_message_id uuid NOT NULL,
	user_id uuid NOT NULL,
	read_at timestamp with time zone NOT NULL DEFAULT current_timestamp,
	PRIMARY KEY (chat_message_id, user_id),
	CONSTRAINT fk_chat_message_reads_messages FOREIGN KEY (chat_message_id) REFERENCES chat_messages(id) ON DELETE CASCADE,
	CONSTRAINT fk_chat_message_reads_users FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_chat_message_reads_user_id ON chat_message_reads (user_id);
CREATE INDEX IF NOT EXISTS idx_chat_message_reads_read_at ON chat_message_reads (read_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS chat_message_reads;
-- +goose StatementEnd
