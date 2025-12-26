CREATE TABLE device_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    access_token TEXT UNIQUE NOT NULL,
    refresh_token TEXT UNIQUE NOT NULL,
    organization_id UUID NOT NULL,
    user_id UUID NOT NULL,
    device_name VARCHAR(255),
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP
);

CREATE INDEX idx_device_tokens_access_token ON device_tokens (access_token);
CREATE INDEX idx_device_tokens_refresh_token ON device_tokens (refresh_token);
CREATE INDEX idx_device_tokens_user_id ON device_tokens (user_id);
CREATE INDEX idx_device_tokens_expires_at ON device_tokens (expires_at) WHERE revoked_at IS NULL;
