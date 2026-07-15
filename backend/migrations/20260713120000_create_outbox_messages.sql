-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS outbox_messages (
  id            uuid primary key not null default gen_random_uuid(),
  topic         text not null,
  aggregate_id  uuid,
  payload       jsonb not null,
  created_at    timestamp with time zone default current_timestamp not null,
  published_at  timestamp with time zone
);

-- partial index so the relay's "unpublished" scan stays cheap as the table grows
CREATE INDEX IF NOT EXISTS idx_outbox_unpublished
  ON outbox_messages (created_at) WHERE published_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_outbox_unpublished;
DROP TABLE IF EXISTS outbox_messages;
-- +goose StatementEnd
