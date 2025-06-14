// Integration Manager Type Definitions

export type ConnectorType = 
  | 'slack' 
  | 'github' 
  | 'aws' 
  | 'gcp' 
  | 'pagerduty' 
  | 'datadog';

export type AuthType = 'oauth2' | 'app_installation' | 'api_key';

export type IntegrationStatus = 'connected' | 'disconnected' | 'error' | 'syncing';

export type ConnectorStatus = 'available' | 'connected' | 'coming_soon';

export interface Connector {
  type: ConnectorType;
  name: string;
  description: string;
  logo: string;
  isImplemented: boolean;
  authType: AuthType;
  capabilities: readonly string[];
  status?: ConnectorStatus; // Will be determined dynamically
}

export interface Integration {
  id: string;
  organizationId: string;
  connectorType: ConnectorType;
  status: IntegrationStatus;
  createdAt: string;
  updatedAt: string;
  lastSyncAt?: string;
  metadata?: Record<string, any>;
  configuration?: IntegrationConfiguration;
}

export interface IntegrationConfiguration {
  // Slack specific
  workspaceName?: string;
  teamId?: string;
  connectedChannels?: string[];
  permissions?: string[];
  
  // GitHub specific
  installationId?: number;
  repositoryCount?: number;
  connectedRepos?: string[];
  webhookUrl?: string;
  
  // AWS specific
  region?: string;
  accountId?: string;
  
  // GCP specific
  projectId?: string;
  
  // PagerDuty specific
  serviceName?: string;
  
  // Datadog specific
  siteName?: string;
}

export interface AuthorizationIntent {
  type: 'redirect' | 'popup';
  url: string;
  state?: string;
  expires_at?: string;
}

// API Request/Response Types
export interface AuthorizeRequest {
  organization_id: string;
  connector_type: ConnectorType;
  redirect_url?: string;
}

export interface AuthorizeResponse {
  type: 'redirect' | 'popup';
  url: string;
  state?: string;
}

export interface CallbackRequest {
  connector_type: ConnectorType;
  code?: string;
  state?: string;
  installation_id?: string;
  setup_action?: string;
  [key: string]: any; // Allow for connector-specific parameters
}

export interface IntegrationsListRequest {
  organization_id: string;
}

export interface IntegrationsListResponse {
  integrations: Integration[];
}

export interface IntegrationDetailsRequest {
  organization_id: string;
  connector_type: ConnectorType;
}

export interface IntegrationStatusRequest {
  integration_id: string;
}

export interface RevokeIntegrationRequest {
  integration_id: string;
}

export interface TestConnectionRequest {
  integration_id: string;
}

export interface TestConnectionResponse {
  success: boolean;
  message?: string;
  last_tested_at: string;
}

// Activity Log Types
export interface IntegrationActivity {
  id: string;
  integrationId: string;
  activityType: 'connected' | 'disconnected' | 'error' | 'sync' | 'config_updated';
  description: string;
  timestamp: string;
  metadata?: Record<string, any>;
}

// Error Types
export class IntegrationError extends Error {
  constructor(
    public statusCode: number,
    message: string,
    public connectorType?: ConnectorType
  ) {
    super(message);
    this.name = 'IntegrationError';
  }
}

// UI State Types
export interface IntegrationUIState {
  loading: boolean;
  error: string | null;
  selectedConnector?: ConnectorType;
  showDetails: boolean;
}

// Connector Capabilities
export const CONNECTOR_CAPABILITIES = {
  slack: ['messaging', 'notifications', 'channels', 'threads'],
  github: ['repositories', 'webhooks', 'pull_requests', 'issues'],
  aws: ['compute', 'storage', 'networking', 'monitoring'],
  gcp: ['compute', 'storage', 'databases', 'ai_ml'],
  pagerduty: ['incident_management', 'alerting', 'escalation'],
  datadog: ['monitoring', 'analytics', 'dashboards', 'alerts']
} as const;