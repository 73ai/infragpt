-- Messages table - stores all messages in conversations
CREATE TABLE messages (
    message_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(conversation_id) ON DELETE CASCADE,
    slack_message_ts VARCHAR(36) NOT NULL, -- Individual message timestamp
    sender_user_id VARCHAR(36) NOT NULL,
    sender_username VARCHAR(255),
    sender_email VARCHAR(255),
    sender_name VARCHAR(255),
    message_text TEXT NOT NULL,
    is_bot_message BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(conversation_id, slack_message_ts)
);

CREATE INDEX idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX idx_messages_created_at ON messages(conversation_id, created_at DESC);