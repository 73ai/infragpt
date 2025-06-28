-- GitHub App Installation Details
-- Stores GitHub-specific installation metadata linked to integrations

CREATE TABLE github_installations (
    id UUID PRIMARY KEY,
    integration_id UUID NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
    github_installation_id BIGINT NOT NULL UNIQUE, -- GitHub's installation ID
    github_app_id BIGINT NOT NULL,
    
    -- Account information
    github_account_id BIGINT NOT NULL,
    github_account_login VARCHAR(255) NOT NULL,
    github_account_type VARCHAR(50) NOT NULL, -- 'User' or 'Organization'
    
    -- Installation configuration
    repository_selection VARCHAR(20) NOT NULL, -- 'all' or 'selected'
    permissions JSONB NOT NULL, -- Current app permissions
    events TEXT[], -- Subscribed webhook events
    
    -- Installation state
    app_slug VARCHAR(255),
    target_type VARCHAR(50), -- 'Organization' or 'User'
    
    -- Access and webhook URLs
    access_tokens_url VARCHAR(512),
    repositories_url VARCHAR(512),
    html_url VARCHAR(512),
    
    -- Suspension tracking
    suspended_at TIMESTAMP,
    suspended_by JSONB, -- User who suspended the installation
    
    -- Repository tracking summary
    total_repository_count INTEGER DEFAULT 0,
    accessible_repository_count INTEGER DEFAULT 0,
    last_repository_sync_at TIMESTAMP,
    
    -- Permission tracking
    permissions_updated_at TIMESTAMP,
    permission_version INTEGER DEFAULT 1, -- Track permission changes
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    github_created_at TIMESTAMP NOT NULL,
    github_updated_at TIMESTAMP NOT NULL,
    
    UNIQUE(integration_id)
);

-- Indexes for performance
CREATE INDEX idx_github_installations_integration ON github_installations (integration_id);
CREATE INDEX idx_github_installations_github_id ON github_installations (github_installation_id);
CREATE INDEX idx_github_installations_account ON github_installations (github_account_id, github_account_login);
CREATE INDEX idx_github_installations_repo_sync ON github_installations (last_repository_sync_at);
CREATE INDEX idx_github_installations_permissions ON github_installations (permissions_updated_at, permission_version);