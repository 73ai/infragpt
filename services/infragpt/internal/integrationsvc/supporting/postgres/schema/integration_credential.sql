CREATE TABLE integration_credentials (
    id UUID PRIMARY KEY,
    integration_id UUID NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
    credential_type VARCHAR(50) NOT NULL,
    credential_data_encrypted TEXT NOT NULL,
    expires_at TIMESTAMP,
    encryption_key_id VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    UNIQUE(integration_id)
);

CREATE INDEX idx_credentials_expiring ON integration_credentials (expires_at) WHERE expires_at IS NOT NULL;