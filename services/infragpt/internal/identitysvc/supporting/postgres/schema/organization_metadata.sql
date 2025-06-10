CREATE TABLE organization_metadata (
    organization_id UUID PRIMARY KEY REFERENCES organizations(id),
    company_size VARCHAR(50) NOT NULL,
    team_size VARCHAR(50) NOT NULL,
    use_cases TEXT[] NOT NULL,
    observability_stack TEXT[] NOT NULL,
    completed_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);