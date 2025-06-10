-- name: CreateOrganization :exec
INSERT INTO organizations (clerk_org_id, name, slug, created_by_user_id)
VALUES ($1, $2, $3, $4);

-- name: GetOrganizationByClerkID :one
SELECT id, clerk_org_id, name, slug, created_by_user_id, created_at, updated_at
FROM organizations
WHERE clerk_org_id = $1;

-- name: GetOrganizationsByUserClerkID :many
SELECT o.id, o.clerk_org_id, o.name, o.slug, o.created_by_user_id, o.created_at, o.updated_at
FROM organizations o
INNER JOIN organization_members om ON o.id = om.organization_id
WHERE om.clerk_user_id = $1;

-- name: UpdateOrganization :exec
UPDATE organizations
SET name = $2, slug = $3, updated_at = NOW()
WHERE clerk_org_id = $1;

-- name: DeleteOrganizationByClerkID :exec
DELETE FROM organizations
WHERE clerk_org_id = $1;