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


-- Indexes for github_repositories table
CREATE INDEX IF NOT EXISTS idx_github_repos_integration ON github_repositories (integration_id);
CREATE INDEX IF NOT EXISTS idx_github_repos_github_id ON github_repositories (github_repository_id);
CREATE INDEX IF NOT EXISTS idx_github_repos_full_name ON github_repositories (repository_full_name);
CREATE INDEX IF NOT EXISTS idx_github_repos_permissions ON github_repositories (integration_id, permission_admin, permission_push, permission_pull);
CREATE INDEX IF NOT EXISTS idx_github_repos_sync ON github_repositories (last_synced_at);
CREATE INDEX IF NOT EXISTS idx_github_repos_private ON github_repositories (is_private);
CREATE INDEX IF NOT EXISTS idx_github_repos_language ON github_repositories (repository_language) WHERE repository_language IS NOT NULL;