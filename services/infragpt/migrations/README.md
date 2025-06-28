# Database Migrations

This directory contains SQL migration files for the InfraGPT database schema.

## Migration Files

- `001_identity_tables.sql` - Identity system tables (users, organizations, members)
- `002_integration_tables.sql` - Integration system tables (integrations, credentials, GitHub repositories)

## Running Migrations

### Prerequisites
- PostgreSQL database running
- Database connection configured in `config.yaml`

### Manual Execution
Run migrations manually using `psql` or your preferred PostgreSQL client:

```bash
# Connect to your database
psql -h localhost -U your_user -d infragpt

# Run migrations in order
\i migrations/001_identity_tables.sql
\i migrations/002_integration_tables.sql
```

### Automated Execution
You can also run migrations programmatically using Go's database/sql package or a migration tool like [golang-migrate](https://github.com/golang-migrate/migrate).

## Migration Order

⚠️ **Important**: Always run migrations in numerical order to ensure proper schema evolution.

1. First run `001_identity_tables.sql`
2. Then run `002_integration_tables.sql`

## Schema Overview

### Identity System (001)
- `users` - User accounts synced from Clerk
- `organizations` - Organizations synced from Clerk  
- `organization_metadata` - Onboarding and configuration data
- `organization_members` - User-organization relationships

### Integration System (002)
- `integrations` - Main integration configurations
- `integration_credentials` - Encrypted credential storage
- `unclaimed_installations` - Temporary GitHub App installation storage
- `github_repositories` - Repository permissions and metadata tracking

## Security Notes

- All credential data is encrypted using AES-256-GCM
- Sensitive fields use appropriate data types and constraints
- Proper foreign key relationships ensure data integrity
- Indexes are optimized for common query patterns

## Development

When adding new tables or modifying existing ones:

1. Create a new migration file with incremented number (e.g., `003_new_feature.sql`)
2. Include both table creation and index creation
3. Add appropriate comments for documentation
4. Test the migration on a local database first
5. Update this README with the new migration details