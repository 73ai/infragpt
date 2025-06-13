-- name: slackToken :one
SELECT token FROM slack_token WHERE team_id = $1 and expired='f';

-- name: saveSlackToken :exec
INSERT INTO slack_token (token_id, team_id, token) VALUES ($1, $2, $3);

-- name: integrations :many
SELECT * FROM integration WHERE business_id = $1 and active='t';

-- name: saveIntegration :exec
INSERT INTO integration (id, provider, status, business_id, provider_project_id) VALUES ($1, $2, $3, $4, $5);