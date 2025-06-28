# Database Guide

How to work with PostgreSQL and SQLC in InfraGPT.

## Overview

InfraGPT uses **PostgreSQL** with **SQLC** for type-safe database operations. Each service has its own database schema but shares the same connection.

## Database Structure

```
PostgreSQL Database: "infragpt"
├── Identity Service Tables
│   ├── users
│   ├── organizations
│   ├── organization_members
│   └── organization_metadata
├── Integration Service Tables
│   ├── integrations
│   └── integration_credentials
└── Conversation Service Tables
    ├── conversations
    ├── channels
    └── messages
```

## How SQLC Works

**The Flow**:
1. Write SQL in `queries/*.sql` files
2. Run `sqlc generate` 
3. Get type-safe Go code in `*.sql.go` files
4. Use generated functions in your repositories

**Key files**:
- `sqlc.json` - Configuration for all services
- `schema/*.sql` - Table definitions  
- `queries/*.sql` - SQL queries
- `*.sql.go` - Generated Go code (never edit!)

## Writing Queries

### Basic Query Patterns

```sql
-- In internal/servicename/supporting/postgres/queries/users.sql

-- name: CreateUser :one
INSERT INTO users (id, name, email, created_at)
VALUES ($1, $2, $3, NOW())
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: ListUsersByOrganization :many
SELECT * FROM users 
WHERE organization_id = $1
ORDER BY created_at DESC;

-- name: UpdateUserName :exec
UPDATE users SET name = $2, updated_at = NOW()
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;
```

### Query Types

**`:one`** - Returns a single row
```sql
-- name: GetUser :one
SELECT * FROM users WHERE id = $1;
```

**`:many`** - Returns multiple rows
```sql
-- name: ListUsers :many
SELECT * FROM users ORDER BY created_at;
```

**`:exec`** - Executes without returning data
```sql
-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;
```

## Using Generated Code

After running `sqlc generate`, use the generated functions:

```go
// In your repository
func (r *userRepository) GetByID(ctx context.Context, id string) (User, error) {
    row, err := r.queries.GetUserByID(ctx, id)
    if err != nil {
        return User{}, fmt.Errorf("failed to get user %s: %w", id, err)
    }
    
    return User{
        ID:    row.ID,
        Name:  row.Name,
        Email: row.Email,
    }, nil
}

func (r *userRepository) Create(ctx context.Context, user User) error {
    _, err := r.queries.CreateUser(ctx, CreateUserParams{
        ID:    user.ID,
        Name:  user.Name,
        Email: user.Email,
    })
    if err != nil {
        return fmt.Errorf("failed to create user: %w", err)
    }
    return nil
}
```

## Schema Management

### Adding New Tables

1. **Create schema file**:
   ```sql
   -- In internal/servicename/supporting/postgres/schema/new_table.sql
   CREATE TABLE new_table (
       id UUID PRIMARY KEY,
       name VARCHAR(255) NOT NULL,
       created_at TIMESTAMP NOT NULL DEFAULT NOW(),
       updated_at TIMESTAMP NOT NULL DEFAULT NOW()
   );
   
   CREATE INDEX idx_new_table_name ON new_table (name);
   ```

2. **Add to migrations** (if needed):
   ```sql
   -- In migrations/004_add_new_table.sql
   CREATE TABLE new_table (
       -- same as above
   );
   ```

3. **Create queries**:
   ```sql
   -- In internal/servicename/supporting/postgres/queries/new_table.sql
   -- name: CreateNewTableEntry :one
   INSERT INTO new_table (id, name) VALUES ($1, $2) RETURNING *;
   ```

4. **Generate code**:
   ```bash
   sqlc generate
   ```

### Modifying Existing Tables

1. **Create migration**:
   ```sql
   -- In migrations/005_modify_table.sql
   ALTER TABLE users ADD COLUMN phone VARCHAR(20);
   CREATE INDEX idx_users_phone ON users (phone);
   ```

2. **Update schema file** to match
3. **Update queries** if needed
4. **Regenerate**: `sqlc generate`

## Repository Pattern

Each service follows this pattern:

```go
// Domain interface
type UserRepository interface {
    Create(ctx context.Context, user User) error
    GetByID(ctx context.Context, id string) (User, error)
    Update(ctx context.Context, user User) error
    Delete(ctx context.Context, id string) error
}

// PostgreSQL implementation
type userRepository struct {
    queries *Queries
}

func NewUserRepository(db *sql.DB) UserRepository {
    return &userRepository{
        queries: New(db),
    }
}
```

## Transaction Handling

For operations that need transactions:

```go
func (r *userRepository) CreateUserWithProfile(ctx context.Context, user User, profile Profile) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()
    
    qtx := r.queries.WithTx(tx)
    
    if _, err := qtx.CreateUser(ctx, CreateUserParams{...}); err != nil {
        return fmt.Errorf("failed to create user: %w", err)
    }
    
    if _, err := qtx.CreateProfile(ctx, CreateProfileParams{...}); err != nil {
        return fmt.Errorf("failed to create profile: %w", err)
    }
    
    return tx.Commit()
}
```

## Common Patterns

### Handling Optional Fields

```go
// For nullable database fields
type User struct {
    ID    string
    Name  string
    Email sql.NullString  // For optional fields
}

// In queries
func (r *userRepository) Create(ctx context.Context, user User) error {
    _, err := r.queries.CreateUser(ctx, CreateUserParams{
        ID:    user.ID,
        Name:  user.Name,
        Email: sql.NullString{String: user.Email, Valid: user.Email != ""},
    })
    return err
}
```

### JSON Fields

```sql
-- For storing JSON data
CREATE TABLE integrations (
    id UUID PRIMARY KEY,
    metadata JSONB
);
```

```go
// In Go code
type Integration struct {
    ID       string
    Metadata map[string]any
}
```

### Foreign Keys and Joins

```sql
-- name: GetUserWithOrganization :one
SELECT 
    u.id, u.name, u.email,
    o.id as org_id, o.name as org_name
FROM users u
JOIN organizations o ON u.organization_id = o.id
WHERE u.id = $1;
```

## Testing with Database

### Integration Tests

```go
func TestUserRepository_Integration(t *testing.T) {
    // Uses testcontainers to spin up real PostgreSQL
    db := postgrestest.NewTestDB(t)
    repo := NewUserRepository(db)
    
    user := User{
        ID:    "test-id",
        Name:  "Test User",
        Email: "test@example.com",
    }
    
    err := repo.Create(context.Background(), user)
    assert.NoError(t, err)
    
    retrieved, err := repo.GetByID(context.Background(), "test-id")
    assert.NoError(t, err)
    assert.Equal(t, user.Name, retrieved.Name)
}
```

## Common Commands

```bash
# Generate Go code from SQL
sqlc generate

# Check SQL syntax
sqlc vet

# Run migrations (handled automatically on startup)
# Migrations are in /migrations/ directory

# Connect to local database
psql -d infragpt -U your_username

# Reset database for testing
dropdb infragpt && createdb infragpt
```

## Best Practices

1. **Always use parameters** (`$1`, `$2`) - never string concatenation
2. **Add indexes** for all query patterns
3. **Use transactions** for multi-table operations  
4. **Handle errors** with proper context wrapping
5. **Never edit generated files** - they'll be overwritten
6. **Test with real database** using testcontainers
7. **Use meaningful names** for queries and parameters