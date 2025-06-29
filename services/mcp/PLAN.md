# MCP Tools Service Implementation Plan

## Overview

The MCP Tools Service is a Go-based service that acts as the external integration layer for the InfraGPT multi-tenant AI platform. It exposes standardized tools via the Model Context Protocol (MCP) that can be called by the Python agent service to interact with external APIs like GitHub, GCP, Slack, and Datadog.

## Architecture

```
┌─────────────────┐    MCP Protocol     ┌─────────────────┐
│  Agent Service  │◄──────────────────► │  MCP Tools      │
│   (Python)      │     (TCP/Stdio)     │   Service (Go)  │
└─────────────────┘                     └─────────────────┘
                                                 │
                                        gRPC     │
                                                 ▼
                                        ┌─────────────────┐
                                        │ Core Go Service │
                                        │ (Credentials)   │
                                        └─────────────────┘
```

### Key Components

1. **MCP Server**: Implements the Model Context Protocol using `github.com/mark3labs/mcp-go`
2. **Tool Providers**: Individual modules for each external service (GitHub, GCP, Slack, Datadog)
3. **Credential Manager**: gRPC client that fetches tenant-specific credentials from the core service
4. **Cache Layer**: Basic caching for expensive operations
5. **Configuration Management**: Environment-based configuration

## Project Structure

```
/services/mcp/
├── cmd/
│   └── main.go                 # Service entry point
├── internal/
│   ├── server/
│   │   ├── mcp_server.go      # MCP server implementation
│   │   └── handler.go         # Tool execution handlers
│   ├── tools/
│   │   ├── github/
│   │   │   ├── client.go      # GitHub API client
│   │   │   └── tools.go       # GitHub tool definitions
│   │   ├── gcp/
│   │   │   ├── client.go      # GCP API client
│   │   │   └── tools.go       # GCP tool definitions
│   │   ├── slack/
│   │   │   ├── client.go      # Slack API client
│   │   │   └── tools.go       # Slack tool definitions
│   │   └── datadog/
│   │       ├── client.go      # Datadog API client
│   │       └── tools.go       # Datadog tool definitions
│   ├── credentials/
│   │   ├── client.go          # gRPC client for core service
│   │   └── manager.go         # Credential management
│   ├── cache/
│   │   └── cache.go           # Basic caching implementation
│   └── config/
│       └── config.go          # Configuration management
├── proto/
│   ├── credentials.proto      # Credential service proto
│   └── generated/             # Generated protobuf files
├── go.mod
├── go.sum
├── README.md
└── PLAN.md                    # This file
```

## Implementation Phases

### Phase 1: Foundation & MCP Server
**Duration**: 2-3 days

#### Tasks:
1. **Project Setup**
   - Initialize Go module with `github.com/mark3labs/mcp-go`
   - Set up project structure
   - Configure logging and basic error handling

2. **MCP Server Implementation**
   - Implement basic MCP server with tool discovery
   - Set up TCP transport for MCP communication
   - Create tool registry system

3. **Configuration Management**
   - Environment variable configuration
   - Service discovery configuration
   - Logging configuration

4. **Credential Service gRPC Client**
   - Define protobuf schema for credential service
   - Implement gRPC client to communicate with core service
   - Add credential caching mechanism

#### Deliverables:
- Basic MCP server that can list tools
- gRPC client for credential fetching
- Configuration management system

### Phase 2: GitHub Integration
**Duration**: 3-4 days

#### GitHub Tools to Implement:
- `github.get_commits` - Get repository commits
- `github.get_issues` - List repository issues  
- `github.create_issue` - Create new issue
- `github.get_pull_requests` - List pull requests
- `github.get_repositories` - List user/org repositories

#### Tasks:
1. **GitHub API Client**
   - Implement GitHub REST API client
   - Handle authentication with tenant-specific tokens
   - Implement rate limiting per tenant

2. **Tool Definitions**
   - Define MCP tool schemas for each GitHub operation
   - Implement tool execution handlers
   - Add input validation and error handling

3. **Caching**
   - Implement basic caching for read operations
   - Configure TTL per operation type

#### Deliverables:
- Fully functional GitHub integration
- Comprehensive error handling
- Basic caching implementation

