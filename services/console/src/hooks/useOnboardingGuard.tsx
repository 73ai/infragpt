import { useEffect, useState } from "react";
import { useOrganization, useAuth, useUser } from "@clerk/clerk-react";
import { useApiClient } from "@/lib/api";
import { userStore } from "@/stores/UserStore";

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
  const { user } = useUser();
  const { getMe } = useApiClient();

  const clerkUserId = user?.id;
  const clerkOrgId = organization?.id;

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

      // Load user profile and check onboarding completion
      try {
        // Load user profile if not already loaded
        if (
          !userStore.userProfile &&
          !userStore.loading &&
          clerkUserId &&
          clerkOrgId
        ) {
          await userStore.loadUserProfile(getMe, clerkUserId, clerkOrgId);
        }

        // Check if metadata exists and has required fields
        const isComplete = Boolean(
          userStore.userProfile?.metadata?.company_size &&
            userStore.userProfile?.metadata?.team_size &&
            userStore.userProfile?.metadata?.use_cases?.length > 0 &&
            userStore.userProfile?.metadata?.observability_stack?.length > 0,
        );

        setStatus({
          isLoading: false,
          isComplete,
          hasOrganization: true,
          error: null,
        });
      } catch {
        // If user profile doesn't exist in our backend yet, onboarding is incomplete
        setStatus({
          isLoading: false,
          isComplete: false,
          hasOrganization: true, // Org exists in Clerk, but user profile not in our backend
          error: null,
        });
      }
    };

    checkOnboardingStatus();
  }, [
    organization?.id,
    isSignedIn,
    authLoaded,
    orgLoaded,
    getMe,
    clerkUserId,
    clerkOrgId,
  ]);

  return status;
};
