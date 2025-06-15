// ConnectorGrid Component - Grid Layout for Connector Cards

import React from 'react';
import { Connector, Integration, ConnectorType } from '../../../types/integration';
import { ConnectorCard } from './ConnectorCard';

interface ConnectorGridProps {
  connectors: Connector[];
  integrations: Integration[];
  onConnectorAction: (connectorType: ConnectorType, action: 'connect' | 'details') => void;
  loadingConnectors: Set<ConnectorType>;
}

export const ConnectorGrid: React.FC<ConnectorGridProps> = ({
  connectors,
  integrations,
  onConnectorAction,
  loadingConnectors
}) => {
  // Create a map for quick integration lookup
  const integrationMap = new Map<ConnectorType, Integration>();
  integrations.forEach(integration => {
    integrationMap.set(integration.connectorType, integration);
  });

  // Group connectors by status for better organization
  const connectedConnectors = connectors.filter(connector => {
    const integration = integrationMap.get(connector.type);
    return integration && (integration.status === 'active' || integration.status === 'connected');
  });
  
  const availableConnectors = connectors.filter(connector => {
    const integration = integrationMap.get(connector.type);
    return connector.isImplemented && 
      (!integration || (integration.status !== 'active' && integration.status !== 'connected'));
  });
  
  const comingSoonConnectors = connectors.filter(connector => 
    !connector.isImplemented
  );

  const renderConnectorSection = (
    title: string,
    connectorList: Connector[],
    showIfEmpty: boolean = false
  ) => {
    if (connectorList.length === 0 && !showIfEmpty) {
      return null;
    }

    return (
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900">
            {title} ({connectorList.length})
          </h2>
        </div>
        
        {connectorList.length > 0 ? (
          <div className="space-y-4">
            {connectorList.map((connector) => (
              <ConnectorCard
                key={connector.type}
                connector={connector}
                integration={integrationMap.get(connector.type)}
                onAction={onConnectorAction}
                loading={loadingConnectors.has(connector.type)}
              />
            ))}
          </div>
        ) : (
          <div className="text-center py-8 text-muted-foreground">
            <p>No {title.toLowerCase()} integrations</p>
          </div>
        )}
      </div>
    );
  };

  // If we have a mix of statuses, show sections
  const hasMultipleSections = connectedConnectors.length > 0 && 
    (availableConnectors.length > 0 || comingSoonConnectors.length > 0);

  if (hasMultipleSections) {
    return (
      <div className="space-y-8">
        {renderConnectorSection("Connected", connectedConnectors)}
        {renderConnectorSection("Available", availableConnectors)}
        {renderConnectorSection("Coming Soon", comingSoonConnectors)}
      </div>
    );
  }

  // Otherwise, show all connectors in a single vertical list
  return (
    <div className="space-y-4">
      <div className="space-y-4">
        {connectors.map((connector) => (
          <ConnectorCard
            key={connector.type}
            connector={connector}
            integration={integrationMap.get(connector.type)}
            onAction={onConnectorAction}
            loading={loadingConnectors.has(connector.type)}
          />
        ))}
      </div>

      {/* Empty state */}
      {connectors.length === 0 && (
        <div className="text-center py-12">
          <div className="mx-auto h-12 w-12 text-gray-400 mb-4">
            <svg fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path 
                strokeLinecap="round" 
                strokeLinejoin="round" 
                strokeWidth={1.5} 
                d="M19 11H5m14-7H5a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2V6a2 2 0 00-2-2zM7 7h0m0 4h0m0 4h0m4-8h0m0 4h0m0 4h0m4-8h0m0 4h0m0 4h0"
              />
            </svg>
          </div>
          <h3 className="text-lg font-medium text-gray-900 mb-2">
            No Integrations Available
          </h3>
          <p className="text-muted-foreground">
            Integration connectors are currently being loaded or configured.
          </p>
        </div>
      )}
    </div>
  );
};