### Phase 3: GCP Integration
**Duration**: 4-5 days

#### GCP Tools to Implement:
- `gcp.deploy` - Trigger Cloud Build deployment
- `gcp.get_services` - List Cloud Run services
- `gcp.get_logs` - Get application logs
- `gcp.scale_service` - Scale Cloud Run service

#### Tasks:
1. **GCP API Client**
   - Implement Google Cloud API clients (Cloud Run, Cloud Build, Logging)
   - Handle service account authentication
   - Implement project-specific operations

2. **Tool Definitions**
   - Define MCP tool schemas for each GCP operation
   - Implement tool execution handlers
   - Handle long-running operations (synchronous for now)

3. **Error Handling**
   - GCP-specific error handling and retry logic
   - Proper error propagation via MCP

#### Deliverables:
- Full GCP integration suite
- Robust error handling for cloud operations
- Documentation for GCP tool usage

### Phase 4: Additional Integrations
**Duration**: 3-4 days

#### Slack Tools to Implement:
- `slack.send_message` - Send message to channel
- `slack.get_channels` - List channels
- `slack.get_users` - List workspace users
- `slack.create_channel` - Create new channel

#### Datadog Tools to Implement:
- `datadog.get_metrics` - Retrieve metrics
- `datadog.create_dashboard` - Create dashboard
- `datadog.get_logs` - Get application logs
- `datadog.create_alert` - Create alert rule

#### Tasks:
1. **Slack Integration**
   - Implement Slack API client
   - Handle workspace-specific tokens
   - Implement Slack tool definitions

2. **Datadog Integration**
   - Implement Datadog API client
   - Handle API key authentication
   - Implement monitoring tool definitions

#### Deliverables:
- Slack integration with messaging capabilities
- Datadog integration with monitoring tools
- Complete external service coverage

### Phase 5: Agent Service Integration
**Duration**: 2-3 days

#### Tasks:
1. **Python MCP Client**
   - Install and configure MCP client in agent service
   - Modify existing tools framework to route calls through MCP
   - Update tool registry to discover MCP tools

2. **Integration Testing**
   - End-to-end testing with Python agent service
   - Test multi-tenant credential isolation
   - Performance testing and optimization

3. **Documentation**
   - Complete API documentation
   - Integration guides
   - Troubleshooting documentation

#### Deliverables:
- Fully integrated MCP Tools Service with agent service
- Comprehensive documentation
- Production-ready deployment

## Technical Specifications

### MCP Tool Definition Pattern

```go
type Tool struct {
    Name        string                 // e.g., "github.get_commits"
    Description string                 // Tool description for LLM
    Provider    string                 // "github", "gcp", "slack", "datadog"
    Schema      map[string]interface{} // JSON schema for parameters
    Execute     func(ctx context.Context, tenantID string, params map[string]interface{}) (interface{}, error)
}
```

### Error Standardization

```go
type ToolError struct {
    Code     string                 // "RATE_LIMIT", "AUTH_FAILED", "NOT_FOUND", etc.
    Message  string                 // Human-readable error message
    Provider string                 // Which service caused the error
    Details  map[string]interface{} // Additional error context
}
```

### Credential Flow

1. MCP Tools Service starts with Core Service gRPC endpoint configured
2. On tool execution request, service extracts tenant_id from context
3. Service calls Core Service with tenant_id + provider to get credentials
4. Core Service returns tenant-specific credentials (API keys, tokens, etc.)
5. MCP Tools Service uses credentials to make external API call
6. Results are returned via MCP protocol to agent service

### Configuration

#### Environment Variables
- `MCP_LISTEN_ADDRESS` - MCP server listen address (default: ":8080")
- `CORE_SERVICE_GRPC_ENDPOINT` - Core service gRPC endpoint
- `LOG_LEVEL` - Logging level (debug, info, warn, error)
- `CACHE_TTL` - Cache TTL duration (default: "5m")
- `MAX_CONCURRENT_REQUESTS` - Maximum concurrent tool executions per tenant

#### Core Service gRPC Interface

