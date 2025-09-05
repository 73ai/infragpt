// Integration Manager Constants and Configuration

import { Connector, ConnectorType, CONNECTOR_CAPABILITIES } from '../types/integration';

export const CONNECTORS: Connector[] = [
  {
    type: 'slack',
    name: 'Slack',
    description: 'Real-time messaging and notifications',
    logo: '/icons/slack.svg',
    isImplemented: true,
    authType: 'oauth2',
    capabilities: CONNECTOR_CAPABILITIES.slack
  },
  {
    type: 'github',
    name: 'GitHub',
    description: 'Code repository and PR management',
    logo: '/icons/github.svg',
    isImplemented: true,
    authType: 'app_installation',
    capabilities: CONNECTOR_CAPABILITIES.github
  },
  {
    type: 'aws',
    name: 'AWS',
    description: 'Cloud infrastructure management',
    logo: '/icons/aws.svg',
    isImplemented: false,
    authType: 'api_key',
    capabilities: CONNECTOR_CAPABILITIES.aws
  },
  {
    type: 'gcp',
    name: 'Google Cloud',
    description: 'Cloud platform and services',
    logo: '/icons/gcp.svg',
    isImplemented: true,
    authType: 'api_key',
    capabilities: CONNECTOR_CAPABILITIES.gcp
  },
  {
    type: 'pagerduty',
    name: 'PagerDuty',
    description: 'Incident response and alerting',
    logo: '/icons/pagerduty.svg',
    isImplemented: false,
    authType: 'api_key',
    capabilities: CONNECTOR_CAPABILITIES.pagerduty
  },
  {
    type: 'datadog',
    name: 'Datadog',
    description: 'Monitoring and analytics platform',
    logo: '/icons/datadog.svg',
    isImplemented: false,
    authType: 'api_key',
    capabilities: CONNECTOR_CAPABILITIES.datadog
  }
];

export const getConnectorByType = (type: ConnectorType): Connector | undefined => {
  return CONNECTORS.find(connector => connector.type === type);
};

export const getImplementedConnectors = (): Connector[] => {
  return CONNECTORS.filter(connector => connector.isImplemented);
};

export const getAvailableConnectors = (): Connector[] => {
  return CONNECTORS.filter(connector => !connector.isImplemented);
};

// UI Constants
export const INTEGRATION_ROUTES = {
  LIST: '/integrations',
  DETAILS: (connectorType: ConnectorType) => `/integrations/${connectorType}`,
  CALLBACK: (connectorType: ConnectorType) => `/integrations/${connectorType}/callback`
} as const;

// Status Colors and Icons
export const STATUS_CONFIG = {
  active: {
    color: 'text-green-600',
    bgColor: 'bg-green-50',
    borderColor: 'border-green-200',
    icon: '🟢',
    text: 'Connected'
  },
  connected: {
    color: 'text-green-600',
    bgColor: 'bg-green-50',
    borderColor: 'border-green-200',
    icon: '✅',
    text: 'Connected'
  },
  inactive: {
    color: 'text-gray-600',
    bgColor: 'bg-gray-50',
    borderColor: 'border-gray-200',
    icon: '⚫',
    text: 'Inactive'
  },
  pending: {
    color: 'text-yellow-600',
    bgColor: 'bg-yellow-50',
    borderColor: 'border-yellow-200',
    icon: '⏳',
    text: 'Pending'
  },
  not_started: {
    color: 'text-gray-600',
    bgColor: 'bg-gray-50',
    borderColor: 'border-gray-200',
    icon: '⏸️',
    text: 'Not Started'
  },
  suspended: {
    color: 'text-orange-600',
    bgColor: 'bg-orange-50',
    borderColor: 'border-orange-200',
    icon: '⏸️',
    text: 'Suspended'
  },
  deleted: {
    color: 'text-red-600',
    bgColor: 'bg-red-50',
    borderColor: 'border-red-200',
    icon: '🗑️',
    text: 'Deleted'
  },
  disconnected: {
    color: 'text-gray-600',
    bgColor: 'bg-gray-50',
    borderColor: 'border-gray-200',
    icon: '⚫',
    text: 'Disconnected'
  },
  error: {
    color: 'text-red-600',
    bgColor: 'bg-red-50',
    borderColor: 'border-red-200',
    icon: '❌',
    text: 'Error'
  },
  syncing: {
    color: 'text-blue-600',
    bgColor: 'bg-blue-50',
    borderColor: 'border-blue-200',
    icon: '🔄',
    text: 'Syncing'
  }
} as const;

// Button Text Configuration
export const BUTTON_TEXT = {
  connect: 'Connect',
  show_details: 'View Details',
  coming_soon: 'Coming Soon'
} as const;

// Error Messages
export const ERROR_MESSAGES = {
  LOAD_INTEGRATIONS_FAILED: 'Failed to load integrations. Please try again.',
  CONNECTION_FAILED: 'Failed to connect integration. Please try again.',
  AUTHORIZATION_FAILED: 'Authorization failed. Please try again.',
  TEST_CONNECTION_FAILED: 'Connection test failed. Please check your configuration.',
  REVOKE_FAILED: 'Failed to disconnect integration. Please try again.',
  NETWORK_ERROR: 'Network error. Please check your connection and try again.',
  UNAUTHORIZED: 'You are not authorized to perform this action.',
  NOT_FOUND: 'Integration not found.',
  UNKNOWN_ERROR: 'An unexpected error occurred. Please try again.'
} as const;

// Success Messages
export const SUCCESS_MESSAGES = {
  CONNECTION_SUCCESS: 'Integration connected successfully!',
  REVOKE_SUCCESS: 'Integration disconnected successfully.',
  TEST_CONNECTION_SUCCESS: 'Connection test successful.',
  CONFIGURATION_UPDATED: 'Configuration updated successfully.'
} as const;