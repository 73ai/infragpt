const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL || "http://localhost:8080";
const DEVICE_API_PREFIX = "/device";

export interface AuthorizeDeviceRequest {
  user_code: string;
  organization_id: string;
  user_id: string;
}

export interface AuthorizeDeviceResponse {
  success: boolean;
  error?: string;
}

class DeviceServiceError extends Error {
  constructor(
    public statusCode: number,
    message: string,
  ) {
    super(message);
    this.name = "DeviceServiceError";
  }
}

class DeviceService {
  private async request<T>(
    endpoint: string,
    options: RequestInit = {},
    authToken?: string,
  ): Promise<T> {
    const url = `${API_BASE_URL}${DEVICE_API_PREFIX}${endpoint}`;

    try {
      const headers: Record<string, string> = {
        "Content-Type": "application/json",
        ...((options.headers as Record<string, string>) || {}),
      };

      if (authToken) {
        headers.Authorization = `Bearer ${authToken}`;
      }

      const response = await fetch(url, {
        ...options,
        headers,
      });

      if (!response.ok) {
        const errorText = await response.text();
        let errorMessage = `HTTP ${response.status}: ${response.statusText}`;
        try {
          const errorJson = JSON.parse(errorText);
          if (errorJson.error) {
            errorMessage = errorJson.error;
          } else if (errorJson.message) {
            errorMessage = errorJson.message;
          }
        } catch {
          if (errorText) {
            errorMessage = errorText;
          }
        }
        throw new DeviceServiceError(response.status, errorMessage);
      }

      return response.json();
    } catch (error) {
      if (error instanceof DeviceServiceError) {
        throw error;
      }

      throw new DeviceServiceError(
        0,
        error instanceof Error ? error.message : "Unknown error occurred",
      );
    }
  }

  async authorizeDevice(
    userCode: string,
    organizationId: string,
    userId: string,
    authToken: string,
  ): Promise<AuthorizeDeviceResponse> {
    const request: AuthorizeDeviceRequest = {
      user_code: userCode,
      organization_id: organizationId,
      user_id: userId,
    };

    return this.request<AuthorizeDeviceResponse>(
      "/auth/authorize",
      {
        method: "POST",
        body: JSON.stringify(request),
      },
      authToken,
    );
  }
}

export const deviceService = new DeviceService();

export { DeviceService, DeviceServiceError };

export const isDeviceServiceError = (
  error: unknown,
): error is DeviceServiceError => {
  return error instanceof DeviceServiceError;
};

export const getDeviceErrorMessage = (error: unknown): string => {
  if (isDeviceServiceError(error)) {
    switch (error.statusCode) {
      case 404:
        return "Invalid code. Please check and try again.";
      case 410:
        return "Code has expired. Please generate a new code from the CLI.";
      case 409:
        return "This code has already been used.";
      default:
        return error.message || "An unexpected error occurred.";
    }
  }

  return (error as Error)?.message || "An unexpected error occurred.";
};