```protobuf
syntax = "proto3";

package credentials;

service CredentialService {
  rpc GetCredentials(GetCredentialsRequest) returns (GetCredentialsResponse);
  rpc ListProviders(ListProvidersRequest) returns (ListProvidersResponse);
}

message GetCredentialsRequest {
  string tenant_id = 1;
  string provider = 2; // "github", "gcp", "slack", "datadog"
}

message GetCredentialsResponse {
  map<string, string> credentials = 1; // API keys, tokens, etc.
  bool success = 2;
  string error_message = 3;
}

message ListProvidersRequest {
  string tenant_id = 1;
}

message ListProvidersResponse {
  repeated string providers = 1;
}
```

## Security Considerations

### Multi-Tenant Isolation
- All credentials are fetched per-tenant from core service
- No credential storage in MCP service
- Tenant context passed with every tool execution
- Rate limiting applied per tenant per provider

### Credential Security
- Credentials encrypted in transit via gRPC TLS
- No credential logging or persistence
- Credential caching with short TTL and encryption at rest
- Secure credential cleanup after use

### API Security
- Input validation for all tool parameters
- Output sanitization to prevent data leakage
- Request size limits to prevent DoS
- Timeout enforcement for external API calls

## Testing Strategy

### Unit Testing
- Individual tool function testing with mocked external APIs
- Credential manager testing with mock gRPC service
- Cache testing with various TTL scenarios
- Error handling testing for all error conditions

### Integration Testing
- End-to-end testing with real external APIs (using test accounts)
- Multi-tenant testing to ensure credential isolation
- Performance testing under load
- Failure scenario testing (network issues, API rate limits)

### Load Testing
- Concurrent tool execution testing
- Memory usage testing with caching
- gRPC connection pooling testing
- Rate limit handling under load

## Monitoring and Observability

### Metrics
- Tool execution count by provider and tenant
- Tool execution duration by provider and operation
- Error rate by provider and error type
- Cache hit/miss ratio
- Active gRPC connections

### Logging
- Structured logging with tenant_id context
- Tool execution logs with duration and status
- Error logs with full context and stack traces
- Audit logs for credential access

### Health Checks
- MCP server health endpoint
- Core service gRPC connection health
- External API connectivity checks
- Cache system health

## Deployment Considerations

### Dependencies
- Core service must be running and accessible via gRPC
- External API credentials must be configured in core service
- Network connectivity to external APIs (GitHub, GCP, Slack, Datadog)

### Resource Requirements
- Memory: 512MB - 1GB (depending on cache size)
- CPU: 1-2 cores (depending on concurrent load)
- Network: Outbound access to external APIs
- Storage: Minimal (logs only, no persistent state)

### Scaling
- Stateless service design allows horizontal scaling
- Load balancing can be added if needed
- Cache can be externalized (Redis) for multi-instance deployments

## Future Enhancements

### Phase 6: Advanced Features
- Async operation support with job tracking
- Webhook integration for long-running operations
- Advanced caching with Redis
- Metrics and monitoring integration
- Tool usage analytics and optimization

### Phase 7: Additional Providers
- AWS integration (EC2, S3, Lambda, CloudWatch)
- Azure integration (VMs, Functions, Monitor)
- Kubernetes integration
- Jenkins/CI-CD integration
- Database integrations (PostgreSQL, MongoDB, Redis)

### Phase 8: Advanced Security
- OAuth2 flow support for better authentication
- Credential rotation and refresh
- Advanced rate limiting with backoff
- Request/response encryption
- Audit trail and compliance features

## Success Criteria

1. **Functional Requirements**
   - All defined tools working correctly with real external APIs
   - Multi-tenant credential isolation working properly
   - Integration with Python agent service complete
   - Error handling comprehensive and user-friendly

2. **Performance Requirements**
   - Tool execution under 5 seconds for most operations
   - Support for 100+ concurrent tool executions
   - Cache hit ratio above 80% for read operations
   - 99.9% uptime in production

3. **Security Requirements**
   - No credential leakage or cross-tenant access
   - All external API calls properly authenticated
   - Input validation preventing injection attacks
   - Secure logging without sensitive data exposure

4. **Operational Requirements**
   - Comprehensive monitoring and alerting
   - Easy deployment and configuration
   - Clear documentation and troubleshooting guides
   - Automated testing covering all critical paths