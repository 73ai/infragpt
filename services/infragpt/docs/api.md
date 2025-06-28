# API Guide

How HTTP and gRPC APIs work in InfraGPT.

## Overview

InfraGPT provides both **HTTP REST** and **gRPC** APIs:

- **HTTP REST**: For web clients and external integrations
- **gRPC**: For communication with AI agents

## HTTP API Structure

### Endpoint Organization

```
http://localhost:8080/
├── /identity/          # User and organization management
├── /integrations/      # External service connections  
└── /*                  # Conversation service (everything else)
```

### API Handler Pattern

All HTTP endpoints use a consistent pattern:

```go
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
        // Your business logic here
        user, err := h.service.CreateUser(ctx, CreateUserCommand{
            Name:  req.Name,
            Email: req.Email,
        })
        if err != nil {
            return response{}, err
        }
        return response{ID: user.ID, Name: user.Name}, nil
    })
}
```

**What this gives you**:
- Automatic JSON marshaling/unmarshaling
- Type-safe request/response handling
- Centralized error handling
- Consistent HTTP status codes

## Identity Service API

Handles users and organizations.

### Endpoints

```bash
# User management
POST /identity/users/create
POST /identity/users/get
POST /identity/users/list

# Organization management  
POST /identity/organizations/create
POST /identity/organizations/get
POST /identity/organizations/members/list

# Webhook (internal)
POST /identity/webhooks/clerk
```

### Example Usage

```bash
# Create a user
curl -X POST http://localhost:8080/identity/users/create \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "clerk_id": "user_123"
  }'

# Response
{
  "id": "uuid-here",
  "name": "John Doe",
  "email": "john@example.com"
}
```

## Integration Service API

Manages external service connections.

### Endpoints

```bash
# Integration lifecycle
POST /integrations/initiate/     # Start OAuth flow
POST /integrations/authorize/    # Complete OAuth flow
POST /integrations/list/         # List integrations
POST /integrations/revoke/       # Remove integration
POST /integrations/status/       # Health check

# Webhook handlers (per connector)
POST /integrations/webhooks/github
POST /integrations/webhooks/slack
```

### OAuth Flow Example

```bash
# 1. Initiate authorization
curl -X POST http://localhost:8080/integrations/initiate/ \
  -H "Content-Type: application/json" \
  -d '{
    "organization_id": "org-123",
    "connector_type": "github"
  }'

# Response
{
  "type": "redirect",
  "url": "https://github.com/login/oauth/authorize?..."
}

# 2. User completes OAuth in browser, then:
curl -X POST http://localhost:8080/integrations/authorize/ \
  -H "Content-Type: application/json" \
  -d '{
    "organization_id": "org-123",
    "connector_type": "github",
    "code": "oauth-code-from-callback"
  }'

# Response
{
  "id": "integration-uuid",
  "connector_type": "github",
  "status": "active"
}
```

## Conversation Service API  

Handles Slack bot functionality and AI agent communication.

### Endpoints

```bash
# Message processing (primarily internal)
POST /process
POST /conversation/status

# Health check
GET /health
```

## gRPC API

Used for AI agent communication.

### Service Definition

```protobuf
// In infragptapi/proto/infragpt.proto
service InfraGPT {
    rpc ProcessMessage(ProcessMessageRequest) returns (ProcessMessageResponse);
}

message ProcessMessageRequest {
    string conversation_id = 1;
    string message = 2;
    string user_id = 3;
    map<string, string> context = 4;
}

message ProcessMessageResponse {
    string response = 1;
    string conversation_id = 2;
}
```

### Usage

```go
// Client example
conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
client := NewInfraGPTClient(conn)

response, err := client.ProcessMessage(ctx, &ProcessMessageRequest{
    ConversationId: "conv-123",
    Message:        "Help me deploy to production",
    UserId:         "user-456",
})
```

## Authentication

### Clerk JWT Middleware

Protected endpoints use Clerk JWT tokens:

```go
// Middleware is applied to specific endpoints
authMiddleware := c.Identity.Clerk.NewAuthMiddleware()

// Usage in handlers
type handler struct {
    service Service
    auth    AuthMiddleware
}

func (h *handler) protectedEndpoint() func(w http.ResponseWriter, r *http.Request) {
    return h.auth.RequireAuth(h.actualHandler())
}
```

### Getting User Context

In protected endpoints, get the authenticated user:

