# InfraGPT Web Application

InfraGPT Web Application is a modern React TypeScript client that provides a comprehensive interface for AI-powered infrastructure management. The application features Clerk authentication, organization onboarding workflows, and seamless integration with the InfraGPT backend services.

## Architecture Overview

The application follows modern React patterns with TypeScript, featuring:

- **Authentication**: Clerk integration for user and organization management
- **Routing**: React Router for client-side navigation with protected routes
- **UI Components**: Radix UI with Tailwind CSS for modern, accessible design
- **State Management**: React hooks with custom API client
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
│   └── onboarding-constants.ts # Onboarding form constants and types
│
├── pages/                     # Page components
│   ├── login.tsx              # Login page
│   ├── signup.tsx             # Signup page
│   └── onboarding.tsx         # Organization onboarding page
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

### 3. UI/UX Features
- **Modern Design**: Clean, professional interface with Tailwind CSS
- **Responsive Layout**: Mobile-first design with adaptive layouts
- **Accessibility**: ARIA-compliant components from Radix UI
- **Loading States**: Comprehensive loading and error state management
- **Toast Notifications**: User feedback with Sonner toast system

### 4. Data Management
- **Type Safety**: Full TypeScript integration with Zod validation
- **Form Validation**: Real-time validation with React Hook Form
- **API Integration**: Type-safe API client with error handling
- **State Management**: Efficient state management with React hooks

## API Integration

### Backend Communication
The application communicates with the InfraGPT Go backend service:

```typescript
// API Client Configuration
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

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

### Data Security
- **Input Validation**: All forms validated client and server-side
- **HTTPS Enforcement**: Production deploys enforce HTTPS
- **CORS Configuration**: Backend configured for specific origins

### Privacy
- **Data Minimization**: Only collect necessary onboarding data
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