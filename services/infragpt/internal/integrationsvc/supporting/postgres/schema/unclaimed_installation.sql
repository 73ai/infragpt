-- Unclaimed GitHub App Installations
-- Stores GitHub App installations that have been received via webhook
-- but not yet claimed/configured through the API

CREATE TABLE unclaimed_installations (
    id UUID PRIMARY KEY,
    github_installation_id VARCHAR(50) NOT NULL UNIQUE, -- GitHub's installation ID
    github_app_id BIGINT NOT NULL,
    
    -- Account/Organization information from GitHub
    github_account_id BIGINT NOT NULL,
    github_account_login VARCHAR(255) NOT NULL,
    github_account_type VARCHAR(50) NOT NULL, -- 'User' or 'Organization'
    
    -- Installation metadata
    repository_selection VARCHAR(20) NOT NULL, -- 'all' or 'selected'
    permissions JSONB NOT NULL, -- Store GitHub permissions as JSON
    events TEXT[], -- Array of subscribed events
    
    -- Installation URLs
    access_tokens_url VARCHAR(512),
    repositories_url VARCHAR(512),
    html_url VARCHAR(512),
    
    -- Installation state
    app_slug VARCHAR(255),
    suspended_at TIMESTAMP,
    suspended_by JSONB, -- User object who suspended the installation
    
    -- Webhook metadata
    webhook_sender JSONB, -- User who triggered the installation
    raw_webhook_payload JSONB, -- Store full webhook payload for debugging
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    github_created_at TIMESTAMP NOT NULL,
    github_updated_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL DEFAULT (NOW() + INTERVAL '7 days'), -- Auto-cleanup after 7 days
    
    -- Tracking
    claimed_at TIMESTAMP,
    claimed_by_organization_id UUID,
    claimed_by_user_id UUID
);

-- Indexes for performance and cleanup
CREATE INDEX idx_unclaimed_installations_github_id ON unclaimed_installations (github_installation_id);
CREATE INDEX idx_unclaimed_installations_account ON unclaimed_installations (github_account_id, github_account_login);
CREATE INDEX idx_unclaimed_installations_expires ON unclaimed_installations (expires_at) WHERE claimed_at IS NULL;
CREATE INDEX idx_unclaimed_installations_unclaimed ON unclaimed_installations (created_at) WHERE claimed_at IS NULL;
CREATE INDEX idx_unclaimed_installations_app ON unclaimed_installations (github_app_id);

-- Partial index for active unclaimed installations
CREATE INDEX idx_unclaimed_installations_active ON unclaimed_installations (github_installation_id, github_account_login) 
WHERE claimed_at IS NULL AND suspended_at IS NULL;