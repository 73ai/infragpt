-- Channels table - track which channels bot monitors
CREATE TABLE channels (
    channel_id VARCHAR(36) NOT NULL,
    team_id VARCHAR(36) NOT NULL,
    channel_name VARCHAR(255),
    is_monitored BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (team_id, channel_id)
);

CREATE INDEX idx_channels_team_monitored ON channels(team_id, is_monitored);