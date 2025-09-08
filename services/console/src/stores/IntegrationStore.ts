import { makeAutoObservable, runInAction } from "mobx";
import {
  Integration,
  Connector,
  ConnectorType,
  ConnectorStatus,
  AuthorizeResponse,
  TestConnectionResponse,
  IntegrationActivity,
} from "../types/integration";
import { CONNECTORS, getConnectorByType } from "../lib/integration-constants";
import {
  integrationService,
  getErrorMessage,
  withErrorHandling,
} from "../services/integrationService";

class IntegrationStore {
  integrations: Map<string, Integration> = new Map();
  connectors: Connector[] = CONNECTORS;
  loading = false;
  loadingConnectors: Set<ConnectorType> = new Set();
  error: string | null = null;
  activities: Map<string, IntegrationActivity[]> = new Map();

  selectedConnector: ConnectorType | null = null;
  showDetails = false;

  constructor() {
    makeAutoObservable(this);
  }

  private transformAPIIntegration(apiIntegration: {
    id: string;
    organization_id: string;
    connector_type: string;
    status: string;
    created_at: string;
    updated_at: string;
    last_sync_at?: string;
    metadata?: Record<string, unknown>;
    configuration?: Record<string, unknown>;
  }): Integration {
    return {
      id: apiIntegration.id,
      organizationId: apiIntegration.organization_id,
      connectorType: apiIntegration.connector_type,
      status: apiIntegration.status,
      createdAt: apiIntegration.created_at,
      updatedAt: apiIntegration.updated_at,
      lastSyncAt: apiIntegration.last_used_at,
      metadata: apiIntegration.metadata,
      configuration: {
        workspaceName: apiIntegration.metadata?.connector_org_name,
      },
    };
  }

  get connectedIntegrations(): Integration[] {
    return Array.from(this.integrations.values()).filter(
      (integration) =>
        integration.status === "active" || integration.status === "connected",
    );
  }

  get availableConnectors(): Connector[] {
    return this.connectors.filter(
      (connector) =>
        connector.isImplemented && !this.isConnectorConnected(connector.type),
    );
  }

  get comingSoonConnectors(): Connector[] {
    return this.connectors.filter((connector) => !connector.isImplemented);
  }

  get connectorsWithStatus(): Connector[] {
    return this.connectors.map((connector) => ({
      ...connector,
      status: this.getConnectorStatus(connector.type),
    }));
  }

  getConnectorStatus(connectorType: ConnectorType): ConnectorStatus {
    const connector = getConnectorByType(connectorType);

    if (!connector?.isImplemented) {
      return "coming_soon";
    }

    return this.isConnectorConnected(connectorType) ? "connected" : "available";
  }

  isConnectorConnected(connectorType: ConnectorType): boolean {
    return Array.from(this.integrations.values()).some(
      (integration) =>
        integration.connectorType === connectorType &&
        (integration.status === "active" || integration.status === "connected"),
    );
  }

  getIntegrationByConnectorType(
    connectorType: ConnectorType,
  ): Integration | undefined {
    return Array.from(this.integrations.values()).find(
      (integration) => integration.connectorType === connectorType,
    );
  }

  isConnectorLoading(connectorType: ConnectorType): boolean {
    return this.loadingConnectors.has(connectorType);
  }

  async loadIntegrations(organizationId: string): Promise<void> {
    await withErrorHandling(async () => {
      runInAction(() => {
        this.loading = true;
        this.error = null;
      });

      const apiResponse =
        await integrationService.getIntegrations(organizationId);

      runInAction(() => {
        this.integrations.clear();
        apiResponse.forEach((apiIntegration) => {
          const integration = this.transformAPIIntegration(apiIntegration);
          this.integrations.set(integration.id, integration);
        });
        this.loading = false;
      });
    }, "loading integrations");
  }

  async initiateConnection(
    connectorType: ConnectorType,
    organizationId: string,
    userId: string,
    redirectUrl?: string,
  ): Promise<AuthorizeResponse> {
    return withErrorHandling(async () => {
      runInAction(() => {
        this.loadingConnectors.add(connectorType);
        this.error = null;
      });

      try {
        const response = await integrationService.initiateAuthorization(
          organizationId,
          userId,
          connectorType,
          redirectUrl,
        );

        return response;
      } finally {
        runInAction(() => {
          this.loadingConnectors.delete(connectorType);
        });
      }
    }, `initiating connection for ${connectorType}`);
  }

