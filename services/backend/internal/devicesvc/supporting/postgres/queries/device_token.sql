-- name: CreateDeviceToken :exec
INSERT INTO device_tokens (id, access_token, refresh_token, organization_id, user_id, device_name, expires_at, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetDeviceTokenByAccessToken :one
SELECT id, access_token, refresh_token, organization_id, user_id, device_name, expires_at, created_at, revoked_at
FROM device_tokens
WHERE access_token = $1;

-- name: GetDeviceTokenByRefreshToken :one
SELECT id, access_token, refresh_token, organization_id, user_id, device_name, expires_at, created_at, revoked_at
FROM device_tokens
WHERE refresh_token = $1;

-- name: RevokeDeviceToken :exec
UPDATE device_tokens
SET revoked_at = NOW()
WHERE access_token = $1 AND revoked_at IS NULL;

-- name: RevokeAllDeviceTokensForUser :exec
UPDATE device_tokens
SET revoked_at = NOW()
WHERE user_id = $1 AND revoked_at IS NULL;

-- name: UpdateDeviceTokens :exec
UPDATE device_tokens
SET access_token = $2, refresh_token = $3, expires_at = $4
WHERE refresh_token = $1 AND revoked_at IS NULL;
