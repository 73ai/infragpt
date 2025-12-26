-- name: CreateDeviceCode :exec
INSERT INTO device_codes (id, device_code, user_code, status, expires_at, created_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetDeviceCodeByUserCode :one
SELECT id, device_code, user_code, status, organization_id, user_id, expires_at, created_at
FROM device_codes
WHERE user_code = $1;

-- name: GetDeviceCodeByDeviceCode :one
SELECT id, device_code, user_code, status, organization_id, user_id, expires_at, created_at
FROM device_codes
WHERE device_code = $1;

-- name: AuthorizeDeviceCode :exec
UPDATE device_codes
SET status = 'authorized', organization_id = $2, user_id = $3
WHERE user_code = $1 AND status = 'pending';

-- name: MarkDeviceCodeAsUsed :exec
UPDATE device_codes
SET status = 'used'
WHERE device_code = $1 AND status = 'authorized';

-- name: DeleteExpiredDeviceCodes :exec
DELETE FROM device_codes
WHERE expires_at < NOW();
