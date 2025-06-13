-- name: StoreCredential :exec
INSERT INTO integration_credentials (
    id, integration_id, credential_type, credential_data_encrypted,
    expires_at, encryption_key_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
);

-- name: FindCredentialByIntegration :one
SELECT id, integration_id, credential_type, credential_data_encrypted,
       expires_at, encryption_key_id, created_at, updated_at
FROM integration_credentials
WHERE integration_id = $1;

-- name: UpdateCredential :exec
UPDATE integration_credentials
SET credential_type = $2,
    credential_data_encrypted = $3,
    expires_at = $4,
    encryption_key_id = $5,
    updated_at = NOW()
WHERE integration_id = $1;

-- name: DeleteCredential :exec
DELETE FROM integration_credentials
WHERE integration_id = $1;

-- name: FindExpiringCredentials :many
SELECT id, integration_id, credential_type, credential_data_encrypted,
       expires_at, encryption_key_id, created_at, updated_at
FROM integration_credentials
WHERE expires_at IS NOT NULL AND expires_at < $1
ORDER BY expires_at ASC;