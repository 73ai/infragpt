-- Conversations table - tracks all conversations/threads
CREATE TABLE conversations (
    conversation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id VARCHAR(36) NOT NULL,
    channel_id VARCHAR(36) NOT NULL,
    thread_ts VARCHAR(36) NOT NULL, -- Slack thread timestamp (unique per channel)
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(team_id, channel_id, thread_ts)
);

CREATE INDEX idx_conversations_team_channel ON conversations(team_id, channel_id);
CREATE INDEX idx_conversations_thread_ts ON conversations(team_id, channel_id, thread_ts);