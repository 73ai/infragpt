-- name: CreateOrganizationMember :exec
INSERT INTO organization_members (user_id, organization_id, clerk_user_id, clerk_org_id, role)
VALUES ($1, $2, $3, $4, $5);

-- name: DeleteOrganizationMemberByClerkIDs :exec
DELETE FROM organization_members
WHERE clerk_user_id = $1 AND clerk_org_id = $2;

-- name: GetOrganizationMembersByOrganizationID :many
SELECT user_id, organization_id, clerk_user_id, clerk_org_id, role, joined_at
FROM organization_members
WHERE organization_id = $1;

-- name: GetOrganizationMembersByUserClerkID :many
SELECT user_id, organization_id, clerk_user_id, clerk_org_id, role, joined_at
FROM organization_members
WHERE clerk_user_id = $1;