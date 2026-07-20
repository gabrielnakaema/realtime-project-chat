-- name: InsertOutboxMessage :exec
INSERT INTO outbox_messages (topic, aggregate_id, payload)
VALUES ($1, $2, $3);

-- name: FetchUnpublishedOutboxMessages :many
SELECT id, topic, aggregate_id, payload
FROM outbox_messages
WHERE published_at IS NULL
ORDER BY created_at, id
FOR UPDATE SKIP LOCKED
LIMIT $1;

-- name: MarkOutboxMessagesPublished :exec
UPDATE outbox_messages
SET published_at = current_timestamp
WHERE id = ANY($1::uuid[]);
