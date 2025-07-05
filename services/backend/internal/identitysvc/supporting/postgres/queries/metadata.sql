-- name: CreateOrganizationMetadata :exec
INSERT INTO organization_metadata (organization_id, company_size, team_size, use_cases, observability_stack)
VALUES ($1, $2, $3, $4, $5);

-- name: GetOrganizationMetadataByOrganizationID :one
SELECT organization_id, company_size, team_size, use_cases, observability_stack, completed_at, updated_at
FROM organization_metadata
WHERE organization_id = $1;

-- name: UpdateOrganizationMetadata :exec
UPDATE organization_metadata
SET company_size = $2, team_size = $3, use_cases = $4, observability_stack = $5, updated_at = NOW()
WHERE organization_id = $1;

-- name: DeleteOrganizationMetadataByOrganizationID :exec
DELETE FROM organization_metadata
WHERE organization_id = $1;