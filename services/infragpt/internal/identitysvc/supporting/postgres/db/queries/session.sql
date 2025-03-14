-- name: CreateSession :exec
insert into user_session (user_id, device_id, session_id, user_agent,
ip_address, ip_country_iso, timezone)
values ($1, $2, $3, $4, $5, $6, $7);

-- name: CreateRefreshToken :exec
insert into refresh_token (token_id, user_id, session_id, token_hash, expiry_at)
values ($1, $2, $3, $4, $5);

-- name: CreateDevice :exec
insert into device (device_id, user_id, device_fingerprint, name, os, brand)
values ($1, $2, $3, $4, $5, $6);

-- name: RevokeRefreshToken :exec
update refresh_token set revoked = true, expiry_at = now() where token_id = $1;

-- name: RefreshToken :one
select token_id, user_id, session_id, token_hash, expiry_at, created_at, revoked
from refresh_token where token_id = $1 and revoked = false;

-- name: UserSessions :many
select user_id, device_id, session_id, user_agent, ip_address, ip_country_iso,
last_activity_at, created_at, timezone, is_expired
from user_session where user_id = $1 and is_expired = false;

-- name: DevicesByUserID :many
select device_id, user_id, device_fingerprint, name, os, brand
from device where user_id = $1;

-- name: UserSession :one
select user_id, device_id, session_id, user_agent, ip_address, ip_country_iso,
last_activity_at, created_at, timezone, is_expired
from user_session where session_id = $1 and is_expired = false;

-- name: Device :one
select device_id, user_id, device_fingerprint, name, os, brand
from device where device_id = $1;