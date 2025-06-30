-- GitHub Repository Queries

-- name: UpsertGitHubRepository :exec
INSERT INTO github_repositories (
    id, integration_id, github_repository_id, repository_name,
    repository_full_name, repository_url, is_private, default_branch,
    permission_admin, permission_push, permission_pull,
    repository_description, repository_language, created_at, updated_at,
    last_synced_at, github_created_at, github_updated_at, github_pushed_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
ON CONFLICT (integration_id, github_repository_id)
DO UPDATE SET
    repository_name = EXCLUDED.repository_name,
    repository_full_name = EXCLUDED.repository_full_name,
    repository_url = EXCLUDED.repository_url,
    is_private = EXCLUDED.is_private,
    default_branch = EXCLUDED.default_branch,
    permission_admin = EXCLUDED.permission_admin,
    permission_push = EXCLUDED.permission_push,
    permission_pull = EXCLUDED.permission_pull,
    repository_description = EXCLUDED.repository_description,
    repository_language = EXCLUDED.repository_language,
    updated_at = EXCLUDED.updated_at,
    last_synced_at = EXCLUDED.last_synced_at,
    github_updated_at = EXCLUDED.github_updated_at,
    github_pushed_at = EXCLUDED.github_pushed_at;

-- name: FindGitHubRepositoriesByIntegrationID :many
SELECT id, integration_id, github_repository_id, repository_name,
    repository_full_name, repository_url, is_private, default_branch,
    permission_admin, permission_push, permission_pull,
    repository_description, repository_language, created_at, updated_at,
    last_synced_at, github_created_at, github_updated_at, github_pushed_at
FROM github_repositories 
WHERE integration_id = $1
ORDER BY repository_full_name;

-- name: FindGitHubRepositoryByGitHubID :one
SELECT id, integration_id, github_repository_id, repository_name,
    repository_full_name, repository_url, is_private, default_branch,
    permission_admin, permission_push, permission_pull,
    repository_description, repository_language, created_at, updated_at,
    last_synced_at, github_created_at, github_updated_at, github_pushed_at
FROM github_repositories 
WHERE integration_id = $1 AND github_repository_id = $2;

-- name: DeleteGitHubRepositoryByGitHubID :exec
DELETE FROM github_repositories 
WHERE integration_id = $1 AND github_repository_id = $2;

-- name: UpdateGitHubRepositoryPermissions :exec
UPDATE github_repositories 
SET permission_admin = $1, 
    permission_push = $2, 
    permission_pull = $3,
    updated_at = NOW()
WHERE integration_id = $4 AND github_repository_id = $5;

-- name: BulkDeleteGitHubRepositories :exec
DELETE FROM github_repositories 
WHERE integration_id = $1 AND github_repository_id = ANY($2::bigint[]);

-- name: UpdateGitHubRepositoryLastSyncTime :exec
UPDATE github_repositories 
SET last_synced_at = $1, updated_at = NOW()
WHERE integration_id = $2;