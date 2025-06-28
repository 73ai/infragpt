# Testing Guide

Testing patterns and utilities for InfraGPT.

## Overview

InfraGPT uses a comprehensive testing approach:

- **Unit tests** for business logic
- **Integration tests** with real PostgreSQL (via testcontainers)
- **API tests** for HTTP endpoints
- **Mock utilities** for external dependencies

## Test Organization

```
internal/servicename/
├── service.go
├── service_test.go           # Service unit tests
├── domain/
│   ├── models.go
│   └── models_test.go        # Domain model tests
├── supporting/
│   └── postgres/
│       ├── repository.go
│       └── repository_test.go # Repository integration tests
└── servicetest/              # Test utilities and mocks
    └── mocks.go
```

## Unit Tests

### Service Tests

```go
// internal/identitysvc/service_test.go
func TestUserService_CreateUser(t *testing.T) {
    // Setup mocks
    mockUserRepo := identitytest.NewMockUserRepository()
    mockOrgRepo := identitytest.NewMockOrganizationRepository()
    
    service := &userService{
        userRepository: mockUserRepo,
        orgRepository:  mockOrgRepo,
    }
    
    // Test data
    cmd := CreateUserCommand{
        Name:  "John Doe",
        Email: "john@example.com",
        OrgID: "org-123",
    }
    
    // Execute
    user, err := service.CreateUser(context.Background(), cmd)
    
    // Verify
    assert.NoError(t, err)
    assert.Equal(t, "John Doe", user.Name)
    assert.Equal(t, "john@example.com", user.Email)
    
    // Verify mock interactions
    assert.Equal(t, 1, mockUserRepo.CreateCallCount)
}
```

### Domain Model Tests

