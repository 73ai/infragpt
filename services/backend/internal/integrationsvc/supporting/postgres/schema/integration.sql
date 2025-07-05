CREATE TABLE integrations (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL,
    user_id UUID NOT NULL,
    connector_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    bot_id VARCHAR(255),
    connector_user_id VARCHAR(255),
    connector_organization_id VARCHAR(255),
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMP,
    
    UNIQUE(organization_id, connector_type)
);

CREATE INDEX idx_integrations_org ON integrations (organization_id);
CREATE INDEX idx_integrations_org_type ON integrations (organization_id, connector_type);
CREATE INDEX idx_integrations_status ON integrations (status);