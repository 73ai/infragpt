// ConnectorCard Component - Individual Integration Card

import React from 'react';
import { Connector, Integration, ConnectorType } from '../../../types/integration';
import { BUTTON_TEXT, STATUS_CONFIG } from '../../../lib/integration-constants';
import { Button } from '../../../components/ui/button';
import { Card, CardContent } from '../../../components/ui/card';
import { Badge } from '../../../components/ui/badge';
import { Loader2 } from 'lucide-react';

interface ConnectorCardProps {
  connector: Connector;
  integration?: Integration;
  onAction: (connectorType: ConnectorType, action: 'connect' | 'details') => void;
  loading?: boolean;
}

export const ConnectorCard: React.FC<ConnectorCardProps> = ({
  connector,
  integration,
  onAction,
  loading = false
}) => {
  const getButtonState = (): 'connect' | 'show_details' | 'coming_soon' => {
    if (!connector.isImplemented) {
      return 'coming_soon';
    }
    
    if (integration?.status === 'active' || integration?.status === 'connected') {
      return 'show_details';
    }
    
    return 'connect';
  };

  const buttonState = getButtonState();
  const isConnected = integration?.status === 'active' || integration?.status === 'connected';
  const hasError = integration?.status === 'error';

  const handleClick = () => {
    if (buttonState === 'coming_soon' || loading) return;
    
    const action = buttonState === 'show_details' ? 'details' : 'connect';
    onAction(connector.type, action);
  };

  const getStatusBadge = () => {
    if (!integration) return null;
    
    const config = STATUS_CONFIG[integration.status];
    return (
      <Badge 
        variant="secondary"
        className={`${config.color} ${config.bgColor} ${config.borderColor} border`}
      >
        <span className="mr-1">{config.icon}</span>
        {config.text}
      </Badge>
    );
  };

  const getCreatedText = () => {
    if (!integration?.createdAt) return null;
    
    const created = new Date(integration.createdAt);
    const now = new Date();
    const diffMinutes = Math.floor((now.getTime() - created.getTime()) / (1000 * 60));
    
    if (diffMinutes < 60) {
      return `Added ${diffMinutes} min ago`;
    } else if (diffMinutes < 1440) {
      const hours = Math.floor(diffMinutes / 60);
      return `Added ${hours} hour${hours > 1 ? 's' : ''} ago`;
    } else {
      const days = Math.floor(diffMinutes / 1440);
      return `Added ${days} day${days > 1 ? 's' : ''} ago`;
    }
  };

  return (
    <Card className={`transition-all duration-200 hover:shadow-md ${
      hasError ? 'border-red-200 bg-red-50' : 
      isConnected ? 'border-green-200 bg-green-50' : 
      'hover:border-gray-300'
    }`}>
      <CardContent className="p-4">
        <div className="flex items-center space-x-4">
          {/* Logo */}
          <div className="h-10 w-10 rounded-lg bg-white shadow-sm flex items-center justify-center flex-shrink-0">
            <img 
              src={connector.logo} 
              alt={`${connector.name} logo`}
              className="h-6 w-6"
              onError={(e) => {
                // Fallback for missing logos
                const target = e.target as HTMLImageElement;
                target.style.display = 'none';
                target.nextElementSibling?.classList.remove('hidden');
              }}
            />
            <div className="hidden h-6 w-6 bg-gray-200 rounded text-xs flex items-center justify-center font-semibold text-gray-600">
              {connector.name.substring(0, 2).toUpperCase()}
            </div>
          </div>

          {/* Content - Name and Description */}
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 mb-1">
              <h3 className="font-semibold text-gray-900">{connector.name}</h3>
              {getStatusBadge()}
            </div>
            <p className="text-sm text-muted-foreground leading-relaxed">
              {connector.description}
            </p>

            {/* Capabilities */}
            <div className="flex flex-wrap gap-1 mt-2">
              {connector.capabilities.slice(0, 3).map((capability) => (
                <Badge 
                  key={capability} 
                  variant="outline"
                  className="text-xs px-2 py-0"
                >
                  {capability.replace('_', ' ')}
                </Badge>
              ))}
              {connector.capabilities.length > 3 && (
                <Badge variant="outline" className="text-xs px-2 py-0">
                  +{connector.capabilities.length - 3} more
                </Badge>
              )}
            </div>

            {/* Integration Details - Condensed for horizontal layout */}
            {isConnected && integration && (
              <div className="flex flex-wrap gap-4 mt-2 text-xs text-muted-foreground">
                {integration.configuration?.workspaceName && (
                  <span>
                    <span className="font-medium">Workspace:</span> {integration.configuration.workspaceName}
                  </span>
                )}
                {integration.configuration?.repositoryCount && (
                  <span>
                    <span className="font-medium">Repos:</span> {integration.configuration.repositoryCount}
                  </span>
                )}
                {integration.configuration?.connectedChannels && (
                  <span>
                    <span className="font-medium">Channels:</span> {integration.configuration.connectedChannels.length}
                  </span>
                )}
                {getCreatedText() && (
                  <span>
                    {getCreatedText()}
                  </span>
                )}
              </div>
            )}

            {/* Error Message */}
            {hasError && (
              <div className="text-xs text-red-600 bg-red-100 rounded p-2 mt-2">
                Integration error - Click to view details
              </div>
            )}
          </div>

          {/* Action Button */}
          <div className="flex-shrink-0">
            <Button
              onClick={handleClick}
              disabled={buttonState === 'coming_soon' || loading}
              variant={
                buttonState === 'show_details' ? 'outline' :
                buttonState === 'coming_soon' ? 'secondary' : 
                'default'
              }
              size="sm"
            >
              {loading ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Connecting...
                </>
              ) : (
                BUTTON_TEXT[buttonState]
              )}
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};