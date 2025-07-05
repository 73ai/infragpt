-- Migration: Add Identity Tables for Clerk Integration
-- Run this against the backend database

-- Users table (synced from Clerk)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clerk_user_id VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Organizations table (synced from Clerk + own ID)
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clerk_org_id VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    created_by_user_id UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Organization metadata (onboarding data)
CREATE TABLE IF NOT EXISTS organization_metadata (
    organization_id UUID PRIMARY KEY REFERENCES organizations(id),
    company_size VARCHAR(50) NOT NULL,
    team_size VARCHAR(50) NOT NULL,
    use_cases TEXT[] NOT NULL,
    observability_stack TEXT[] NOT NULL,
    completed_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Organization members (synced from Clerk)
CREATE TABLE IF NOT EXISTS organization_members (
    user_id UUID REFERENCES users(id),
    organization_id UUID REFERENCES organizations(id),
    clerk_user_id VARCHAR(255) NOT NULL,
    clerk_org_id VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    joined_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (user_id, organization_id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_clerk_user_id ON users(clerk_user_id);
CREATE INDEX IF NOT EXISTS idx_organizations_clerk_org_id ON organizations(clerk_org_id);
CREATE INDEX IF NOT EXISTS idx_organization_members_clerk_user_id ON organization_members(clerk_user_id);
CREATE INDEX IF NOT EXISTS idx_organization_members_clerk_org_id ON organization_members(clerk_org_id);