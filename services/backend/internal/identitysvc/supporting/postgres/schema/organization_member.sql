CREATE TABLE organization_members (
    user_id UUID REFERENCES users(id),
    organization_id UUID REFERENCES organizations(id),
    clerk_user_id VARCHAR(255) NOT NULL,
    clerk_org_id VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    joined_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (user_id, organization_id)
);