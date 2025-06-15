# Integration Manager UI Implementation Plan

## Overview
This document outlines the detailed implementation plan for the Integration Manager UI in the InfraGPT web application. The integration manager will provide a card-based interface for users to connect, configure, and manage external service integrations.

## Architecture Design

### Page Structure
- **Main List Page** (`/integrations`) - Grid of connector cards
- **Integration Details Page** (`/integrations/:connectorType`) - Configuration and status
- **OAuth Callback Pages** (`/integrations/:connectorType/callback`) - Handle authorization flows

### Component Hierarchy
```
IntegrationsPage
├── ConnectorGrid
│   └── ConnectorCard (repeated for each connector)
├── IntegrationDetailsPage
│   ├── IntegrationStatus
│   ├── IntegrationConfiguration
│   └── IntegrationActions
└── IntegrationCallbackPage
    ├── SuccessState
    └── ErrorState
```

## Implementation Steps

### Phase 1: Foundation Setup
**Estimated Time: 1-2 days**

#### 1.1 Create Base File Structure
```
src/
├── pages/integrations/
│   ├── IntegrationsPage.tsx
│   ├── IntegrationDetailsPage.tsx
│   ├── IntegrationCallbackPage.tsx
│   └── components/
│       ├── ConnectorCard.tsx
│       ├── ConnectorGrid.tsx
│       ├── IntegrationStatus.tsx
│       ├── IntegrationConfiguration.tsx
│       └── IntegrationActions.tsx
├── stores/
│   └── IntegrationStore.ts
├── services/
│   └── integrationService.ts
├── types/
│   └── integration.ts
└── assets/icons/
    ├── slack.svg
    ├── github.svg
    ├── aws.svg
    ├── gcp.svg
    ├── pagerduty.svg
    └── datadog.svg
```

#### 1.2 Define TypeScript Interfaces
Create comprehensive type definitions for:
- `Connector` interface with type, name, description, status
- `Integration` interface with connection details
- `AuthorizationIntent` for OAuth flows
- API request/response types

#### 1.3 Setup Routing
Add new routes to existing React Router configuration:
```typescript
{
  path: "/integrations",
  element: <IntegrationsPage />
},
{
  path: "/integrations/:connectorType",
  element: <IntegrationDetailsPage />
},
{
  path: "/integrations/:connectorType/callback",
  element: <IntegrationCallbackPage />
}
```

### Phase 2: Core Components Development
**Estimated Time: 2-3 days**

#### 2.1 ConnectorCard Component
**Features:**
- Display connector logo, name, and description
- Three button states: "Connect", "Show Details", "Coming Soon"
- Responsive design with hover effects
- Status indicators (connected/available/coming soon)

**Props Interface:**
```typescript
interface ConnectorCardProps {
  connector: Connector;
  integration?: Integration;
  onConnect: (connectorType: string) => void;
  onShowDetails: (connectorType: string) => void;
}
```

#### 2.2 ConnectorGrid Component
**Features:**
- Responsive grid layout (2-4 columns based on screen size)
- Filter/search functionality (future enhancement)
- Loading states and empty states

#### 2.3 IntegrationsPage Component
**Features:**
- Page header with title and description
- ConnectorGrid integration
- Navigation breadcrumbs
- Error boundary handling

### Phase 3: Integration Store (MobX)
**Estimated Time: 1-2 days**

#### 3.1 Store Implementation
**Core Methods:**
```typescript
class IntegrationStore {
  // Observable state
  integrations: Map<string, Integration>
  connectors: Connector[]
  loading: boolean
  error: string | null
  
  // Actions
  async loadIntegrations(organizationId: string)
  async initiateConnection(connectorType: string)
  async revokeIntegration(integrationId: string)
  
  // Computed values
  get connectedIntegrations()
  get availableConnectors()
  getIntegrationStatus(connectorType: string)
}
```

#### 3.2 Store Integration
- Connect store to React components using MobX observers
- Implement error handling and loading states
- Add automatic refresh mechanisms

### Phase 4: API Service Layer
**Estimated Time: 1-2 days**

#### 4.1 Integration Service Implementation
**API Endpoints to Implement:**
```typescript
class IntegrationService {
  // List integrations for organization
  async getIntegrations(organizationId: string): Promise<Integration[]>
  
  // Initiate OAuth/installation flow
  async initiateAuthorization(request: AuthorizeRequest): Promise<AuthorizationIntent>
  
  // Handle OAuth callback
  async handleCallback(connectorType: string, callbackData: any): Promise<Integration>
  
  // Get detailed integration status
  async getIntegrationDetails(connectorType: string, organizationId: string): Promise<Integration>
  
  // Test connection
  async testConnection(integrationId: string): Promise<boolean>
  
  // Revoke integration
  async revokeIntegration(integrationId: string): Promise<void>
}
```

#### 4.2 Error Handling
- Implement retry logic for failed requests
- Handle authentication errors (redirect to login)
- Provide user-friendly error messages

### Phase 5: Integration Details Page
**Estimated Time: 2-3 days**

#### 5.1 IntegrationStatus Component
**Features:**
- Connection status indicator (connected/error/syncing)
- Last sync timestamp
- Service-specific information (workspace name, connected channels, etc.)
- Real-time status updates

#### 5.2 IntegrationConfiguration Component
**Features:**
- Display current configuration settings
- Service-specific configuration options
- Permission levels and scopes
- Configuration editing (future enhancement)

#### 5.3 IntegrationActions Component
**Features:**
- Test Connection button with loading states
- Reconfigure button (reinitiate OAuth)
- Disconnect button with confirmation modal
- Settings button for advanced options

