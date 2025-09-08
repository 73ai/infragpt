// GitHub App Setup Handler Page

import { useEffect, useState } from "react";
import { observer } from "mobx-react-lite";
import { useParams, useNavigate, useSearchParams } from "react-router-dom";
import { integrationStore } from "../../stores/IntegrationStore";
import { getConnectorByType } from "../../lib/integration-constants";
import { Button } from "../../components/ui/button";
import { Card } from "../../components/ui/card";
import { CheckCircle, XCircle, Loader2 } from "lucide-react";

type SetupState = "processing" | "success" | "error";

const IntegrationSetupPage = observer(() => {
  const { connectorType } = useParams<{ connectorType: string }>();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const [state, setState] = useState<SetupState>("processing");
  const [message, setMessage] = useState("");

  const connector = connectorType
    ? getConnectorByType(connectorType as any)
    : null;

  useEffect(() => {
    const handleSetup = async () => {
      if (!connectorType || !connector) {
        setState("error");
        setMessage("Invalid integration type");
        return;
      }

      // Only handle GitHub for now
      if (connectorType !== "github") {
        setState("error");
        setMessage("Setup page only supports GitHub integration");
        return;
      }

      try {
        // Extract GitHub setup parameters
        const setupData = extractGitHubSetupData(searchParams);
        const setupAction = searchParams.get("setup_action");

        if (!setupData) {
          setState("error");
          setMessage("Missing GitHub setup data");
          return;
        }

        // Handle different setup actions
        if (setupAction === "install") {
          // For new installations, use the existing callback flow
          await integrationStore.handleCallback(
            connectorType as any,
            setupData,
          );
          setState("success");
          setMessage(
            `${connector.name} has been successfully installed and connected to your organization.`,
          );
        } else if (setupAction === "update") {
          // For updates, we need a sync API call
          // TODO: Implement sync API when backend is ready
          // For now, treat as successful
          setState("success");
          setMessage(
            `${connector.name} configuration has been updated successfully.`,
          );
        } else {
          setState("error");
          setMessage(`Unknown setup action: ${setupAction}`);
        }
      } catch (error) {
        setState("error");
        setMessage(
          error instanceof Error
            ? error.message
            : `Failed to complete ${connector?.name} setup. Please try again.`,
        );
        integrationStore.handleError(error, "handling setup");
      }
    };

    // Add a small delay to show the processing state
    const timer = setTimeout(handleSetup, 1000);
    return () => clearTimeout(timer);
  }, [connectorType, connector, searchParams]);

  const extractGitHubSetupData = (
    searchParams: URLSearchParams,
  ): Record<string, any> | null => {
    const installationId = searchParams.get("installation_id");
    const setupAction = searchParams.get("setup_action");
    const state = searchParams.get("state");

    if (!installationId) {
      return null;
    }

    return {
      installation_id: installationId,
      state: state,
    };
  };

  const handleViewDetails = () => {
    navigate(`/integrations/${connectorType}`);
  };

  const handleBackToIntegrations = () => {
    navigate("/integrations");
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 p-4">
      <Card className="w-full max-w-md p-8 text-center">
        {state === "processing" && (
          <>
            <Loader2 className="h-12 w-12 text-blue-600 animate-spin mx-auto mb-4" />
            <h1 className="text-xl font-semibold mb-2">
              Setting up Integration
            </h1>
            <p className="text-muted-foreground mb-6">
              Please wait while we configure your {connector?.name}{" "}
              integration...
            </p>
          </>
        )}

        {state === "success" && (
          <>
            <CheckCircle className="h-12 w-12 text-green-600 mx-auto mb-4" />
            <h1 className="text-xl font-semibold text-green-800 mb-2">
              Setup Successful!
            </h1>
            <p className="text-muted-foreground mb-6">{message}</p>
            <div className="space-y-3">
              <Button onClick={handleViewDetails} className="w-full">
                View Integration Details
              </Button>
              <Button
                variant="outline"
                onClick={handleBackToIntegrations}
                className="w-full"
              >
                Back to Integrations
              </Button>
            </div>
          </>
        )}

        {state === "error" && (
          <>
            <XCircle className="h-12 w-12 text-red-600 mx-auto mb-4" />
            <h1 className="text-xl font-semibold text-red-800 mb-2">
              Setup Failed
            </h1>
            <p className="text-muted-foreground mb-6">{message}</p>

            {/* Additional error context */}
            {integrationStore.error && (
              <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4">
                <p className="text-sm text-red-700">{integrationStore.error}</p>
              </div>
            )}

            <div className="space-y-3">
              <Button onClick={handleBackToIntegrations} className="w-full">
                Try Again
              </Button>
              <Button
                variant="outline"
                onClick={handleBackToIntegrations}
                className="w-full"
              >
                Back to Integrations
              </Button>
            </div>
          </>
        )}

        {/* Debug information (only in development) */}
        {import.meta.env.DEV && (
          <details className="mt-6 text-left">
            <summary className="text-xs text-muted-foreground cursor-pointer">
              Debug Info
            </summary>
            <div className="mt-2 text-xs bg-gray-100 p-2 rounded">
              <pre>
                {JSON.stringify(
                  {
                    connectorType,
                    params: Object.fromEntries(searchParams),
                    state,
                    message,
                  },
                  null,
                  2,
                )}
              </pre>
            </div>
          </details>
        )}
      </Card>
    </div>
  );
});

export default IntegrationSetupPage;
