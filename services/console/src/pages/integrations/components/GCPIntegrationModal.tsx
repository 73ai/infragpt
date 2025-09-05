// GCP Integration Modal - Service Account JSON Configuration

import React, { useState, useCallback } from 'react';
import { observer } from 'mobx-react-lite';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '../../../components/ui/dialog';
import { Button } from '../../../components/ui/button';
import { Alert, AlertDescription, AlertTitle } from '../../../components/ui/alert';
import { Badge } from '../../../components/ui/badge';
import { Textarea } from '../../../components/ui/textarea';
import { Loader2, AlertCircle, CheckCircle2, ExternalLink, Shield } from 'lucide-react';
import { integrationStore } from '../../../stores/IntegrationStore';
import { userStore } from '../../../stores/UserStore';
import { useApiClient } from '../../../lib/api';

interface GCPIntegrationModalProps {
  isOpen: boolean;
  onClose: () => void;
}

interface ServiceAccountInfo {
  type: string;
  project_id: string;
  client_email: string;
  private_key: string;
}

export const GCPIntegrationModal: React.FC<GCPIntegrationModalProps> = observer(({ 
  isOpen, 
  onClose 
}) => {
  const [serviceAccountJSON, setServiceAccountJSON] = useState('');
  const [isValidating, setIsValidating] = useState(false);
  const [isConnecting, setIsConnecting] = useState(false);
  const [serviceAccountInfo, setServiceAccountInfo] = useState<ServiceAccountInfo | null>(null);
  const [validationResult, setValidationResult] = useState<any>(null);
  const [error, setError] = useState<string | null>(null);
  const [jsonError, setJsonError] = useState<string | null>(null);
  const { apiPost } = useApiClient();

  // Format JSON for display (currently unused but kept for future enhancement)
  // const formatJSON = (json: string): string => {
  //   try {
  //     const parsed = JSON.parse(json);
  //     return JSON.stringify(parsed, null, 2);
  //   } catch {
  //     return json;
  //   }
  // };

  // Handle JSON input change
  const handleJSONChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const value = e.target.value;
    setServiceAccountJSON(value);
    setError(null);
    setJsonError(null);
    setValidationResult(null);
    
    // Parse and validate JSON structure on change
    if (value.trim()) {
      try {
        const parsed = JSON.parse(value) as ServiceAccountInfo;
        
        // Basic validation
        const errors: string[] = [];
        if (parsed.type !== 'service_account') {
          errors.push(`Invalid type: expected 'service_account', got '${parsed.type}'`);
        }
        if (!parsed.project_id) {
          errors.push('project_id is required');
        }
        if (!parsed.client_email) {
          errors.push('client_email is required');
        }
        if (!parsed.private_key) {
          errors.push('private_key is required');
        }

        if (errors.length > 0) {
          setJsonError(errors.join(', '));
          setServiceAccountInfo(null);
        } else {
          setServiceAccountInfo(parsed);
          setJsonError(null);
        }
      } catch (e) {
        setJsonError('Invalid JSON format');
        setServiceAccountInfo(null);
      }
    } else {
      setServiceAccountInfo(null);
    }
  };

  // Validate credentials using the generic API
  const validateCredentials = useCallback(async () => {
    if (!serviceAccountInfo || jsonError) {
      setError('Please provide valid service account JSON');
      return;
    }

    setIsValidating(true);
    setError(null);

    try {
      const response = await apiPost('/integrations/validate', {
        connector_type: 'gcp',
        credentials: {
          service_account_json: serviceAccountJSON
        }
      });

      setValidationResult(response);
      
      if (!response.valid) {
        const errorMessage = response.errors?.join(', ') || 'Validation failed';
        setError(errorMessage);
      }
    } catch (err: any) {
      setError(err.message || 'Failed to validate credentials');
    } finally {
      setIsValidating(false);
    }
  }, [serviceAccountInfo, jsonError, serviceAccountJSON, apiPost]);

  // Connect the integration
  const handleConnect = useCallback(async () => {
    if (!validationResult?.valid) {
      setError('Please validate your credentials first');
      return;
    }

    if (!userStore.organizationId || !userStore.userId) {
      setError('Organization information not available');
      return;
    }

    setIsConnecting(true);
    setError(null);

    try {
      // Use the standard integration flow
      await integrationStore.handleCallback('gcp', {
        code: serviceAccountJSON,
        state: `${userStore.organizationId}:${userStore.userId}`
      });

      // Reload integrations to show the new connection
      await integrationStore.loadIntegrations(userStore.organizationId);

      // Close modal on success
      onClose();
    } catch (err: any) {
      setError(err.message || 'Failed to connect GCP integration');
    } finally {
      setIsConnecting(false);
    }
  }, [validationResult, serviceAccountJSON, onClose]);


  // Reset modal state when closed
  const handleClose = () => {
    setServiceAccountJSON('');
    setServiceAccountInfo(null);
    setValidationResult(null);
    setError(null);
    setJsonError(null);
    setIsValidating(false);
    setIsConnecting(false);
    onClose();
  };

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="max-w-3xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <img 
              src="/icons/gcp.svg" 
              alt="Google Cloud Platform" 
              className="w-6 h-6"
            />
            Connect Google Cloud Platform
          </DialogTitle>
          <DialogDescription>
            Configure GCP integration using a service account with Viewer role permissions.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {/* Requirements Alert */}
          <Alert>
            <Shield className="h-4 w-4" />
            <AlertTitle>Service Account Requirements</AlertTitle>
            <AlertDescription>
              <ul className="mt-2 space-y-1 text-sm">
                <li>• Service account must have the <strong>Viewer role</strong> on your GCP project</li>
                <li>• You can grant this role in the{' '}
                  <a 
                    href="https://console.cloud.google.com/iam-admin/roles/details/roles%3Cviewer"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-600 hover:underline inline-flex items-center gap-1"
                  >
                    IAM & Admin Console
                    <ExternalLink className="h-3 w-3" />
                  </a>
                </li>
                <li>• Download the service account key as JSON from the GCP Console</li>
              </ul>
            </AlertDescription>
          </Alert>

          {/* JSON Editor */}
          <div className="space-y-2">
            <label className="text-sm font-medium">Service Account JSON</label>
            <Textarea
              value={serviceAccountJSON}
              onChange={handleJSONChange}
              placeholder={`Paste your service account JSON here...

Example:
{
  "type": "service_account",
  "project_id": "your-project-id",
  "private_key_id": "...",
  "private_key": "-----BEGIN PRIVATE KEY-----\\n...\\n-----END PRIVATE KEY-----\\n",
  "client_email": "your-service-account@your-project.iam.gserviceaccount.com",
  ...
}`}
              className="font-mono text-xs h-64 resize-none"
              spellCheck={false}
            />
          </div>

          {/* Validation Result */}
          {validationResult && validationResult.valid && (
            <Alert className="border-green-200 bg-green-50">
              <CheckCircle2 className="h-4 w-4 text-green-600" />
              <AlertTitle className="text-green-800">Credentials Validated</AlertTitle>
              <AlertDescription>
                <div className="mt-2 space-y-1 text-sm text-green-700">
                  {validationResult.details && (
                    <>
                      <div className="flex items-center gap-2">
                        <span className="font-medium">Project ID:</span>
                        <Badge variant="secondary">{validationResult.details.project_id}</Badge>
                      </div>
                      <div className="flex items-center gap-2">
                        <span className="font-medium">Service Account:</span>
                        <span className="font-mono text-xs">{validationResult.details.client_email}</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <span className="font-medium">Viewer Role:</span>
                        {validationResult.details.has_viewer ? (
                          <Badge className="bg-green-600">Granted</Badge>
                        ) : (
                          <Badge variant="destructive">Not Granted</Badge>
                        )}
                      </div>
                    </>
                  )}
                </div>
              </AlertDescription>
            </Alert>
          )}

          {/* Error Display */}
          {error && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertTitle>Validation Error</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {/* Security Notice */}
          <div className="rounded-lg bg-blue-50 p-3 text-sm">
            <div className="flex items-start space-x-2">
              <Shield className="h-4 w-4 text-blue-600 mt-0.5" />
              <div>
                <p className="text-blue-800 font-medium">Security Information</p>
                <p className="text-blue-700 mt-1">
                  Your service account credentials will be encrypted using AES-256-GCM and stored securely. 
                  InfraGPT only accesses the permissions explicitly granted to the service account.
                </p>
              </div>
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button 
            variant="outline" 
            onClick={handleClose}
            disabled={isValidating || isConnecting}
          >
            Cancel
          </Button>
          
          {!validationResult?.valid ? (
            <Button 
              onClick={validateCredentials}
              disabled={isValidating || !serviceAccountInfo || jsonError}
            >
              {isValidating ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Validating...
                </>
              ) : (
                'Validate Credentials'
              )}
            </Button>
          ) : (
            <Button 
              onClick={handleConnect}
              disabled={isConnecting || (validationResult.details && !validationResult.details.has_viewer)}
              className="bg-green-600 hover:bg-green-700"
            >
              {isConnecting ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Connecting...
                </>
              ) : (
                'Connect GCP'
              )}
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
});

export default GCPIntegrationModal;