#### 5.4 Activity Log Component
**Features:**
- Recent activity timeline
- Event filtering and pagination
- Error event highlighting

### Phase 6: OAuth Callback Handling
**Estimated Time: 1-2 days**

#### 6.1 Callback Page Implementation
**Features:**
- Handle different callback scenarios (success/error)
- Extract authorization data from URL parameters
- Redirect users appropriately after processing
- Loading states during callback processing

#### 6.2 Connector-Specific Flows
**Slack OAuth2:**
- Handle `code` and `state` parameters
- Exchange code for access token
- Store integration credentials

**GitHub App Installation:**
- Handle `installation_id` and `setup_action` parameters
- Complete app installation process
- Configure webhook endpoints

### Phase 7: Styling and Polish
**Estimated Time: 1-2 days**

#### 7.1 Visual Design
- Apply consistent styling using existing Tailwind/Radix UI components
- Create connector-specific color schemes
- Implement responsive design patterns
- Add loading skeletons and animations

#### 7.2 Accessibility
- Ensure keyboard navigation support
- Add proper ARIA labels
- Implement screen reader compatibility
- Test with accessibility tools

### Phase 8: Integration and Testing
**Estimated Time: 1-2 days**

#### 8.1 Component Integration
- Wire all components together
- Test navigation flows
- Verify state management
- Test error scenarios

#### 8.2 API Integration Testing
- Test with actual backend endpoints
- Verify OAuth flows with real services
- Test error handling and edge cases
- Performance testing with multiple integrations

## Technical Specifications

### Connector Definitions
```typescript
const CONNECTORS: Connector[] = [
  {
    type: 'slack',
    name: 'Slack',
    description: 'Real-time messaging and notifications',
    logo: '/icons/slack.svg',
    isImplemented: true,
    authType: 'oauth2',
    capabilities: ['messaging', 'notifications', 'channels']
  },
  {
    type: 'github',
    name: 'GitHub',
    description: 'Code repository and PR management',
    logo: '/icons/github.svg',
    isImplemented: true,
    authType: 'app_installation',
    capabilities: ['repositories', 'webhooks', 'pull_requests']
  },
  {
    type: 'aws',
    name: 'AWS',
    description: 'Cloud infrastructure management',
    logo: '/icons/aws.svg',
    isImplemented: false,
    authType: 'api_key',
    capabilities: ['compute', 'storage', 'networking']
  },
  // ... additional connectors
];
```

### State Management Pattern
```typescript
// Use MobX for reactive state management
const integrationStore = new IntegrationStore();

// Component usage
const IntegrationsPage = observer(() => {
  const { integrationStore } = useStores();
  
  useEffect(() => {
    integrationStore.loadIntegrations(organizationId);
  }, []);
  
  return (
    <ConnectorGrid 
      connectors={integrationStore.connectors}
      integrations={integrationStore.integrations}
      onConnect={integrationStore.initiateConnection}
    />
  );
});
```

### API Integration Pattern
```typescript
// Service layer with proper error handling
class IntegrationService {
  private async request<T>(endpoint: string, options: RequestInit): Promise<T> {
    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      ...options,
      headers: {
        'Authorization': `Bearer ${getAuthToken()}`,
        'Content-Type': 'application/json',
        ...options.headers,
      },
    });
    
    if (!response.ok) {
      throw new IntegrationError(response.status, await response.text());
    }
    
    return response.json();
  }
}
```

## Security Considerations

### Authentication
- Use Clerk authentication for all API calls
- Store JWT tokens securely
- Handle token refresh automatically
- Redirect to login on authentication failures

### Data Handling
- Never store sensitive credentials in frontend
- Use secure callback URLs for OAuth flows
- Implement CSRF protection for state parameters
- Validate all callback data on backend

### Error Handling
- Don't expose sensitive error details to users
- Log security-relevant events
- Implement rate limiting awareness
- Handle malicious callback attempts gracefully

## Performance Considerations

### Loading Optimization
- Lazy load integration details
- Implement connector card virtualization for large lists
- Cache integration status data with TTL
- Use skeleton loading states

### Real-time Updates
- Implement WebSocket connection for status updates
- Use optimistic updates for user actions
- Debounce frequent status checks
- Handle connection failures gracefully

## Future Enhancements

### Phase 9: Advanced Features (Future)
- Bulk integration management
- Integration health monitoring dashboard
- Custom connector configuration templates
- Integration usage analytics
- Automated integration testing
- Integration marketplace

### Phase 10: Mobile Optimization (Future)
- Mobile-responsive card layouts
- Touch-friendly interaction patterns
- Offline status indicators
- Mobile OAuth flow optimization

## Success Metrics

### User Experience
- Time to complete integration setup < 2 minutes
- Integration success rate > 95%
- User satisfaction score > 4.5/5
- Zero critical accessibility issues

### Technical Performance
- Page load time < 2 seconds
- API response time < 500ms
- Error rate < 1%
- 99.9% uptime for integration services

## Dependencies

### External Libraries
- React Router for navigation
- MobX for state management
- Tailwind CSS for styling
- Radix UI for component primitives
- React Hook Form for form handling (if needed)

### Backend Dependencies
- InfraGPT Core Service API endpoints
- Clerk authentication service
- Integration Service webhooks
- External service OAuth endpoints

## Conclusion

This implementation plan provides a comprehensive roadmap for building a robust integration manager UI. The phased approach ensures steady progress while maintaining code quality and user experience standards. Each phase builds upon the previous one, allowing for iterative testing and refinement.

The estimated total development time is 10-15 days, depending on complexity of OAuth flows and styling requirements. The modular architecture ensures easy maintenance and future extensibility as new connectors are added to the platform.