  async handleCallback(
    connectorType: ConnectorType,
    callbackData: Record<string, unknown>,
  ): Promise<Integration> {
    return withErrorHandling(async () => {
      runInAction(() => {
        this.loadingConnectors.add(connectorType);
        this.error = null;
      });

      try {
        const integration = await integrationService.handleCallback(
          connectorType,
          callbackData,
        );

        runInAction(() => {
          this.integrations.set(integration.id, integration);
        });

        return integration;
      } finally {
        runInAction(() => {
          this.loadingConnectors.delete(connectorType);
        });
      }
    }, `handling callback for ${connectorType}`);
  }

  async getIntegrationDetails(
    organizationId: string,
    connectorType: ConnectorType,
  ): Promise<Integration> {
    return withErrorHandling(async () => {
      runInAction(() => {
        this.loadingConnectors.add(connectorType);
      });

      try {
        const integration = await integrationService.getIntegrationDetails(
          organizationId,
          connectorType,
        );

        runInAction(() => {
          this.integrations.set(integration.id, integration);
        });

        return integration;
      } finally {
        runInAction(() => {
          this.loadingConnectors.delete(connectorType);
        });
      }
    }, `getting details for ${connectorType}`);
  }

  async testConnection(integrationId: string): Promise<TestConnectionResponse> {
    return withErrorHandling(async () => {
      const integration = this.integrations.get(integrationId);
      if (!integration) {
        throw new Error("Integration not found");
      }

      runInAction(() => {
        this.loadingConnectors.add(integration.connectorType);
      });

      try {
        const result = await integrationService.testConnection(integrationId);

        runInAction(() => {
          const updatedIntegration = {
            ...integration,
            metadata: {
              ...integration.metadata,
              lastTestedAt: result.last_tested_at,
            },
          };
          this.integrations.set(integrationId, updatedIntegration);
        });

        return result;
      } finally {
        runInAction(() => {
          this.loadingConnectors.delete(integration.connectorType);
        });
      }
    }, `testing connection for ${integrationId}`);
  }

  async revokeIntegration(integrationId: string): Promise<void> {
    await withErrorHandling(async () => {
      const integration = this.integrations.get(integrationId);
      if (!integration) {
        throw new Error("Integration not found");
      }

      runInAction(() => {
        this.loadingConnectors.add(integration.connectorType);
      });

      try {
        await integrationService.revokeIntegration(
          integrationId,
          integration.organizationId,
        );

        runInAction(() => {
          this.integrations.delete(integrationId);
        });
      } finally {
        runInAction(() => {
          this.loadingConnectors.delete(integration.connectorType);
        });
      }
    }, `revoking integration ${integrationId}`);
  }

  async loadIntegrationActivity(integrationId: string): Promise<void> {
    await withErrorHandling(async () => {
      const activities =
        await integrationService.getIntegrationActivity(integrationId);

      runInAction(() => {
        this.activities.set(integrationId, activities);
      });
    }, `loading activity for ${integrationId}`);
  }

  async refreshIntegrationStatus(integrationId: string): Promise<void> {
    await withErrorHandling(async () => {
      const integration = this.integrations.get(integrationId);
      if (!integration) return;

      runInAction(() => {
        this.loadingConnectors.add(integration.connectorType);
      });

      try {
        const updatedIntegration =
          await integrationService.getIntegrationStatus(integrationId);

        runInAction(() => {
          this.integrations.set(integrationId, updatedIntegration);
        });
      } finally {
        runInAction(() => {
          this.loadingConnectors.delete(integration.connectorType);
        });
      }
    }, `refreshing status for ${integrationId}`);
  }

  setSelectedConnector(connectorType: ConnectorType | null): void {
    this.selectedConnector = connectorType;
  }

  setShowDetails(show: boolean): void {
    this.showDetails = show;
  }

  clearError(): void {
    this.error = null;
  }

  handleError(error: unknown, context?: string): void {
    const errorMessage = getErrorMessage(error);

    runInAction(() => {
      this.error = errorMessage;
      console.error(
        `Integration store error${context ? ` (${context})` : ""}:`,
        error,
      );
    });
  }

  reset(): void {
    runInAction(() => {
      this.integrations.clear();
      this.activities.clear();
      this.loadingConnectors.clear();
      this.loading = false;
      this.error = null;
      this.selectedConnector = null;
      this.showDetails = false;
    });
  }
}

export const integrationStore = new IntegrationStore();

export { IntegrationStore };
