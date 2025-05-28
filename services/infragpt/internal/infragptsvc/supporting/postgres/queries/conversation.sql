
-- name: CreateConversation :one
INSERT INTO conversations (team_id, channel_id, thread_ts)
VALUES ($1, $2, $3)
RETURNING conversation_id, team_id, channel_id, thread_ts, created_at, updated_at;

-- name: GetConversationByThread :one
SELECT conversation_id, team_id, channel_id, thread_ts, created_at, updated_at
FROM conversations
WHERE team_id = $1 AND channel_id = $2 AND thread_ts = $3;

-- name: UpdateConversationTimestamp :exec
UPDATE conversations
SET updated_at = NOW()
WHERE conversation_id = $1;

-- name: StoreMessage :one
INSERT INTO messages (conversation_id, slack_message_ts, sender_user_id, sender_username, sender_email, sender_name, message_text, is_bot_message)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING message_id, conversation_id, slack_message_ts, sender_user_id, sender_username, sender_email, sender_name, message_text, is_bot_message, created_at;

-- name: MessageBySlackTS :one
SELECT message_id, conversation_id, slack_message_ts, sender_user_id, sender_username, sender_email, sender_name, message_text, is_bot_message, created_at
FROM messages
WHERE conversation_id = $1 AND slack_message_ts = $2 AND sender_user_id = $3;

-- name: GetConversationHistory :many
SELECT message_id, conversation_id, slack_message_ts, sender_user_id, sender_username, sender_email, sender_name, message_text, is_bot_message, created_at
FROM messages
WHERE conversation_id = $1
ORDER BY created_at ASC;

-- name: GetConversationHistoryDesc :many
SELECT message_id, conversation_id, slack_message_ts, sender_user_id, sender_username, sender_email, sender_name, message_text, is_bot_message, created_at
FROM messages
WHERE conversation_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: AddChannel :exec
INSERT INTO channels (team_id, channel_id, channel_name, is_monitored)
VALUES ($1, $2, $3, false)
ON CONFLICT (team_id, channel_id) 
DO UPDATE SET channel_name = EXCLUDED.channel_name;

-- name: SetChannelMonitoring :exec
UPDATE channels
SET is_monitored = $3
WHERE team_id = $1 AND channel_id = $2;

-- name: GetMonitoredChannels :many
SELECT channel_id, team_id, channel_name, is_monitored, created_at
FROM channels
WHERE team_id = $1 AND is_monitored = true;

-- name: IsChannelMonitored :one
SELECT COALESCE(is_monitored, false) as is_monitored
FROM channels
WHERE team_id = $1 AND channel_id = $2;