import React from 'react';
import { Integration, Connector } from '../../../types/integration';
import { STATUS_CONFIG } from '../../../lib/integration-constants';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Badge } from '../../../components/ui/badge';
import { CheckCircle, XCircle, Clock, AlertCircle } from 'lucide-react';

interface IntegrationStatusProps {
  integration: Integration;
  connector: Connector;
}

export const IntegrationStatus: React.FC<IntegrationStatusProps> = ({
  integration,
  connector
}) => {
  const statusConfig = STATUS_CONFIG[integration.status];

  const getStatusIcon = () => {
    switch (integration.status) {
      case 'connected':
        return <CheckCircle className="h-5 w-5 text-green-600" />;
      case 'error':
        return <XCircle className="h-5 w-5 text-red-600" />;
      case 'syncing':
        return <Clock className="h-5 w-5 text-blue-600 animate-pulse" />;
      default:
        return <AlertCircle className="h-5 w-5 text-gray-600" />;
    }
  };

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffMinutes = Math.floor((now.getTime() - date.getTime()) / (1000 * 60));
    
    if (diffMinutes < 1) return 'Just now';
    if (diffMinutes < 60) return `${diffMinutes} minute${diffMinutes > 1 ? 's' : ''} ago`;
    if (diffMinutes < 1440) {
      const hours = Math.floor(diffMinutes / 60);
      return `${hours} hour${hours > 1 ? 's' : ''} ago`;
    }
    
    return date.toLocaleDateString();
  };

  const renderConnectorSpecificInfo = () => {
    const config = integration.configuration;
    if (!config) return null;

    switch (connector.type) {
      case 'slack':
        return (
          <div className="grid grid-cols-2 gap-4">
            {config.workspaceName && (
              <div>
                <dt className="text-sm font-medium text-gray-500">Workspace</dt>
                <dd className="text-sm text-gray-900">{config.workspaceName}</dd>
              </div>
            )}
            {config.teamId && (
              <div>
                <dt className="text-sm font-medium text-gray-500">Team ID</dt>
                <dd className="text-sm text-gray-900 font-mono">{config.teamId}</dd>
              </div>
            )}
            {config.connectedChannels && config.connectedChannels.length > 0 && (
              <div className="col-span-2">
                <dt className="text-sm font-medium text-gray-500 mb-2">Connected Channels</dt>
                <dd className="flex flex-wrap gap-1">
                  {config.connectedChannels.slice(0, 5).map((channel) => (
                    <Badge key={channel} variant="outline" className="text-xs">
                      #{channel}
                    </Badge>
                  ))}
                  {config.connectedChannels.length > 5 && (
                    <Badge variant="outline" className="text-xs">
                      +{config.connectedChannels.length - 5} more
                    </Badge>
                  )}
                </dd>
              </div>
            )}
          </div>
        );

      case 'github':
        return (
          <div className="grid grid-cols-2 gap-4">
            {config.installationId && (
              <div>
                <dt className="text-sm font-medium text-gray-500">Installation ID</dt>
                <dd className="text-sm text-gray-900 font-mono">{config.installationId}</dd>
              </div>
            )}
            {config.repositoryCount !== undefined && (
              <div>
                <dt className="text-sm font-medium text-gray-500">Repositories</dt>
                <dd className="text-sm text-gray-900">{config.repositoryCount}</dd>
              </div>
            )}
            {config.connectedRepos && config.connectedRepos.length > 0 && (
              <div className="col-span-2">
                <dt className="text-sm font-medium text-gray-500 mb-2">Connected Repositories</dt>
                <dd className="flex flex-wrap gap-1">
                  {config.connectedRepos.slice(0, 3).map((repo) => (
                    <Badge key={repo} variant="outline" className="text-xs">
                      {repo}
                    </Badge>
                  ))}
                  {config.connectedRepos.length > 3 && (
                    <Badge variant="outline" className="text-xs">
                      +{config.connectedRepos.length - 3} more
                    </Badge>
                  )}
                </dd>
              </div>
            )}
            {config.webhookUrl && (
              <div className="col-span-2">
                <dt className="text-sm font-medium text-gray-500">Webhook URL</dt>
                <dd className="text-sm text-gray-900 font-mono break-all">{config.webhookUrl}</dd>
              </div>
            )}
          </div>
        );

      case 'aws':
        return (
          <div className="grid grid-cols-2 gap-4">
            {config.region && (
              <div>
                <dt className="text-sm font-medium text-gray-500">Region</dt>
                <dd className="text-sm text-gray-900">{config.region}</dd>
              </div>
            )}
            {config.accountId && (
              <div>
                <dt className="text-sm font-medium text-gray-500">Account ID</dt>
                <dd className="text-sm text-gray-900 font-mono">{config.accountId}</dd>
              </div>
            )}
          </div>
        );

      case 'gcp':
        return (
          <div className="grid grid-cols-2 gap-4">
            {config.projectId && (
              <div>
                <dt className="text-sm font-medium text-gray-500">Project ID</dt>
                <dd className="text-sm text-gray-900 font-mono">{config.projectId}</dd>
              </div>
            )}
          </div>
        );

      default:
        return null;
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center space-x-2">
          {getStatusIcon()}
          <span>Status</span>
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="flex items-center space-x-4">
          <Badge 
            variant="secondary"
            className={`${statusConfig.color} ${statusConfig.bgColor} ${statusConfig.borderColor} border px-3 py-1`}
          >
            <span className="mr-1">{statusConfig.icon}</span>
            {integration.status.charAt(0).toUpperCase() + integration.status.slice(1)}
          </Badge>
          
        </div>

        {(() => {
          const connectorSpecificInfo = renderConnectorSpecificInfo();
          return connectorSpecificInfo && (
            <div>
              <h4 className="text-sm font-medium text-gray-900 mb-3">Configuration Details</h4>
              <dl className="space-y-3">
                {connectorSpecificInfo}
              </dl>
            </div>
          );
        })()}

        {integration.configuration?.permissions && integration.configuration.permissions.length > 0 && (
          <div>
            <h4 className="text-sm font-medium text-gray-900 mb-3">Permissions</h4>
            <div className="flex flex-wrap gap-1">
              {integration.configuration.permissions.map((permission) => (
                <Badge key={permission} variant="outline" className="text-xs">
                  {permission.replace(/[_:]/g, ' ')}
                </Badge>
              ))}
            </div>
          </div>
        )}

        <div className="border-t pt-4">
          <h4 className="text-sm font-medium text-gray-900 mb-3">Timeline</h4>
          <dl className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <dt className="text-gray-500">Connected</dt>
              <dd className="text-gray-900">{formatTimestamp(integration.createdAt)}</dd>
            </div>
            <div>
              <dt className="text-gray-500">Last Updated</dt>
              <dd className="text-gray-900">{formatTimestamp(integration.updatedAt)}</dd>
            </div>
          </dl>
        </div>

        {integration.status === 'error' && integration.metadata?.errorMessage && (
          <div className="border-t pt-4">
            <h4 className="text-sm font-medium text-red-800 mb-2">Error Details</h4>
            <div className="bg-red-50 border border-red-200 rounded-lg p-3">
              <p className="text-sm text-red-700">{integration.metadata.errorMessage}</p>
              {integration.metadata.errorCode && (
                <p className="text-xs text-red-600 mt-1 font-mono">
                  Error Code: {integration.metadata.errorCode}
                </p>
              )}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
};