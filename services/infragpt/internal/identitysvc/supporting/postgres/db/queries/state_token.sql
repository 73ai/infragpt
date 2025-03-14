-- name: CreateStateToken :exec
insert into  google_state_token (token, expires_at) values ($1, $2);

-- name: StateToken :one
select * from google_state_token where token = $1;

-- name: RevokeStateToken :exec
update google_state_token set revoked = true, revoked_at = now() where token = $1;
