-- name: CreateUser :exec
insert into users (user_id, email, password_hash)
values ($1, $2, $3);

-- name: CreateEmailVerification :exec
insert into email_verification (verification_id, user_id, email, expiry_at)
values ($1, $2, $3, $4);

-- name: UserByEmail :one
select user_id, email, password_hash from users where email = $1;

-- name: UserByID :one
select user_id, email, password_hash, is_email_verified, created_at, updated_at from users where user_id = $1;

-- name: VerifyEmail :exec
update users set is_email_verified = true, updated_at = now() where user_id = $1;

-- name: EmailVerification :one
select verification_id, user_id, email, expiry_at from email_verification where verification_id = $1;

-- name: MarkEmailVerificationAsExpired :exec
update email_verification set expiry_at = now() where verification_id = $1;

-- name: CreateResetPassword :exec
insert into reset_password (reset_id, user_id, expiry_at)
values ($1, $2, $3);

-- name: ResetPassword :one
select reset_id, user_id, expiry_at from reset_password where reset_id = $1;

-- name: SetNewPassword :exec
update users set password_hash = $2, updated_at = now() where user_id = $1;

-- name: MarkResetPasswordAsExpired :exec
update reset_password set expiry_at = now() where reset_id = $1;