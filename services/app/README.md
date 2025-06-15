# InfraGPT Web Application

InfraGPT Web Application is a modern React TypeScript client that provides a comprehensive interface for AI-powered infrastructure management. The application features Clerk authentication, organization onboarding workflows, integration management with external services, and seamless integration with the InfraGPT backend services.

## Architecture Overview

The application follows modern React patterns with TypeScript, featuring:

- **Authentication**: Clerk integration for user and organization management
- **Routing**: React Router for client-side navigation with protected routes
- **UI Components**: Radix UI with Tailwind CSS for modern, accessible design
- **State Management**: MobX stores for application state management
- **Integration Management**: Comprehensive external service integration workflows
- **Form Handling**: React Hook Form with Zod validation
- **Build System**: Vite for fast development and optimized production builds

## Project Structure

```
src/
├── main.tsx                    # Application entry point
├── App.tsx                     # Root component with routing
├── index.css                   # Global styles and Tailwind imports
├── vite-env.d.ts              # Vite environment types
│
├── components/                 # Reusable UI components
│   ├── MainPage.tsx           # Main dashboard component
│   ├── app-sidebar.tsx        # Application sidebar navigation
│   ├── nav-main.tsx           # Main navigation component
│   ├── nav-user.tsx           # User navigation component
│   │
│   ├── onboarding/            # Organization onboarding flow
│   │   ├── OnboardingForm.tsx # Multi-step onboarding form
│   │   └── steps/             # Individual form steps
│   │       ├── CompanySizeStep.tsx
│   │       ├── UseCasesStep.tsx
│   │       ├── ObservabilityStackStep.tsx
│   │       └── SummaryStep.tsx
│   │
│   ├── data-table/            # Reusable data table components
│   │   ├── data-table.tsx
│   │   ├── columns.tsx
│   │   ├── data-table-pagination.tsx
│   │   └── data-table-view-options.tsx
│   │
│   └── ui/                    # shadcn/ui components
│       ├── avatar.tsx
│       ├── button.tsx
│       ├── card.tsx
│       ├── checkbox.tsx
│       ├── collapsible.tsx
│       ├── dropdown-menu.tsx
│       ├── form.tsx
│       ├── input.tsx
│       ├── label.tsx
│       ├── layout.tsx
│       ├── select.tsx
│       ├── separator.tsx
│       ├── sheet.tsx
│       ├── sidebar.tsx
│       ├── skeleton.tsx
│       ├── sonner.tsx
│       ├── table.tsx
│       └── tooltip.tsx
│
├── hooks/                     # Custom React hooks
│   ├── use-mobile.tsx         # Mobile device detection
│   └── useOnboardingGuard.tsx # Organization onboarding status
│
├── lib/                       # Utility functions and configurations
│   ├── utils.ts               # Utility functions (cn, etc.)
│   ├── api.ts                 # API client with Clerk authentication
│   ├── onboarding-constants.ts # Onboarding form constants and types
│   └── integration-constants.ts # Integration connector definitions and constants
│
├── stores/                    # MobX state management stores
│   ├── IntegrationStore.ts    # Integration management state and actions
│   └── UserStore.ts           # User profile and organization state
│
├── services/                  # API service layer
│   └── integrationService.ts  # Integration API client with security features
│
├── types/                     # TypeScript type definitions
│   └── integration.ts         # Integration-related type definitions
│
├── pages/                     # Page components
│   ├── login.tsx              # Login page
│   ├── signup.tsx             # Signup page
│   ├── onboarding.tsx         # Organization onboarding page
│   └── integrations/          # Integration management pages
│       ├── IntegrationsPage.tsx        # Integration listing and management
│       ├── IntegrationDetailsPage.tsx  # Individual integration details
│       ├── IntegrationCallbackPage.tsx # OAuth callback handling
│       └── components/        # Integration-specific components
│           ├── ConnectorCard.tsx       # Integration card display
│           ├── ConnectorGrid.tsx       # Integration grid layout
│           ├── IntegrationStatus.tsx   # Status and timeline display
│           ├── IntegrationConfiguration.tsx # Configuration details
│           └── IntegrationActions.tsx  # Action buttons and controls
│
└── assets/                    # Static assets
    ├── fonts/                 # Custom font files
    └── react.svg              # React logo
```

