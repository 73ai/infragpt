# Development Guide

Common development workflows and patterns.

## Daily Development Commands

```bash
# Start the server
go run ./cmd/main.go

# Run tests
go test ./...                      # All tests
go test ./internal/identitysvc/... # Specific service

# Format code before committing
find . -name "*.go" -not -path "./vendor/*" -exec goimports -w {} \;

# Generate database code after SQL changes
sqlc generate

# Check for issues
go vet ./...
```

## Making Changes

### Adding a New API Endpoint

1. **Add to the appropriate handler**:
   ```go
   // In identityapi/handler.go, integrationapi/handler.go, or infragptapi/handler.go
   func (h *handler) newEndpoint() func(w http.ResponseWriter, r *http.Request) {
       type request struct {
           Name string `json:"name"`
       }
       type response struct {
           ID string `json:"id"`
       }
       return ApiHandlerFunc(func(ctx context.Context, req request) (response, error) {
           // Your logic here
           return response{ID: "123"}, nil
       })
   }
   ```

2. **Wire it up in the handler's router**:
   ```go
   mux.HandleFunc("POST /your/endpoint", h.newEndpoint())
   ```

### Adding Database Operations

1. **Write SQL in the queries file**:
   ```sql
   -- In internal/servicename/supporting/postgres/queries/your_table.sql
   
   -- name: CreateUser :one
   INSERT INTO users (id, name, email)
   VALUES ($1, $2, $3)
   RETURNING *;
   
   -- name: GetUserByID :one
   SELECT * FROM users WHERE id = $1;
   ```

2. **Generate Go code**:
   ```bash
   sqlc generate
   ```

3. **Use in your repository**:
   ```go
   func (r *userRepository) Create(ctx context.Context, user User) error {
       _, err := r.queries.CreateUser(ctx, CreateUserParams{
           ID:    user.ID,
           Name:  user.Name,
           Email: user.Email,
       })
       return err
   }
   ```

### Adding a New Service Method

1. **Define in service interface** (if it doesn't exist):
   ```go
   // In internal/servicename/domain/interfaces.go
   type UserService interface {
       CreateUser(ctx context.Context, cmd CreateUserCommand) (User, error)
   }
   ```

2. **Implement in service**:
   ```go
   // In internal/servicename/service.go
   func (s *service) CreateUser(ctx context.Context, cmd CreateUserCommand) (User, error) {
       user := User{
           ID:    generateID(),
           Name:  cmd.Name,
           Email: cmd.Email,
       }
       
       if err := s.userRepository.Create(ctx, user); err != nil {
           return User{}, fmt.Errorf("failed to create user: %w", err)
       }
       
       return user, nil
   }
   ```

## Code Patterns to Follow

### Error Handling
```go
// ✅ Always wrap errors with context
if err != nil {
    return fmt.Errorf("failed to process user %s: %w", userID, err)
}

// ✅ Use structured logging
slog.Error("user creation failed", 
    "user_id", userID, 
    "email", email, 
    "error", err)
```

### API Handlers
```go
// ✅ Use the ApiHandlerFunc pattern
func (h *handler) createUser() func(w http.ResponseWriter, r *http.Request) {
    type request struct {
        Name  string `json:"name"`
        Email string `json:"email"`
    }
    type response struct {
        ID   string `json:"id"`
        Name string `json:"name"`
    }
    return ApiHandlerFunc(func(ctx context.Context, req request) (response, error) {
        cmd := CreateUserCommand{
            Name:  req.Name,
            Email: req.Email,
        }
        user, err := h.service.CreateUser(ctx, cmd)
        if err != nil {
            return response{}, err
        }
        return response{ID: user.ID, Name: user.Name}, nil
    })
}
```

### Configuration
```go
// ✅ Use mapstructure tags for YAML parsing
type Config struct {
    APIKey    string `mapstructure:"api_key"`
    Timeout   int    `mapstructure:"timeout"`
    EnableSSL bool   `mapstructure:"enable_ssl"`
}

// ✅ Factory method for service creation
func (c Config) New() (*Service, error) {
    return &Service{
        client: &http.Client{Timeout: time.Duration(c.Timeout) * time.Second},
    }, nil
}
```

## Testing Patterns

### Unit Tests
```go
func TestUserService_CreateUser(t *testing.T) {
    mockRepo := usertest.NewMockRepository()
    service := &userService{userRepository: mockRepo}
    
    user, err := service.CreateUser(context.Background(), CreateUserCommand{
        Name:  "John Doe",
        Email: "john@example.com",
    })
    
    assert.NoError(t, err)
    assert.Equal(t, "John Doe", user.Name)
}
```

### Integration Tests
```go
func TestUserRepository_Integration(t *testing.T) {
    db := postgrestest.NewTestDB(t) // Uses testcontainers
    repo := NewUserRepository(db)
    
    user := User{ID: "123", Name: "John", Email: "john@example.com"}
    err := repo.Create(context.Background(), user)
    
    assert.NoError(t, err)
    
    retrieved, err := repo.GetByID(context.Background(), "123")
    assert.NoError(t, err)
    assert.Equal(t, user.Name, retrieved.Name)
}
```

## Working with Integrations

### Adding a New Connector

1. **Create connector package**:
   ```
   internal/integrationsvc/connectors/newservice/
   ├── config.go     # Configuration
   ├── connector.go  # Main implementation
   ├── events.go     # Event types
   └── webhook.go    # HTTP handlers (if needed)
   ```

2. **Implement Connector interface**:
   ```go
   type Connector interface {
       InitiateAuthorization(organizationID, userID string) (IntegrationAuthorizationIntent, error)
       CompleteAuthorization(authData AuthorizationData) (Credentials, error)
       Subscribe(ctx context.Context, handler func(ctx context.Context, event any) error) error
       // ... other methods
   }
   ```

3. **Register in integration service**:
   ```go
   // In internal/integrationsvc/config.go
   if c.NewService.APIKey != "" {
       connectors[ConnectorTypeNewService] = c.NewService.NewConnector()
   }
   ```

## Common Gotchas

**SQLC**: Never edit generated `.sql.go` files - they'll be overwritten

**JSON Tags**: Only use on API boundary structs, not internal domain models

**Error Wrapping**: Always use `fmt.Errorf` with `%w` to preserve error chains

**Context**: Always pass context through the call chain, don't create new ones

**Database Transactions**: Use the existing patterns in repository implementations

## Git Workflow

1. **Create feature branch**: `git checkout -b feature/description`
2. **Make changes and commit frequently**
3. **Before pushing, format code**: `goimports -w .`
4. **Run tests**: `go test ./...`
5. **Push and create PR**