// OAuth Callback Handler Page

import { useEffect, useState } from 'react';
import { observer } from 'mobx-react-lite';
import { useParams, useNavigate, useSearchParams } from 'react-router-dom';
import { integrationStore } from '../../stores/IntegrationStore';
import { getConnectorByType } from '../../lib/integration-constants';
import { Button } from '../../components/ui/button';
import { Card } from '../../components/ui/card';
import { CheckCircle, XCircle, Loader2 } from 'lucide-react';

type CallbackState = 'processing' | 'success' | 'error';

const IntegrationCallbackPage = observer(() => {
  const { connectorType } = useParams<{ connectorType: string }>();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  
  const [state, setState] = useState<CallbackState>('processing');
  const [message, setMessage] = useState('');

  const connector = connectorType ? getConnectorByType(connectorType as any) : null;

  useEffect(() => {
    const handleCallback = async () => {
      if (!connectorType || !connector) {
        setState('error');
        setMessage('Invalid integration type');
        return;
      }

      try {
        // Extract callback parameters based on connector type
        const callbackData = extractCallbackData(connectorType, searchParams);
        
        if (!callbackData || Object.keys(callbackData).length === 0) {
          setState('error');
          setMessage('Missing authorization data');
          return;
        }

        // Handle the callback
        await integrationStore.handleCallback(
          connectorType as any,
          callbackData
        );

        setState('success');
        setMessage(`${connector.name} has been successfully connected to your organization.`);

      } catch (error) {
        setState('error');
        setMessage(
          error instanceof Error 
            ? error.message 
            : `Failed to complete ${connector?.name} integration. Please try again.`
        );
        integrationStore.handleError(error, 'handling callback');
      }
    };

    // Add a small delay to show the processing state
    const timer = setTimeout(handleCallback, 1000);
    return () => clearTimeout(timer);
  }, [connectorType, connector, searchParams]);

  const extractCallbackData = (
    connectorType: string, 
    searchParams: URLSearchParams
  ): Record<string, any> | null => {
    switch (connectorType) {
      case 'slack':
        // Slack OAuth2 flow parameters
        const code = searchParams.get('code');
        const state = searchParams.get('state');
        const error = searchParams.get('error');
        
        if (error) {
          throw new Error(`Slack authorization failed: ${error}`);
        }
        
        if (!code) {
          return null;
        }
        
        return { code, state };

      case 'github':
        // GitHub App installation flow parameters
        const installationId = searchParams.get('installation_id');
        const setupAction = searchParams.get('setup_action');
        const githubCode = searchParams.get('code');
        const githubState = searchParams.get('state');
        
        if (!installationId && !githubCode) {
          return null;
        }
        
        return { 
          installation_id: installationId ? parseInt(installationId) : undefined,
          setup_action: setupAction,
          code: githubCode,
          state: githubState
        };

      default:
        // Generic OAuth2 flow
        const genericCode = searchParams.get('code');
        const genericState = searchParams.get('state');
        
        if (!genericCode) {
          return null;
        }
        
        return { code: genericCode, state: genericState };
    }
  };

  const handleViewDetails = () => {
    navigate(`/integrations/${connectorType}`);
  };

  const handleBackToIntegrations = () => {
    navigate('/integrations');
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 p-4">
      <Card className="w-full max-w-md p-8 text-center">
        {state === 'processing' && (
          <>
            <Loader2 className="h-12 w-12 text-blue-600 animate-spin mx-auto mb-4" />
            <h1 className="text-xl font-semibold mb-2">Setting up Integration</h1>
            <p className="text-muted-foreground mb-6">
              Please wait while we complete the {connector?.name} integration...
            </p>
          </>
        )}

        {state === 'success' && (
          <>
            <CheckCircle className="h-12 w-12 text-green-600 mx-auto mb-4" />
            <h1 className="text-xl font-semibold text-green-800 mb-2">
              Integration Successful!
            </h1>
            <p className="text-muted-foreground mb-6">{message}</p>
            <div className="space-y-3">
              <Button 
                onClick={handleViewDetails}
                className="w-full"
              >
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

        {state === 'error' && (
          <>
            <XCircle className="h-12 w-12 text-red-600 mx-auto mb-4" />
            <h1 className="text-xl font-semibold text-red-800 mb-2">
              Integration Failed
            </h1>
            <p className="text-muted-foreground mb-6">{message}</p>
            
            {/* Additional error context */}
            {integrationStore.error && (
              <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4">
                <p className="text-sm text-red-700">
                  {integrationStore.error}
                </p>
              </div>
            )}
            
            <div className="space-y-3">
              <Button 
                onClick={handleBackToIntegrations}
                className="w-full"
              >
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
              <pre>{JSON.stringify({
                connectorType,
                params: Object.fromEntries(searchParams),
                state,
                message
              }, null, 2)}</pre>
            </div>
          </details>
        )}
      </Card>
    </div>
  );
});

export default IntegrationCallbackPage;