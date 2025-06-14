// Integration Details Page - Configuration and Status View

import { useEffect, useState } from 'react';
import { observer } from 'mobx-react-lite';
import { useParams, useNavigate } from 'react-router-dom';
import { useUser } from '@clerk/clerk-react';
import { integrationStore } from '../../stores/IntegrationStore';
import { getConnectorByType } from '../../lib/integration-constants';
import { IntegrationStatus } from './components/IntegrationStatus';
import { IntegrationConfiguration } from './components/IntegrationConfiguration';
import { IntegrationActions } from './components/IntegrationActions';
import { Button } from '../../components/ui/button';
import { Card } from '../../components/ui/card';
import { ArrowLeft } from 'lucide-react';

const IntegrationDetailsPage = observer(() => {
  const { connectorType } = useParams<{ connectorType: string }>();
  const navigate = useNavigate();
  const { user } = useUser();
  const organizationId = user?.organizationMemberships?.[0]?.organization?.id;
  
  const [loading, setLoading] = useState(true);

  const connector = connectorType ? getConnectorByType(connectorType as any) : null;
  const integration = connectorType ? integrationStore.getIntegrationByConnectorType(connectorType as any) : null;

  useEffect(() => {
    const loadData = async () => {
      if (!organizationId || !connectorType || !connector) {
        setLoading(false);
        return;
      }

      try {
        // Load integration details if it exists
        if (integrationStore.isConnectorConnected(connectorType as any)) {
          await integrationStore.getIntegrationDetails(organizationId, connectorType as any);
        }
        
        // Load activity log if integration exists
        if (integration) {
          await integrationStore.loadIntegrationActivity(integration.id);
        }
      } catch (error) {
        integrationStore.handleError(error, 'loading integration details');
      } finally {
        setLoading(false);
      }
    };

    loadData();
  }, [organizationId, connectorType, connector, integration?.id]);

  const handleTestConnection = async () => {
    if (!integration) return;
    
    try {
      const result = await integrationStore.testConnection(integration.id);
      if (result.success) {
        // Show success message (you can use your toast system)
        console.log('Connection test successful');
      } else {
        // Show error message
        console.error('Connection test failed:', result.message);
      }
    } catch (error) {
      integrationStore.handleError(error, 'testing connection');
    }
  };

  const handleReconfigure = async () => {
    if (!organizationId || !connectorType) return;
    
    try {
      const response = await integrationStore.initiateConnection(
        connectorType as any,
        organizationId,
        `${window.location.origin}/integrations/${connectorType}/callback`
      );
      
      // Redirect to authorization URL
      if (response.type === 'redirect') {
        window.location.href = response.url;
      }
    } catch (error) {
      integrationStore.handleError(error, 'reconfiguring integration');
    }
  };

  const handleDisconnect = async () => {
    if (!integration) return;
    
    // Show confirmation dialog
    const confirmed = window.confirm(
      `Are you sure you want to disconnect ${connector?.name}? This will remove all associated configurations and data.`
    );
    
    if (!confirmed) return;
    
    try {
      await integrationStore.revokeIntegration(integration.id);
      // Navigate back to integrations list
      navigate('/integrations');
    } catch (error) {
      integrationStore.handleError(error, 'disconnecting integration');
    }
  };

  if (!connector) {
    return (
      <div className="space-y-6">
        <Button variant="ghost" onClick={() => navigate('/integrations')}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to Integrations
        </Button>
        <Card className="p-6">
          <div className="text-center">
            <h2 className="text-xl font-semibold mb-2">Integration Not Found</h2>
            <p className="text-muted-foreground">
              The requested integration type could not be found.
            </p>
          </div>
        </Card>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <Button variant="ghost" onClick={() => navigate('/integrations')}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to Integrations
        </Button>
        <div className="flex items-center justify-center py-12">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
            <p className="text-muted-foreground">Loading integration details...</p>
          </div>
        </div>
      </div>
    );
  }

  if (!integration) {
    return (
      <div className="space-y-6">
        <Button variant="ghost" onClick={() => navigate('/integrations')}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to Integrations
        </Button>
        
        {/* Header */}
        <div className="flex items-center space-x-4">
          <div className="h-12 w-12 rounded-lg bg-gray-100 flex items-center justify-center">
            <img 
              src={connector.logo} 
              alt={connector.name}
              className="h-8 w-8"
              onError={(e) => {
                // Fallback for missing logos
                (e.target as HTMLImageElement).style.display = 'none';
              }}
            />
          </div>
          <div>
            <h1 className="text-2xl font-bold">{connector.name} Integration</h1>
            <p className="text-muted-foreground">{connector.description}</p>
          </div>
        </div>

        <Card className="p-6">
          <div className="text-center space-y-4">
            <h2 className="text-xl font-semibold">Not Connected</h2>
            <p className="text-muted-foreground">
              This integration is not currently connected to your organization.
            </p>
            <Button 
              onClick={() => navigate('/integrations')}
              className="mt-4"
            >
              Connect {connector.name}
            </Button>
          </div>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Navigation */}
      <Button variant="ghost" onClick={() => navigate('/integrations')}>
        <ArrowLeft className="h-4 w-4 mr-2" />
        Back to Integrations
      </Button>

      {/* Header */}
      <div className="flex items-center space-x-4">
        <div className="h-12 w-12 rounded-lg bg-gray-100 flex items-center justify-center">
          <img 
            src={connector.logo} 
            alt={connector.name}
            className="h-8 w-8"
            onError={(e) => {
              (e.target as HTMLImageElement).style.display = 'none';
            }}
          />
        </div>
        <div>
          <h1 className="text-2xl font-bold">{connector.name} Integration</h1>
          <p className="text-muted-foreground">{connector.description}</p>
        </div>
      </div>

      {/* Error Display */}
      {integrationStore.error && (
        <Card className="border-destructive bg-destructive/5 p-4">
          <div className="flex items-center justify-between">
            <p className="text-sm text-destructive">{integrationStore.error}</p>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => integrationStore.clearError()}
            >
              ✕
            </Button>
          </div>
        </Card>
      )}

      {/* Status Section */}
      <IntegrationStatus 
        integration={integration}
        connector={connector}
      />

      {/* Configuration Section */}
      <IntegrationConfiguration 
        integration={integration}
        connector={connector}
      />

      {/* Actions Section */}
      <IntegrationActions
        integration={integration}
        connector={connector}
        onTestConnection={handleTestConnection}
        onReconfigure={handleReconfigure}
        onDisconnect={handleDisconnect}
        loading={integrationStore.isConnectorLoading(connector.type as any)}
      />

      {/* Activity Log */}
      {integration && integrationStore.activities.has(integration.id) && (
        <Card className="p-6">
          <h3 className="text-lg font-semibold mb-4">Recent Activity</h3>
          <div className="space-y-3">
            {integrationStore.activities.get(integration.id)?.slice(0, 5).map((activity) => (
              <div key={activity.id} className="flex items-start space-x-3 text-sm">
                <div className="h-2 w-2 rounded-full bg-gray-400 mt-2 flex-shrink-0" />
                <div className="flex-1">
                  <p>{activity.description}</p>
                  <p className="text-muted-foreground text-xs">
                    {new Date(activity.timestamp).toLocaleString()}
                  </p>
                </div>
              </div>
            ))}
            {integrationStore.activities.get(integration.id)?.length === 0 && (
              <p className="text-muted-foreground text-sm">No recent activity</p>
            )}
          </div>
        </Card>
      )}
    </div>
  );
});

export default IntegrationDetailsPage;