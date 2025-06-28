-- Migration: Add Integration System Tables
-- Run this against the infragpt database
-- This migration adds all tables needed for the integration service

-- Main integrations table
CREATE TABLE IF NOT EXISTS integrations (
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

-- Integration credentials table (encrypted storage)
CREATE TABLE IF NOT EXISTS integration_credentials (
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

-- Unclaimed GitHub App Installations
-- Stores GitHub App installations that have been received via webhook
-- but not yet claimed/configured through the API
CREATE TABLE IF NOT EXISTS unclaimed_installations (
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

-- GitHub Repository Permissions Tracking
-- Tracks repository-level access and permissions for GitHub App installations
CREATE TABLE IF NOT EXISTS github_repositories (
    id UUID PRIMARY KEY,
    integration_id UUID NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
    github_repository_id BIGINT NOT NULL, -- GitHub's internal repository ID
    repository_name VARCHAR(255) NOT NULL,
    repository_full_name VARCHAR(512) NOT NULL, -- org/repo format
    repository_url VARCHAR(512) NOT NULL,
    is_private BOOLEAN NOT NULL DEFAULT false,
    default_branch VARCHAR(255) DEFAULT 'main',
    
    -- Repository-level permissions
    permission_admin BOOLEAN NOT NULL DEFAULT false,
    permission_push BOOLEAN NOT NULL DEFAULT false,
    permission_pull BOOLEAN NOT NULL DEFAULT false,
    
    -- Repository metadata
    repository_description TEXT,
    repository_language VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_synced_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- GitHub repository timestamps
    github_created_at TIMESTAMP,
    github_updated_at TIMESTAMP,
    github_pushed_at TIMESTAMP,
    
    UNIQUE(integration_id, github_repository_id)
);

-- Indexes for integrations table
CREATE INDEX IF NOT EXISTS idx_integrations_org ON integrations (organization_id);
CREATE INDEX IF NOT EXISTS idx_integrations_org_type ON integrations (organization_id, connector_type);
CREATE INDEX IF NOT EXISTS idx_integrations_status ON integrations (status);
CREATE INDEX IF NOT EXISTS idx_integrations_connector_type ON integrations (connector_type);
CREATE INDEX IF NOT EXISTS idx_integrations_bot_id ON integrations (bot_id) WHERE bot_id IS NOT NULL;

-- Indexes for integration_credentials table
CREATE INDEX IF NOT EXISTS idx_credentials_expiring ON integration_credentials (expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_credentials_integration ON integration_credentials (integration_id);

-- Indexes for unclaimed_installations table
CREATE INDEX IF NOT EXISTS idx_unclaimed_installations_github_id ON unclaimed_installations (github_installation_id);
CREATE INDEX IF NOT EXISTS idx_unclaimed_installations_account ON unclaimed_installations (github_account_id, github_account_login);
CREATE INDEX IF NOT EXISTS idx_unclaimed_installations_expires ON unclaimed_installations (expires_at) WHERE claimed_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_unclaimed_installations_unclaimed ON unclaimed_installations (created_at) WHERE claimed_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_unclaimed_installations_app ON unclaimed_installations (github_app_id);

-- Partial index for active unclaimed installations
CREATE INDEX IF NOT EXISTS idx_unclaimed_installations_active ON unclaimed_installations (github_installation_id, github_account_login) 
WHERE claimed_at IS NULL AND suspended_at IS NULL;

-- Indexes for github_repositories table
CREATE INDEX IF NOT EXISTS idx_github_repos_integration ON github_repositories (integration_id);
CREATE INDEX IF NOT EXISTS idx_github_repos_github_id ON github_repositories (github_repository_id);
CREATE INDEX IF NOT EXISTS idx_github_repos_full_name ON github_repositories (repository_full_name);
CREATE INDEX IF NOT EXISTS idx_github_repos_permissions ON github_repositories (integration_id, permission_admin, permission_push, permission_pull);
CREATE INDEX IF NOT EXISTS idx_github_repos_sync ON github_repositories (last_synced_at);
CREATE INDEX IF NOT EXISTS idx_github_repos_private ON github_repositories (is_private);
CREATE INDEX IF NOT EXISTS idx_github_repos_language ON github_repositories (repository_language) WHERE repository_language IS NOT NULL;

-- Comments for documentation
COMMENT ON TABLE integrations IS 'Main integrations table storing connector configurations and metadata';
COMMENT ON TABLE integration_credentials IS 'Encrypted storage for integration credentials (OAuth tokens, API keys, etc.)';
COMMENT ON TABLE unclaimed_installations IS 'Temporary storage for GitHub App installations waiting to be claimed by organizations';
COMMENT ON TABLE github_repositories IS 'Repository-level permissions and metadata for GitHub App integrations';

COMMENT ON COLUMN integrations.connector_type IS 'Type of connector: github, slack, gcp, aws, pagerduty, datadog';
COMMENT ON COLUMN integrations.status IS 'Integration status: active, inactive, pending, not_started, suspended, deleted';
COMMENT ON COLUMN integrations.bot_id IS 'External bot/app ID (e.g., GitHub installation ID, Slack app ID)';
COMMENT ON COLUMN integration_credentials.credential_data_encrypted IS 'AES-256-GCM encrypted credential data';
COMMENT ON COLUMN unclaimed_installations.expires_at IS 'Auto-cleanup timestamp - unclaimed installations expire after 7 days';
COMMENT ON COLUMN github_repositories.last_synced_at IS 'Timestamp of last repository sync operation';