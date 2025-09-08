import { useAuth } from "@clerk/clerk-react";
import { useCallback } from "react";

const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL || "http://localhost:8080";

export interface Organization {
  id: string;
  clerk_org_id: string;
  name: string;
  slug: string;
  created_at: string;
  metadata?: OrganizationMetadata;
}

export interface OrganizationMetadata {
  company_size: string;
  team_size: string;
  use_cases: string[];
  observability_stack: string[];
  completed_at: string;
}

export interface OrganizationMetadataRequest {
  organization_id: string;
  company_size: string;
  team_size: string;
  use_cases: string[];
  observability_stack: string[];
}

export interface UserProfile {
  id: string;
  name: string;
  slug: string;
  created_at: string;
  organization_id: string;
  user_id: string;
  metadata: {
    company_size: string;
    team_size: string;
    use_cases: string[];
    observability_stack: string[];
    completed_at: string;
  };
}

// Hook to use API client with Clerk auth
export const useApiClient = () => {
  const { getToken } = useAuth();

  const makeAuthenticatedRequest = useCallback(
    async <T>(endpoint: string, options: RequestInit = {}): Promise<T> => {
      const token = await getToken();

      const response = await fetch(`${API_BASE_URL}${endpoint}`, {
        ...options,
        headers: {
          "Content-Type": "application/json",
          Authorization: token ? `Bearer ${token}` : "",
          ...options.headers,
        },
      });

      if (!response.ok) {
        const error = await response.text();
        throw new Error(`API Error: ${response.status} - ${error}`);
      }

      return response.json();
    },
    [getToken],
  );

  const getOrganization = useCallback(
    async (clerkOrgId: string): Promise<Organization> => {
      return makeAuthenticatedRequest<Organization>("/identity/organization/", {
        method: "POST",
        body: JSON.stringify({ clerk_org_id: clerkOrgId }),
      });
    },
    [makeAuthenticatedRequest],
  );

  const setOrganizationMetadata = useCallback(
    async (metadata: OrganizationMetadataRequest): Promise<void> => {
      return makeAuthenticatedRequest<void>(
        "/identity/organization/set-metadata/",
        {
          method: "POST",
          body: JSON.stringify(metadata),
        },
      );
    },
    [makeAuthenticatedRequest],
  );

  const getMe = useCallback(
    async (clerkUserId: string, clerkOrgId: string): Promise<UserProfile> => {
      return makeAuthenticatedRequest<UserProfile>("/identity/me/", {
        method: "POST",
        body: JSON.stringify({
          clerk_user_id: clerkUserId,
          clerk_org_id: clerkOrgId,
        }),
      });
    },
    [makeAuthenticatedRequest],
  );

  const apiPost = useCallback(
    async <T>(endpoint: string, data?: unknown): Promise<T> => {
      return makeAuthenticatedRequest<T>(endpoint, {
        method: "POST",
        body: data ? JSON.stringify(data) : undefined,
      });
    },
    [makeAuthenticatedRequest],
  );

  return {
    getOrganization,
    setOrganizationMetadata,
    getMe,
    apiPost,
  };
};
