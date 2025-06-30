// Integration Manager - Main List Page

import { useEffect } from 'react';
import { observer } from 'mobx-react-lite';
import { useUser } from '@clerk/clerk-react';
import { integrationStore } from '../../stores/IntegrationStore';
import { userStore } from '../../stores/UserStore';
import { ConnectorGrid } from './components/ConnectorGrid';
import { Button } from '../../components/ui/button';
import { Card } from '../../components/ui/card';
import { Skeleton } from '../../components/ui/skeleton';
import { SidebarTrigger } from "@/components/ui/sidebar";
import { useApiClient } from '../../lib/api';

const IntegrationsPage = observer(() => {
  const { user } = useUser();
  const { getMe } = useApiClient();
  
  const clerkUserId = user?.id;
  const clerkOrgId = user?.organizationMemberships?.[0]?.organization?.id;

  // Load user profile on mount
  useEffect(() => {
    if (!userStore.userProfile && !userStore.loading && clerkUserId && clerkOrgId) {
      userStore.loadUserProfile(getMe, clerkUserId, clerkOrgId);
    }
  }, [getMe, clerkUserId, clerkOrgId]);

  // Load integrations once we have the user profile
  useEffect(() => {
    if (userStore.organizationId) {
      integrationStore.loadIntegrations(userStore.organizationId);
    }
  }, [userStore.organizationId]);

  const handleConnectorAction = async (connectorType: string, action: 'connect' | 'details') => {
    if (!userStore.organizationId || !userStore.userId) return;

    if (action === 'connect') {
      try {
        const response = await integrationStore.initiateConnection(
          connectorType as any,
          userStore.organizationId,
          userStore.userId,
          `${window.location.origin}/integrations/${connectorType}/authorize`
        );
        
        // Redirect to authorization URL
        if (response.type === 'redirect' || response.type === 'oauth2' || response.type === 'installation') {
          // open in a new tab 
          window.open(response.url, '_blank');

        } else if (response.type === 'popup') {
          // Handle popup flow (future enhancement)
          window.open(response.url, 'integration-auth', 'width=600,height=600');
        }
      } catch (error) {
        integrationStore.handleError(error, 'connecting integration');
      }
    } else if (action === 'details') {
      // Navigate to details page
      window.location.href = `/integrations/${connectorType}`;
    }
  };

  if ((userStore.loading || integrationStore.loading) && integrationStore.integrations.size === 0) {
    return (
      <div className="h-full flex flex-col">
        {/* Header */}
        <div className="border-b">
          <div className="flex h-16 items-center px-4 gap-4">
            <SidebarTrigger />
            <h1 className="text-xl font-semibold">Integrations</h1>
          </div>
        </div>
        
        {/* Loading Content */}
        <div className="flex-1 p-6 space-y-6">
          <div className="space-y-2">
            <Skeleton className="h-4 w-96" />
          </div>
          <div className="space-y-4">
            {Array.from({ length: 6 }).map((_, i) => (
              <Card key={i} className="p-4">
                <div className="flex items-center space-x-4">
                  <Skeleton className="h-10 w-10 rounded" />
                  <div className="flex-1 space-y-2">
                    <Skeleton className="h-5 w-32" />
                    <Skeleton className="h-4 w-64" />
                  </div>
                  <Skeleton className="h-9 w-24" />
                </div>
              </Card>
            ))}
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="border-b">
        <div className="flex h-16 items-center px-4 gap-4">
          <SidebarTrigger />
          <h1 className="text-xl font-semibold">Integrations</h1>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 p-6 space-y-6">
        {/* Subtext */}
        <p className="text-muted-foreground">
          Connect your tools to streamline DevOps workflows and automate your infrastructure management.
        </p>

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


        {/* Connectors Grid */}
        <ConnectorGrid
          connectors={integrationStore.connectorsWithStatus}
          integrations={Array.from(integrationStore.integrations.values())}
          onConnectorAction={handleConnectorAction}
          loadingConnectors={integrationStore.loadingConnectors}
        />

      </div>
    </div>
  );
});

export default IntegrationsPage;