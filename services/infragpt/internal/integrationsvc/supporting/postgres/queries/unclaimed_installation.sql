-- Unclaimed Installation Queries

-- name: StoreUnclaimedInstallation :exec
INSERT INTO unclaimed_installations (
    id, github_installation_id, github_app_id, github_account_id,
    github_account_login, github_account_type, repository_selection,
    permissions, events, access_tokens_url, repositories_url, html_url,
    app_slug, suspended_at, suspended_by, webhook_sender, raw_webhook_payload,
    created_at, github_created_at, github_updated_at, expires_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21);

-- name: FindUnclaimedInstallationByInstallationID :one
SELECT id, github_installation_id, github_app_id, github_account_id,
    github_account_login, github_account_type, repository_selection,
    permissions, events, access_tokens_url, repositories_url, html_url,
    app_slug, suspended_at, suspended_by, webhook_sender, raw_webhook_payload,
    created_at, github_created_at, github_updated_at, expires_at,
    claimed_at, claimed_by_organization_id, claimed_by_user_id
FROM unclaimed_installations 
WHERE github_installation_id = $1;

-- name: MarkUnclaimedInstallationAsClaimed :exec
UPDATE unclaimed_installations 
SET claimed_at = NOW(), 
    claimed_by_organization_id = $1, 
    claimed_by_user_id = $2
WHERE github_installation_id = $3;

-- name: DeleteExpiredUnclaimedInstallations :exec
DELETE FROM unclaimed_installations 
WHERE expires_at < NOW() AND claimed_at IS NULL;

-- name: FindUnclaimedInstallations :many
SELECT id, github_installation_id, github_app_id, github_account_id,
    github_account_login, github_account_type, repository_selection,
    permissions, events, access_tokens_url, repositories_url, html_url,
    app_slug, suspended_at, suspended_by, webhook_sender, raw_webhook_payload,
    created_at, github_created_at, github_updated_at, expires_at,
    claimed_at, claimed_by_organization_id, claimed_by_user_id
FROM unclaimed_installations 
WHERE claimed_at IS NULL 
ORDER BY created_at DESC 
LIMIT $1;

-- name: DeleteUnclaimedInstallation :exec
DELETE FROM unclaimed_installations 
WHERE github_installation_id = $1;