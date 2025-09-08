import { useEffect, useState } from "react";
import { observer } from "mobx-react-lite";
import { useParams, useNavigate } from "react-router-dom";
import { useUser } from "@clerk/clerk-react";
import { integrationStore } from "../../stores/IntegrationStore";
import { userStore } from "../../stores/UserStore";
import { getConnectorByType } from "../../lib/integration-constants";
import { IntegrationStatus } from "./components/IntegrationStatus";
import { IntegrationConfiguration } from "./components/IntegrationConfiguration";
import { IntegrationActions } from "./components/IntegrationActions";
import { Button } from "../../components/ui/button";
import { Card } from "../../components/ui/card";
import { ArrowLeft } from "lucide-react";
import { useApiClient } from "../../lib/api";

const IntegrationDetailsPage = observer(() => {
  const { connectorType } = useParams<{ connectorType: string }>();
  const navigate = useNavigate();
  const { user } = useUser();
  const { getMe } = useApiClient();
  const [loading, setLoading] = useState(true);

  const clerkUserId = user?.id;
  const clerkOrgId = user?.organizationMemberships?.[0]?.organization?.id;

  const connector = connectorType
    ? getConnectorByType(connectorType as any)
    : null;
  const integration = connectorType
    ? integrationStore.getIntegrationByConnectorType(connectorType as any)
    : null;

  useEffect(() => {
    if (
      !userStore.userProfile &&
      !userStore.loading &&
      clerkUserId &&
      clerkOrgId
    ) {
      userStore.loadUserProfile(getMe, clerkUserId, clerkOrgId);
    }
  }, [getMe, clerkUserId, clerkOrgId]);

  useEffect(() => {
    if (
      userStore.organizationId &&
      integrationStore.integrations.size === 0 &&
      !integrationStore.loading
    ) {
      integrationStore.loadIntegrations(userStore.organizationId);
    }
  }, [userStore.organizationId]);

  useEffect(() => {
    const loadData = async () => {
      if (!connector || !userStore.organizationId) {
        setLoading(false);
        return;
      }

      setLoading(false);
    };

    if (userStore.organizationId && !userStore.loading) {
      loadData();
    }
  }, [connector, integration?.id, userStore.organizationId, userStore.loading]);

  const handleTestConnection = async () => {
    if (!integration) return;

    try {
      const result = await integrationStore.testConnection(integration.id);
      if (result.success) {
        console.log("Connection test successful");
      } else {
        console.error("Connection test failed:", result.message);
      }
    } catch (error) {
      integrationStore.handleError(error, "testing connection");
    }
  };

  const handleReconfigure = async () => {
    if (!userStore.organizationId || !userStore.userId || !connectorType)
      return;

    try {
      const response = await integrationStore.initiateConnection(
        connectorType as any,
        userStore.organizationId,
        userStore.userId,
        `${window.location.origin}/integrations/${connectorType}/authorize`,
      );

      if (response.type === "redirect" || response.type === "oauth2") {
        window.location.href = response.url;
      } else if (response.type === "popup") {
        window.open(response.url, "integration-auth", "width=600,height=600");
      }
    } catch (error) {
      integrationStore.handleError(error, "reconfiguring integration");
    }
  };

  const handleDisconnect = async () => {
    if (!integration) return;

    const confirmed = window.confirm(
      `Are you sure you want to disconnect ${connector?.name}? This will remove all associated configurations and data.`,
    );

    if (!confirmed) return;

    try {
      await integrationStore.revokeIntegration(integration.id);
      navigate("/integrations");
    } catch (error) {
      integrationStore.handleError(error, "disconnecting integration");
    }
  };

  if (!connector) {
    return (
      <div className="h-full flex flex-col">
        {/* Header */}
        <div className="border-b">
          <div className="flex h-16 items-center px-4 gap-4">
            <Button
              variant="ghost"
              onClick={() => navigate("/integrations")}
              className="p-0"
            >
              <ArrowLeft className="h-4 w-4 mr-2" />
              Back to Integrations
            </Button>
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 p-6">
          <Card className="p-6">
            <div className="text-center">
              <h2 className="text-xl font-semibold mb-2">
                Integration Not Found
              </h2>
              <p className="text-muted-foreground">
                The requested integration type could not be found.
              </p>
            </div>
          </Card>
        </div>
      </div>
    );
  }

  if (
    userStore.loading ||
    integrationStore.loading ||
    (loading && !userStore.organizationId)
  ) {
    return (
      <div className="h-full flex flex-col">
        {/* Header */}
        <div className="border-b">
          <div className="flex h-16 items-center px-4 gap-4">
            <Button
              variant="ghost"
              onClick={() => navigate("/integrations")}
              className="p-0"
            >
              <ArrowLeft className="h-4 w-4 mr-2" />
              Back to Integrations
            </Button>
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 p-6 flex items-center justify-center">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
            <p className="text-muted-foreground">
              Loading integration details...
            </p>
          </div>
        </div>
      </div>
    );
  }

  if (!integration) {
    return (
      <div className="h-full flex flex-col">
        {/* Header */}
        <div className="border-b">
          <div className="flex h-16 items-center px-4 gap-4">
            <Button
              variant="ghost"
              onClick={() => navigate("/integrations")}
              className="p-0"
            >
              <ArrowLeft className="h-4 w-4 mr-2" />
              Back to Integrations
            </Button>
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 p-6 space-y-6">
          {/* Header */}
          <div className="flex items-center space-x-4">
            <div className="h-12 w-12 rounded-lg bg-gray-100 flex items-center justify-center">
              <img
                src={connector.logo}
                alt={connector.name}
                className="h-8 w-8"
                onError={(e) => {
                  (e.target as HTMLImageElement).style.display = "none";
                }}
              />
            </div>
            <div>
              <h1 className="text-2xl font-bold">
                {connector.name} Integration
              </h1>
              <p className="text-muted-foreground">{connector.description}</p>
            </div>
          </div>

          <Card className="p-6">
            <div className="text-center space-y-4">
              <h2 className="text-xl font-semibold">Not Connected</h2>
              <p className="text-muted-foreground">
                This integration is not currently connected to your
                organization.
              </p>
              <Button
                onClick={() => navigate("/integrations")}
                className="mt-4"
              >
                Connect {connector.name}
              </Button>
            </div>
          </Card>
        </div>
      </div>
    );
  }

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="border-b">
        <div className="flex h-16 items-center px-4 gap-4">
          <Button
            variant="ghost"
            onClick={() => navigate("/integrations")}
            className="p-0"
          >
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to Integrations
          </Button>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 p-6 space-y-6">
        {/* Header */}
        <div className="flex items-center space-x-4">
          <div className="h-12 w-12 rounded-lg bg-gray-100 flex items-center justify-center">
            <img
              src={connector.logo}
              alt={connector.name}
              className="h-8 w-8"
              onError={(e) => {
                (e.target as HTMLImageElement).style.display = "none";
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
              <p className="text-sm text-destructive">
                {integrationStore.error}
              </p>
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
        <IntegrationStatus integration={integration} connector={connector} />

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

        {/* Activity Log - Disabled until API is implemented */}
        {false && integration && (
          <Card className="p-6">
            <h3 className="text-lg font-semibold mb-4">Recent Activity</h3>
            <div className="space-y-3">
              <p className="text-muted-foreground text-sm">
                Activity log will be available soon
              </p>
            </div>
          </Card>
        )}
      </div>
    </div>
  );
});

export default IntegrationDetailsPage;
