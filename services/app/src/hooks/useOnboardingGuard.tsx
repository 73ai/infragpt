import { useEffect, useState } from 'react';
import { useOrganization, useAuth } from '@clerk/clerk-react';
import { useApiClient } from '@/lib/api';

export interface OnboardingStatus {
  isLoading: boolean;
  isComplete: boolean;
  hasOrganization: boolean;
  error: string | null;
}

export const useOnboardingGuard = (): OnboardingStatus => {
  const [status, setStatus] = useState<OnboardingStatus>({
    isLoading: true,
    isComplete: false,
    hasOrganization: false,
    error: null,
  });

  const { organization, isLoaded: orgLoaded } = useOrganization();
  const { isSignedIn, isLoaded: authLoaded } = useAuth();
  const { getOrganization } = useApiClient();

  useEffect(() => {
    const checkOnboardingStatus = async () => {
      // Wait for auth to load first
      if (!authLoaded) {
        return;
      }

      // If not signed in, mark as not complete
      if (!isSignedIn) {
        setStatus({
          isLoading: false,
          isComplete: false,
          hasOrganization: false,
          error: null,
        });
        return;
      }

      // Wait for organization to load
      if (!orgLoaded) {
        return;
      }

      // If no organization, user needs to create one
      if (!organization?.id) {
        setStatus({
          isLoading: false,
          isComplete: false,
          hasOrganization: false,
          error: null,
        });
        return;
      }

      // Check if organization metadata is complete
      try {
        const orgData = await getOrganization(organization.id);
        
        // Check if metadata exists and has required fields
        const isComplete = Boolean(
          orgData.metadata?.company_size &&
          orgData.metadata?.team_size &&
          orgData.metadata?.use_cases?.length > 0 &&
          orgData.metadata?.observability_stack?.length > 0
        );

        setStatus({
          isLoading: false,
          isComplete,
          hasOrganization: true,
          error: null,
        });
      } catch (error) {
        // If organization doesn't exist in our backend yet, onboarding is incomplete
        setStatus({
          isLoading: false,
          isComplete: false,
          hasOrganization: true, // Org exists in Clerk, but not in our backend
          error: null,
        });
      }
    };

    checkOnboardingStatus();
  }, [organization?.id, isSignedIn, authLoaded, orgLoaded, getOrganization]);

  return status;
};