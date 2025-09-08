import {
  Integration,
  AuthorizeRequest,
  AuthorizeResponse,
  CallbackRequest,
  IntegrationsListRequest,
  IntegrationsListResponse,
  IntegrationDetailsRequest,
  IntegrationStatusRequest,
  RevokeIntegrationRequest,
  TestConnectionRequest,
  TestConnectionResponse,
  IntegrationError,
  ConnectorType,
  IntegrationActivity,
} from "../types/integration";

const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL || "http://localhost:8080";
const INTEGRATION_API_PREFIX = "/integrations";

class IntegrationService {
  private async request<T>(
    endpoint: string,
    options: RequestInit = {},
  ): Promise<T> {
    const url = `${API_BASE_URL}${INTEGRATION_API_PREFIX}${endpoint}`;

    try {
      const authToken = await this.getAuthToken();

      const response = await fetch(url, {
        ...options,
        headers: {
          Authorization: `Bearer ${authToken}`,
          "Content-Type": "application/json",
          ...options.headers,
        },
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new IntegrationError(
          response.status,
          errorText || `HTTP ${response.status}: ${response.statusText}`,
        );
      }

      return response.json();
    } catch (error) {
      if (error instanceof IntegrationError) {
        throw error;
      }

      throw new IntegrationError(
        0,
        error instanceof Error ? error.message : "Unknown error occurred",
      );
    }
  }

  private async getAuthToken(): Promise<string> {
    return "";
  }

  /**
   * Get all integrations for an organization
   */
  async getIntegrations(organizationId: string): Promise<Integration[]> {
    const request: IntegrationsListRequest = {
      organization_id: organizationId,
    };

    const response = await this.request<IntegrationsListResponse>("/list/", {
      method: "POST",
      body: JSON.stringify(request),
    });

    return response.integrations;
  }

  /**
   * Initiate authorization flow for a connector
   */
  async initiateAuthorization(
    organizationId: string,
    userId: string,
    connectorType: ConnectorType,
    redirectUrl?: string,
  ): Promise<AuthorizeResponse> {
    const request: AuthorizeRequest = {
      organization_id: organizationId,
      user_id: userId,
      connector_type: connectorType,
      redirect_url: redirectUrl,
    };

    return this.request<AuthorizeResponse>("/initiate/", {
      method: "POST",
      body: JSON.stringify(request),
    });
  }

  private async requestUnauthenticated<T>(
    endpoint: string,
    options: RequestInit = {},
  ): Promise<T> {
    const url = `${API_BASE_URL}${INTEGRATION_API_PREFIX}${endpoint}`;

    try {
      const response = await fetch(url, {
        ...options,
        headers: {
          "Content-Type": "application/json",
          ...options.headers,
        },
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new IntegrationError(
          response.status,
          errorText || `HTTP ${response.status}: ${response.statusText}`,
        );
      }

      return response.json();
    } catch (error) {
      if (error instanceof IntegrationError) {
        throw error;
      }

      throw new IntegrationError(
        0,
        error instanceof Error ? error.message : "Unknown error occurred",
      );
    }
  }

  /**
   * Handle OAuth callback and complete integration
   * This endpoint doesn't require authentication as it's called by external OAuth providers
   */
  async handleCallback(
    connectorType: ConnectorType,
    callbackData: Record<string, unknown>,
  ): Promise<Integration> {
    const request: CallbackRequest = {
      connector_type: connectorType,
      ...callbackData,
    };

    return this.requestUnauthenticated<Integration>("/authorize/", {
      method: "POST",
      body: JSON.stringify(request),
    });
  }

  /**
   * Get detailed information about a specific integration
   */
  async getIntegrationDetails(
    organizationId: string,
    connectorType: ConnectorType,
  ): Promise<Integration> {
    const request: IntegrationDetailsRequest = {
      organization_id: organizationId,
      connector_type: connectorType,
    };

    return this.request<Integration>("/details/", {
      method: "POST",
      body: JSON.stringify(request),
    });
  }

  /**
   * Get current status of an integration
   */
  async getIntegrationStatus(integrationId: string): Promise<Integration> {
    const request: IntegrationStatusRequest = { integration_id: integrationId };

    return this.request<Integration>("/status/", {
      method: "POST",
      body: JSON.stringify(request),
    });
  }

  /**
   * Test connection for an integration
   */
  async testConnection(integrationId: string): Promise<TestConnectionResponse> {
    const request: TestConnectionRequest = { integration_id: integrationId };

    return this.request<TestConnectionResponse>("/test/", {
      method: "POST",
      body: JSON.stringify(request),
    });
  }

  /**
   * Revoke/disconnect an integration
   */
  async revokeIntegration(
    integrationId: string,
    organizationId: string,
  ): Promise<void> {
    const request: RevokeIntegrationRequest = {
      integration_id: integrationId,
      organization_id: organizationId,
    };

    await this.request<void>("/revoke/", {
      method: "POST",
      body: JSON.stringify(request),
    });
  }

  /**
   * Get activity log for an integration
   */
  async getIntegrationActivity(
    integrationId: string,
  ): Promise<IntegrationActivity[]> {
    return this.request<IntegrationActivity[]>(`/activity/${integrationId}`, {
      method: "GET",
    });
  }

  /**
   * Check if a specific connector type is available
   */
  async checkConnectorAvailability(
    connectorType: ConnectorType,
  ): Promise<boolean> {
    try {
      await this.request<{ available: boolean }>(
        `/availability/${connectorType}`,
        {
          method: "GET",
        },
      );
      return true;
    } catch (error) {
      if (error instanceof IntegrationError && error.statusCode === 404) {
        return false;
      }
      throw error;
    }
  }

  /**
   * Refresh integration credentials (for connectors that support it)
   */
  async refreshCredentials(integrationId: string): Promise<Integration> {
    return this.request<Integration>(`/refresh/${integrationId}`, {
      method: "POST",
    });
  }

  /**
   * Get health check for integration service
   */
  async healthCheck(): Promise<{ status: string; timestamp: string }> {
    return this.request<{ status: string; timestamp: string }>("/health", {
      method: "GET",
    });
  }
}

export const integrationService = new IntegrationService();

export { IntegrationService };

export const isIntegrationError = (error: unknown): error is IntegrationError => {
  return error instanceof IntegrationError;
};

export const getErrorMessage = (error: unknown): string => {
  if (isIntegrationError(error)) {
    switch (error.statusCode) {
      case 401:
        return "You are not authorized to perform this action.";
      case 404:
        return "Integration not found.";
      case 409:
        return "Integration already exists for this service.";
      case 422:
        return "Invalid configuration provided.";
      case 429:
        return "Too many requests. Please try again later.";
      case 500:
        return "Server error. Please try again later.";
      default:
        return error.message || "An unexpected error occurred.";
    }
  }

  return error?.message || "An unexpected error occurred.";
};

export const withErrorHandling = async <T>(
  operation: () => Promise<T>,
  errorContext?: string,
): Promise<T> => {
  try {
    return await operation();
  } catch (error) {
    const sanitizedContext = errorContext
      ? errorContext.replace(/[\r\n\t]/g, " ").substring(0, 100)
      : "";
    console.error(
      "Integration service error%s:",
      sanitizedContext ? ` (${sanitizedContext})` : "",
      error,
    );
    throw error;
  }
};