## Key Features

### 1. Authentication & Authorization
- **Clerk Integration**: Complete OAuth authentication flow
- **Organization Management**: Multi-tenant organization support
- **Protected Routes**: Route guards that enforce authentication
- **Session Management**: Automatic token refresh and validation

### 2. Organization Onboarding
- **Mandatory Onboarding**: Users must complete organization setup
- **Multi-Step Form**: Progressive data collection with validation
- **Form Persistence**: Automatic saving and prefilling of existing data
- **Progress Tracking**: Visual progress indicator across steps

#### Onboarding Flow:
1. **Step 1 - Company Info**: Company size and team size selection
2. **Step 2 - Use Cases**: Multi-select use case identification
3. **Step 3 - Observability Stack**: Current tooling selection
4. **Step 4 - Summary**: Review and submit collected information

### 3. Integration Management
- **External Service Connectors**: Support for Slack, GitHub, AWS, GCP, and other services
- **OAuth Authentication**: Secure OAuth2 flows for service connections
- **Real-time Status**: Live integration status monitoring with visual indicators
- **Configuration Management**: Detailed connector-specific configuration views
- **Timeline Tracking**: Creation and update timestamps with human-readable formatting
- **Security Features**: Protected callback handling and URI decoding safety

#### Supported Integrations:
- **Slack**: Workspace and channel connectivity with team ID tracking
- **GitHub**: Repository management with installation ID and webhook configuration
- **AWS**: Account and region-based integration with credential management
- **GCP**: Project-based integration with service account handling
- **Datadog, PagerDuty**: Additional monitoring and alerting service support

### 4. UI/UX Features
- **Modern Design**: Clean, professional interface with Tailwind CSS
- **Responsive Layout**: Mobile-first design with adaptive layouts
- **Accessibility**: ARIA-compliant components from Radix UI
- **Loading States**: Comprehensive loading and error state management
- **Toast Notifications**: User feedback with Sonner toast system

### 5. Data Management
- **Type Safety**: Full TypeScript integration with Zod validation
- **Form Validation**: Real-time validation with React Hook Form
- **API Integration**: Type-safe API client with error handling
- **State Management**: MobX stores for efficient state management and reactivity

## API Integration

### Backend Communication
The application communicates with the InfraGPT Go backend service:

```typescript
// API Client Configuration
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

// Integration Service with Security Features
class IntegrationService {
  async getIntegrations(organizationId: string): Promise<Integration[]>
  async initiateAuthorization(organizationId: string, userId: string, connectorType: ConnectorType): Promise<AuthorizeResponse>
  async handleCallback(connectorType: ConnectorType, callbackData: Record<string, any>): Promise<Integration>
  async getIntegrationDetails(organizationId: string, connectorType: ConnectorType): Promise<Integration>
  async testConnection(integrationId: string): Promise<TestConnectionResponse>
  async revokeIntegration(integrationId: string): Promise<void>
}
```

### Authentication Flow
1. User authenticates with Clerk
2. Application receives JWT token
3. API client includes token in Authorization header
4. Backend validates token with Clerk
5. Protected operations proceed with verified identity

## Onboarding Data Types

### Company Information
```typescript
type CompanySize = 'startup' | 'small' | 'medium' | 'large' | 'enterprise'
type TeamSize = '1-5' | '6-20' | '21-50' | '51-100' | '100+'
```

### Use Cases
```typescript
type UseCase = 
  | 'infrastructure_monitoring'
  | 'application_performance_monitoring' 
  | 'log_management'
  | 'incident_response'
  | 'compliance_auditing'
  | 'cost_optimization'
  | 'security_monitoring'
  | 'devops_automation'
```

