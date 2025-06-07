-- name: CreateUser :exec
INSERT INTO users (clerk_user_id, email, first_name, last_name)
VALUES ($1, $2, $3, $4);

-- name: GetUserByClerkID :one
SELECT id, clerk_user_id, email, first_name, last_name, created_at, updated_at
FROM users
WHERE clerk_user_id = $1;

-- name: UpdateUser :exec
UPDATE users
SET email = $2, first_name = $3, last_name = $4, updated_at = NOW()
WHERE clerk_user_id = $1;