```go
func (h *handler) someEndpoint() func(w http.ResponseWriter, r *http.Request) {
    return ApiHandlerFunc(func(ctx context.Context, req request) (response, error) {
        userID, err := GetUserIDFromContext(ctx)
        if err != nil {
            return response{}, fmt.Errorf("unauthorized: %w", err)
        }
        
        // Use userID in your logic
        return response{}, nil
    })
}
```

## Error Handling

### HTTP Error Responses

The `ApiHandlerFunc` automatically converts Go errors to HTTP responses:

```go
// In your handler
if user.Email == "" {
    return response{}, &httperrors.BadRequest{Message: "email is required"}
}

if !found {
    return response{}, &httperrors.NotFound{Message: "user not found"}
}

// Generic errors become 500 Internal Server Error
if err != nil {
    return response{}, fmt.Errorf("database error: %w", err)
}
```

### Available Error Types

```go
// In internal/generic/httperrors/
&httperrors.BadRequest{Message: "invalid input"}     // 400
&httperrors.Unauthorized{Message: "not logged in"}   // 401
&httperrors.Forbidden{Message: "access denied"}      // 403
&httperrors.NotFound{Message: "resource not found"}  // 404
&httperrors.Conflict{Message: "already exists"}      // 409
```

## Testing APIs

### HTTP API Tests

```go
func TestCreateUser(t *testing.T) {
    // Set up test server
    handler := NewHandler(mockService, mockAuth)
    server := httptest.NewServer(handler)
    defer server.Close()
    
    // Make request
    resp, err := http.Post(server.URL+"/identity/users/create", 
        "application/json",
        strings.NewReader(`{"name":"John","email":"john@example.com"}`))
    
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    // Parse response
    var result CreateUserResponse
    json.NewDecoder(resp.Body).Decode(&result)
    assert.Equal(t, "John", result.Name)
}
```

### gRPC API Tests

```go
func TestProcessMessage(t *testing.T) {
    // Set up test gRPC server
    listener, err := net.Listen("tcp", ":0")
    require.NoError(t, err)
    
    server := NewGRPCServer(mockService)
    go server.Serve(listener)
    defer server.Stop()
    
    // Create client
    conn, err := grpc.Dial(listener.Addr().String(), grpc.WithInsecure())
    require.NoError(t, err)
    defer conn.Close()
    
    client := NewInfraGPTClient(conn)
    
    // Test the call
    resp, err := client.ProcessMessage(context.Background(), &ProcessMessageRequest{
        ConversationId: "test-conv",
        Message:        "hello",
    })
    
    assert.NoError(t, err)
    assert.NotEmpty(t, resp.Response)
}
```

## Adding New Endpoints

### 1. Add to Handler

```go
// In appropriate handler file (identityapi/, integrationapi/, infragptapi/)
func (h *handler) newEndpoint() func(w http.ResponseWriter, r *http.Request) {
    type request struct {
        Field string `json:"field"`
    }
    type response struct {
        Result string `json:"result"`
    }
    return ApiHandlerFunc(func(ctx context.Context, req request) (response, error) {
        // Call your service
        result, err := h.service.DoSomething(ctx, req.Field)
        if err != nil {
            return response{}, err
        }
        return response{Result: result}, nil
    })
}
```

### 2. Register Route

```go
// In the handler's constructor
func NewHandler(service Service, auth AuthMiddleware) http.Handler {
    mux := http.NewServeMux()
    h := &handler{service: service, auth: auth}
    
    mux.HandleFunc("POST /your/endpoint", h.newEndpoint())
    
    return mux
}
```

### 3. Add Service Method

```go
// In your service implementation
func (s *service) DoSomething(ctx context.Context, input string) (string, error) {
    // Your business logic here
    return "result", nil
}
```

### 4. Test It

```bash
curl -X POST http://localhost:8080/your/endpoint \
  -H "Content-Type: application/json" \
  -d '{"field": "value"}'
```

## Best Practices

1. **Use ApiHandlerFunc** for all HTTP endpoints
2. **Local request/response types** - don't share across handlers  
3. **JSON tags only on API boundaries** - keep internal models clean
4. **Proper error types** - use httperrors package
5. **Authentication** - apply middleware where needed
6. **Context passing** - always pass context through the chain
7. **Validation** - validate input in handlers before calling service
8. **Testing** - test both success and error cases