```go
// internal/identitysvc/domain/user_test.go
func TestUser_Validate(t *testing.T) {
    tests := []struct {
        name    string
        user    User
        wantErr bool
    }{
        {
            name: "valid user",
            user: User{
                ID:    "123",
                Name:  "John Doe",
                Email: "john@example.com",
            },
            wantErr: false,
        },
        {
            name: "missing email",
            user: User{
                ID:   "123",
                Name: "John Doe",
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.user.Validate()
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## Integration Tests

### Repository Tests with Real Database

```go
// internal/identitysvc/supporting/postgres/user_repository_test.go
func TestUserRepository_Integration(t *testing.T) {
    // Create test database using testcontainers
    db := postgrestest.NewTestDB(t)
    repo := NewUserRepository(db)
    
    ctx := context.Background()
    
    // Test data
    user := User{
        ID:    "test-id-123",
        Name:  "Test User",
        Email: "test@example.com",
        OrgID: "org-456",
    }
    
    // Test Create
    err := repo.Create(ctx, user)
    assert.NoError(t, err)
    
    // Test GetByID
    retrieved, err := repo.GetByID(ctx, "test-id-123")
    assert.NoError(t, err)
    assert.Equal(t, user.Name, retrieved.Name)
    assert.Equal(t, user.Email, retrieved.Email)
    
    // Test Update
    user.Name = "Updated Name"
    err = repo.Update(ctx, user)
    assert.NoError(t, err)
    
    updated, err := repo.GetByID(ctx, "test-id-123")
    assert.NoError(t, err)
    assert.Equal(t, "Updated Name", updated.Name)
    
    // Test Delete
    err = repo.Delete(ctx, "test-id-123")
    assert.NoError(t, err)
    
    _, err = repo.GetByID(ctx, "test-id-123")
    assert.Error(t, err) // Should not exist
}
```

### Test Database Setup

```go
// identitytest/postgres.go
func NewTestDB(t *testing.T) *sql.DB {
    ctx := context.Background()
    
    // Start PostgreSQL container
    container, err := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:15"),
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("testuser"),
        postgres.WithPassword("testpass"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready to accept connections").
                WithOccurrence(2).WithStartupTimeout(5*time.Second)),
    )
    require.NoError(t, err)
    
    // Clean up container when test finishes
    t.Cleanup(func() {
        require.NoError(t, container.Terminate(ctx))
    })
    
    // Get connection string
    connStr, err := container.ConnectionString(ctx, "sslmode=disable")
    require.NoError(t, err)
    
    // Connect to database
    db, err := sql.Open("postgres", connStr)
    require.NoError(t, err)
    
    // Run migrations
    err = runMigrations(db)
    require.NoError(t, err)
    
    return db
}
```

## API Tests

### HTTP Handler Tests

```go
// identityapi/handler_test.go
func TestCreateUserHandler(t *testing.T) {
    // Setup mock service
    mockService := identitytest.NewMockService()
    mockAuth := identitytest.NewMockAuth()
    
    // Create test server
    handler := NewHandler(mockService, mockAuth)
    server := httptest.NewServer(handler)
    defer server.Close()
    
    // Test request
    requestBody := `{
        "name": "John Doe",
        "email": "john@example.com",
        "organization_id": "org-123"
    }`
    
    resp, err := http.Post(
        server.URL+"/identity/users/create",
        "application/json",
        strings.NewReader(requestBody),
    )
    require.NoError(t, err)
    defer resp.Body.Close()
    
    // Verify response
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    var result CreateUserResponse
    err = json.NewDecoder(resp.Body).Decode(&result)
    require.NoError(t, err)
    
    assert.Equal(t, "John Doe", result.Name)
    assert.Equal(t, "john@example.com", result.Email)
}
```

### gRPC Handler Tests

```go
// infragptapi/grpc_test.go
func TestGRPCProcessMessage(t *testing.T) {
    // Setup mock service
    mockService := conversationtest.NewMockService()
    
    // Create gRPC test server
    listener, err := net.Listen("tcp", "localhost:0")
    require.NoError(t, err)
    
    server := NewGRPCServer(mockService)
    go server.Serve(listener)
    defer server.Stop()
    
    // Create client
    conn, err := grpc.Dial(
        listener.Addr().String(),
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    require.NoError(t, err)
    defer conn.Close()
    
    client := NewInfraGPTClient(conn)
    
    // Test the call
    resp, err := client.ProcessMessage(context.Background(), &ProcessMessageRequest{
        ConversationId: "test-conv-123",
        Message:        "Hello, deploy to staging",
        UserId:         "user-456",
    })
    
    require.NoError(t, err)
    assert.Equal(t, "test-conv-123", resp.ConversationId)
    assert.NotEmpty(t, resp.Response)
}
```

## Mock Utilities

### Service Mocks

```go
// identitytest/mocks.go
type MockUserService struct {
    CreateUserFunc     func(ctx context.Context, cmd CreateUserCommand) (User, error)
    GetUserFunc        func(ctx context.Context, query GetUserQuery) (User, error)
    CreateUserCallCount int
    GetUserCallCount   int
}

func NewMockUserService() *MockUserService {
    return &MockUserService{
        CreateUserFunc: func(ctx context.Context, cmd CreateUserCommand) (User, error) {
            return User{
                ID:    "mock-id",
                Name:  cmd.Name,
                Email: cmd.Email,
            }, nil
        },
        GetUserFunc: func(ctx context.Context, query GetUserQuery) (User, error) {
            return User{
                ID:    query.ID,
                Name:  "Mock User",
                Email: "mock@example.com",
            }, nil
        },
    }
}

func (m *MockUserService) CreateUser(ctx context.Context, cmd CreateUserCommand) (User, error) {
    m.CreateUserCallCount++
    return m.CreateUserFunc(ctx, cmd)
}

func (m *MockUserService) GetUser(ctx context.Context, query GetUserQuery) (User, error) {
    m.GetUserCallCount++
    return m.GetUserFunc(ctx, query)
}
```

### Repository Mocks

```go
// identitytest/repository_mocks.go
type MockUserRepository struct {
    CreateFunc      func(ctx context.Context, user User) error
    GetByIDFunc     func(ctx context.Context, id string) (User, error)
    CreateCallCount int
    GetByIDCallCount int
    Users           map[string]User
}

func NewMockUserRepository() *MockUserRepository {
    return &MockUserRepository{
        Users: make(map[string]User),
        CreateFunc: func(ctx context.Context, user User) error {
            // Default implementation stores in memory
            return nil
        },
        GetByIDFunc: func(ctx context.Context, id string) (User, error) {
            // Default implementation retrieves from memory
            return User{}, fmt.Errorf("user not found")
        },
    }
}

func (m *MockUserRepository) Create(ctx context.Context, user User) error {
    m.CreateCallCount++
    m.Users[user.ID] = user
    return m.CreateFunc(ctx, user)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (User, error) {
    m.GetByIDCallCount++
    if user, exists := m.Users[id]; exists {
        return user, nil
    }
    return m.GetByIDFunc(ctx, id)
}
```

## Test Data Utilities

### Test Builders

```go
// identitytest/builders.go
type UserBuilder struct {
    user User
}

func NewUserBuilder() *UserBuilder {
    return &UserBuilder{
        user: User{
            ID:    "default-id",
            Name:  "Default User",
            Email: "default@example.com",
            OrgID: "default-org",
        },
    }
}

func (b *UserBuilder) WithID(id string) *UserBuilder {
    b.user.ID = id
    return b
}

func (b *UserBuilder) WithName(name string) *UserBuilder {
    b.user.Name = name
    return b
}

func (b *UserBuilder) WithEmail(email string) *UserBuilder {
    b.user.Email = email
    return b
}

func (b *UserBuilder) Build() User {
    return b.user
}

// Usage in tests:
func TestSomething(t *testing.T) {
    user := NewUserBuilder().
        WithName("John Doe").
        WithEmail("john@example.com").
        Build()
    
    // Use user in test...
}
```

## Running Tests

### Basic Commands

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Run specific service tests
go test ./internal/identitysvc/...

# Run only unit tests (exclude integration tests)
go test -short ./...

# Run integration tests
go test -tags=integration ./...
```

### Test Organization

```bash
# Service-specific tests
go test ./internal/identitysvc/
go test ./internal/integrationsvc/
go test ./internal/conversationsvc/

# API tests
go test ./identityapi/
go test ./integrationapi/
go test ./infragptapi/

# Repository tests (require database)
go test ./internal/*/supporting/postgres/
```

## Test Configuration

### Test-specific config

```go
// In test files
func TestMain(m *testing.M) {
    // Setup test environment
    os.Setenv("LOG_LEVEL", "error") // Reduce log noise
    
    // Run tests
    code := m.Run()
    
    // Cleanup
    os.Exit(code)
}
```

### Skip integration tests in CI

```go
func TestIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    // Integration test code...
}
```

## Best Practices

1. **Use table-driven tests** for multiple scenarios
2. **Test both success and error cases**
3. **Use real database for repository tests** (testcontainers)
4. **Mock external dependencies** (APIs, services)
5. **Test API boundaries** with HTTP/gRPC clients
6. **Verify mock interactions** to ensure correct behavior
7. **Use builders for complex test data**
8. **Clean up resources** in test cleanup functions
9. **Run tests in parallel** where possible
10. **Keep tests focused** - one thing per test

## Common Patterns

### Testing Errors

```go
func TestServiceWithError(t *testing.T) {
    mockRepo := identitytest.NewMockUserRepository()
    mockRepo.CreateFunc = func(ctx context.Context, user User) error {
        return fmt.Errorf("database error")
    }
    
    service := &userService{userRepository: mockRepo}
    
    _, err := service.CreateUser(context.Background(), CreateUserCommand{
        Name: "Test",
        Email: "test@example.com",
    })
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "database error")
}
```

### Testing Concurrent Operations

```go
func TestConcurrentAccess(t *testing.T) {
    db := postgrestest.NewTestDB(t)
    repo := NewUserRepository(db)
    
    var wg sync.WaitGroup
    errors := make(chan error, 10)
    
    // Run 10 concurrent operations
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            user := User{
                ID:    fmt.Sprintf("user-%d", id),
                Name:  fmt.Sprintf("User %d", id),
                Email: fmt.Sprintf("user%d@example.com", id),
            }
            
            if err := repo.Create(context.Background(), user); err != nil {
                errors <- err
            }
        }(i)
    }
    
    wg.Wait()
    close(errors)
    
    // Check for errors
    for err := range errors {
        t.Errorf("Concurrent operation failed: %v", err)
    }
}
```