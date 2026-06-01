-- name: CreateMCPAPIKey :one
INSERT INTO mcp_api_keys (user_id, name, key_prefix, secret_hash, created_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: CreateMCPAPIKeyScope :exec
INSERT INTO mcp_api_key_scopes (api_key_id, scope, created_at)
VALUES ($1, $2, $3);

-- name: ListMCPAPIKeysByUserID :many
SELECT
  k.id,
  k.user_id,
  k.name,
  k.key_prefix,
  k.secret_hash,
  k.created_at,
  k.last_used_at,
  k.revoked_at,
  COALESCE(array_agg(s.scope ORDER BY s.scope) FILTER (WHERE s.scope IS NOT NULL), '{}'::text[]) AS scopes
FROM mcp_api_keys k
LEFT JOIN mcp_api_key_scopes s ON s.api_key_id = k.id
WHERE k.user_id = $1
GROUP BY k.id
ORDER BY k.created_at DESC, k.id DESC;

-- name: GetMCPAPIKeyByIDForUser :one
SELECT
  k.id,
  k.user_id,
  k.name,
  k.key_prefix,
  k.secret_hash,
  k.created_at,
  k.last_used_at,
  k.revoked_at,
  COALESCE(array_agg(s.scope ORDER BY s.scope) FILTER (WHERE s.scope IS NOT NULL), '{}'::text[]) AS scopes
FROM mcp_api_keys k
LEFT JOIN mcp_api_key_scopes s ON s.api_key_id = k.id
WHERE k.id = $1 AND k.user_id = $2
GROUP BY k.id;

-- name: GetMCPAPIKeyByPrefix :one
SELECT
  k.id,
  k.user_id,
  k.name,
  k.key_prefix,
  k.secret_hash,
  k.created_at,
  k.last_used_at,
  k.revoked_at,
  COALESCE(array_agg(s.scope ORDER BY s.scope) FILTER (WHERE s.scope IS NOT NULL), '{}'::text[]) AS scopes
FROM mcp_api_keys k
LEFT JOIN mcp_api_key_scopes s ON s.api_key_id = k.id
WHERE k.key_prefix = $1
GROUP BY k.id;

-- name: RevokeMCPAPIKey :execrows
UPDATE mcp_api_keys
SET revoked_at = COALESCE(revoked_at, sqlc.arg('revoked_at')::timestamptz)
WHERE id = $1 AND user_id = $2;

-- name: TouchMCPAPIKeyLastUsedAt :exec
UPDATE mcp_api_keys
SET last_used_at = sqlc.arg('last_used_at')::timestamptz
WHERE id = $1 AND (last_used_at IS NULL OR last_used_at < sqlc.arg('cutoff_at')::timestamptz);
