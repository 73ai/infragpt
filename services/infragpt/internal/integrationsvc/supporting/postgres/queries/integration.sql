-- name: StoreIntegration :exec
INSERT INTO integrations (
    id, organization_id, user_id, connector_type, status, 
    bot_id, connector_user_id, connector_organization_id, 
    metadata, created_at, updated_at, last_used_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
);

-- name: FindIntegrationByID :one
SELECT id, organization_id, user_id, connector_type, status,
       bot_id, connector_user_id, connector_organization_id,
       metadata, created_at, updated_at, last_used_at
FROM integrations
WHERE id = $1;

-- name: FindIntegrationsByOrganization :many
SELECT id, organization_id, user_id, connector_type, status,
       bot_id, connector_user_id, connector_organization_id,
       metadata, created_at, updated_at, last_used_at
FROM integrations
WHERE organization_id = $1
ORDER BY created_at DESC;

-- name: FindIntegrationsByOrganizationAndType :many
SELECT id, organization_id, user_id, connector_type, status,
       bot_id, connector_user_id, connector_organization_id,
       metadata, created_at, updated_at, last_used_at
FROM integrations
WHERE organization_id = $1 AND connector_type = $2
ORDER BY created_at DESC;

-- name: UpdateIntegrationStatus :exec
UPDATE integrations
SET status = $2, updated_at = NOW()
WHERE id = $1;

-- name: UpdateIntegrationLastUsed :exec
UPDATE integrations
SET last_used_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- name: DeleteIntegration :exec
DELETE FROM integrations
WHERE id = $1;

-- name: FindIntegrationByBotIDAndType :one
SELECT id, organization_id, user_id, connector_type, status,
       bot_id, connector_user_id, connector_organization_id,
       metadata, created_at, updated_at, last_used_at
FROM integrations
WHERE bot_id = $1 AND connector_type = $2;

-- name: UpdateIntegrationMetadata :exec
UPDATE integrations
SET metadata = $2, updated_at = NOW()
WHERE id = $1;