### Observability Stack
```typescript
type ObservabilityStack = 
  | 'datadog' | 'new_relic' | 'splunk' | 'elastic_stack'
  | 'prometheus_grafana' | 'app_dynamics' | 'dynatrace'
  | 'cloudwatch' | 'azure_monitor' | 'google_cloud_monitoring'
  | 'pagerduty' | 'opsgenie' | 'other'
```

## Integration Management Data Types

### Integration Status
```typescript
type IntegrationStatus = 'connected' | 'error' | 'syncing' | 'disconnected'

interface Integration {
  id: string
  organizationId: string
  userId: string
  connectorType: ConnectorType
  status: IntegrationStatus
  configuration?: Record<string, any>
  metadata?: {
    errorMessage?: string
    errorCode?: string
    lastSync?: string
  }
  createdAt: string
  updatedAt: string
}
```

### Connector Types
```typescript
type ConnectorType = 'slack' | 'github' | 'aws' | 'gcp' | 'datadog' | 'pagerduty'

interface Connector {
  type: ConnectorType
  name: string
  description: string
  logo: string
  category: 'communication' | 'code' | 'cloud' | 'monitoring'
  authType: 'oauth2' | 'api_key' | 'service_account'
  isAvailable: boolean
}
```

### OAuth Response Types
```typescript
interface AuthorizeResponse {
  type: 'redirect' | 'popup' | 'oauth2'
  url: string
  state?: string
}

interface CallbackRequest {
  connector_type: ConnectorType
  code?: string
  state?: string
  [key: string]: any
}
```

## Component Architecture

### Protected Routes
```typescript
const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
  const { isSignedIn, isLoaded } = useAuth();
  const { isLoading, isComplete } = useOnboardingGuard();
  
  // Authentication and onboarding validation logic
  // Redirects to appropriate flow based on state
}
```

### Onboarding Guard Hook
```typescript
export const useOnboardingGuard = (): OnboardingStatus => {
  // Checks organization existence and metadata completion
  // Returns loading state and completion status
  // Integrates with Clerk organization context
}
```

### API Client Hook
```typescript
export const useApiClient = () => {
  const { getToken } = useAuth();
  
  // Provides authenticated API methods
  // Handles token refresh and error states
  // Returns typed API functions
}
```

### MobX State Stores
```typescript
// Integration Store - Manages integration state and operations
class IntegrationStore {
  integrations = new Map<string, Integration>()
  loading = false
  error: string | null = null
  
  async loadIntegrations(organizationId: string): Promise<void>
  async initiateConnection(connectorType: ConnectorType, organizationId: string, userId: string): Promise<AuthorizeResponse>
  async testConnection(integrationId: string): Promise<TestConnectionResponse>
  async revokeIntegration(integrationId: string): Promise<void>
  getIntegrationByConnectorType(connectorType: ConnectorType): Integration | undefined
}

// User Store - Manages user profile and organization data
class UserStore {
  userProfile: UserProfile | null = null
  loading = false
  organizationId: string | null = null
  userId: string | null = null
  
  async loadUserProfile(apiClient: any, clerkUserId: string, clerkOrgId: string): Promise<void>
}
```

## Development Setup

### Prerequisites
- Node.js 18+ 
- npm or yarn package manager
- Access to InfraGPT backend service
- Clerk application credentials

### Environment Configuration
Create a `.env.local` file:

```env
VITE_CLERK_PUBLISHABLE_KEY=pk_test_...
VITE_API_BASE_URL=http://localhost:8080
```

### Installation & Setup
```bash
# Clone repository
git clone https://github.com/priyanshujain/infragpt.git
cd infragpt/services/app

# Install dependencies
npm install

# Start development server
npm run dev
```

### Development Commands

#### Core Commands
```bash
npm run dev          # Start development server (localhost:5173)
npm run build        # Build for production (TypeScript + Vite)
npm run preview      # Preview production build locally
npm run lint         # Run ESLint with TypeScript support
```

#### Development Workflow
```bash
# Start backend service (in separate terminal)
cd ../infragpt
go run ./cmd/main.go

# Start frontend development
npm run dev

# Access application
open http://localhost:5173
```

## Build & Deployment

### Production Build
```bash
npm run build
```

