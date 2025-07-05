-- GitHub Repository Permissions Tracking
-- Tracks repository-level access and permissions for GitHub App installations

CREATE TABLE github_repositories (
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

-- Indexes for performance
CREATE INDEX idx_github_repos_integration ON github_repositories (integration_id);
CREATE INDEX idx_github_repos_github_id ON github_repositories (github_repository_id);
CREATE INDEX idx_github_repos_full_name ON github_repositories (repository_full_name);
CREATE INDEX idx_github_repos_permissions ON github_repositories (integration_id, permission_admin, permission_push, permission_pull);
CREATE INDEX idx_github_repos_sync ON github_repositories (last_synced_at);