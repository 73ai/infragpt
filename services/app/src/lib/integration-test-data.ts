// Test data for integration manager development

import { Integration, ConnectorType } from '../types/integration';

// Mock integration data for testing
export const mockIntegrations: Integration[] = [
  {
    id: 'slack-integration-1',
    organizationId: 'org-123',
    connectorType: 'slack' as ConnectorType,
    status: 'connected',
    createdAt: '2024-01-15T10:00:00Z',
    updatedAt: '2024-01-20T15:30:00Z',
    lastSyncAt: '2024-01-20T15:25:00Z',
    configuration: {
      workspaceName: 'InfraGPT Team',
      teamId: 'T1234567890',
      connectedChannels: ['general', 'alerts', 'deployments', 'infrastructure'],
      permissions: ['bot:basic', 'channels:read', 'chat:write', 'users:read']
    }
  },
  {
    id: 'github-integration-1',
    organizationId: 'org-123',
    connectorType: 'github' as ConnectorType,
    status: 'error',
    createdAt: '2024-01-10T08:00:00Z',
    updatedAt: '2024-01-20T12:00:00Z',
    lastSyncAt: '2024-01-19T18:45:00Z',
    configuration: {
      installationId: 12345678,
      repositoryCount: 15,
      connectedRepos: ['infragpt/core', 'infragpt/agent', 'infragpt/web'],
      webhookUrl: 'https://api.infragpt.io/webhooks/github'
    },
    metadata: {
      errorMessage: 'Webhook delivery failed - invalid URL',
      errorCode: 'WEBHOOK_DELIVERY_FAILED'
    }
  }
];

// Helper function to simulate API delay
export const simulateApiDelay = (ms: number = 1000): Promise<void> => {
  return new Promise(resolve => setTimeout(resolve, ms));
};

// Mock API responses for testing
export const mockApiResponses = {
  getIntegrations: async (organizationId: string): Promise<Integration[]> => {
    await simulateApiDelay(800);
    return mockIntegrations.filter(integration => 
      integration.organizationId === organizationId
    );
  },

  testConnection: async (integrationId: string) => {
    await simulateApiDelay(2000);
    const integration = mockIntegrations.find(i => i.id === integrationId);
    
    if (!integration) {
      throw new Error('Integration not found');
    }

    if (integration.status === 'error') {
      return {
        success: false,
        message: 'Connection test failed: Webhook endpoint unreachable',
        last_tested_at: new Date().toISOString()
      };
    }

    return {
      success: true,
      message: 'Connection test successful',
      last_tested_at: new Date().toISOString()
    };
  }
};

// Environment flag for using mock data
// export const USE_MOCK_DATA = import.meta.env.NODE_ENV === 'development' && 
//   import.meta.env.VITE_USE_MOCK_INTEGRATIONS === 'true';

export const USE_MOCK_DATA = true;