Output directory: `dist/`

### Build Optimization
- **Vite**: Fast bundling with hot module replacement
- **TypeScript**: Compile-time type checking
- **Tree Shaking**: Automatic dead code elimination
- **Code Splitting**: Automatic route-based code splitting
- **Asset Optimization**: Automatic image and font optimization

### Production Configuration
- **Environment Variables**: Configured via build system
- **API Endpoints**: Configurable backend URL
- **Authentication**: Production Clerk configuration
- **Performance**: Optimized bundle size and loading

## Testing Strategy

### Component Testing
- React Testing Library for component interaction testing
- Jest for unit testing utilities and hooks
- MSW (Mock Service Worker) for API mocking

### Integration Testing
- End-to-end onboarding flow testing
- Authentication flow validation
- API integration testing with real backend

### Type Safety
- TypeScript for compile-time error detection
- Zod schemas for runtime validation
- API contract validation

## Security Considerations

### Authentication Security
- **JWT Validation**: Clerk handles token validation and refresh
- **Route Protection**: All sensitive routes require authentication
- **Session Management**: Automatic session timeout and renewal
- **OAuth Security**: Protected callback handling with state validation

### Data Security
- **Input Validation**: All forms validated client and server-side
- **HTTPS Enforcement**: Production deploys enforce HTTPS
- **CORS Configuration**: Backend configured for specific origins
- **Log Injection Prevention**: Sanitized error context in logging to prevent injection attacks
- **URI Decoding Safety**: Protected decodeURIComponent calls with error handling for malformed URLs

### Integration Security
- **Secure Callbacks**: OAuth callback handling with validation and error boundaries
- **API Token Management**: Secure storage and transmission of integration credentials
- **Error Sanitization**: Prevents sensitive data exposure in error messages
- **Request Validation**: All integration API requests include proper authentication and validation

### Privacy
- **Data Minimization**: Only collect necessary onboarding and integration data
- **Secure Storage**: Sensitive data stored server-side only
- **Audit Logging**: Backend tracks data access and modifications

## Performance Optimization

### Frontend Performance
- **Lazy Loading**: Route-based code splitting
- **Component Memoization**: React.memo for expensive components
- **Bundle Analysis**: Webpack bundle analyzer for optimization
- **Caching**: Browser caching for static assets

### API Performance
- **Request Batching**: Minimize API calls where possible
- **Loading States**: Immediate feedback for user actions
- **Error Boundaries**: Graceful error handling and recovery

## Browser Support

### Supported Browsers
- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

### Progressive Enhancement
- Core functionality works without JavaScript
- Enhanced experience with modern browser features
- Graceful degradation for older browsers

## Contributing

### Code Style
- **ESLint**: Configured with TypeScript and React rules
- **Prettier**: Automatic code formatting
- **Husky**: Pre-commit hooks for quality assurance

### Development Guidelines
- Use TypeScript for all new components
- Follow React hooks patterns
- Implement comprehensive error handling
- Write tests for complex business logic
- Document complex components and hooks

### Pull Request Process
1. Create feature branch from main
2. Implement changes with tests
3. Ensure all lints and builds pass
4. Submit PR with clear description
5. Address review feedback

## Troubleshooting

### Common Issues

#### Authentication Issues
```bash
# Clear Clerk cache
localStorage.clear()
sessionStorage.clear()

# Verify environment variables
echo $VITE_CLERK_PUBLISHABLE_KEY
```

#### Build Issues
```bash
# Clear node_modules and reinstall
rm -rf node_modules package-lock.json
npm install

# Clear Vite cache
rm -rf node_modules/.vite
```

#### API Connection Issues
```bash
# Verify backend is running
curl http://localhost:8080/api/v1/health

# Check CORS configuration
curl -H "Origin: http://localhost:5173" \
     -H "Access-Control-Request-Method: POST" \
     -H "Access-Control-Request-Headers: Authorization" \
     -X OPTIONS http://localhost:8080/api/v1/organizations/get
```

## License

This project is part of the InfraGPT platform and follows the project's licensing terms.