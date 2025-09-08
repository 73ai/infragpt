import React, { useState } from "react";
import { Integration, Connector } from "../../../types/integration";
import { Button } from "../../../components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "../../../components/ui/card";
import {
  TestTube,
  RefreshCw,
  Unplug,
  Settings,
  Loader2,
  AlertTriangle,
} from "lucide-react";

interface IntegrationActionsProps {
  integration: Integration;
  connector: Connector;
  onTestConnection: () => Promise<void>;
  onReconfigure: () => Promise<void>;
  onDisconnect: () => Promise<void>;
  loading?: boolean;
}

export const IntegrationActions: React.FC<IntegrationActionsProps> = ({
  integration,
  connector,
  onTestConnection,
  onReconfigure,
  onDisconnect,
  loading = false,
}) => {
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  const handleAction = async (
    actionType: string,
    action: () => Promise<void>,
  ) => {
    if (loading || actionLoading) return;

    setActionLoading(actionType);
    try {
      await action();
    } finally {
      setActionLoading(null);
    }
  };

  const isActionLoading = (actionType: string) => {
    return loading || actionLoading === actionType;
  };

  const canTestConnection =
    integration.status === "connected" || integration.status === "error";
  const canReconfigure = true; // Most connectors support reconfiguration
  const canDisconnect = true;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center space-x-2">
          <Settings className="h-5 w-5" />
          <span>Actions</span>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <Button
            variant="outline"
            onClick={() => handleAction("test", onTestConnection)}
            disabled={!canTestConnection || isActionLoading("test")}
            className="w-full"
          >
            {isActionLoading("test") ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                Testing...
              </>
            ) : (
              <>
                <TestTube className="h-4 w-4 mr-2" />
                Test Connection
              </>
            )}
          </Button>

          <Button
            variant="outline"
            onClick={() => handleAction("reconfigure", onReconfigure)}
            disabled={!canReconfigure || isActionLoading("reconfigure")}
            className="w-full"
          >
            {isActionLoading("reconfigure") ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                Reconfiguring...
              </>
            ) : (
              <>
                <RefreshCw className="h-4 w-4 mr-2" />
                Reconfigure
              </>
            )}
          </Button>

          <Button
            variant="destructive"
            onClick={() => handleAction("disconnect", onDisconnect)}
            disabled={!canDisconnect || isActionLoading("disconnect")}
            className="w-full"
          >
            {isActionLoading("disconnect") ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                Disconnecting...
              </>
            ) : (
              <>
                <Unplug className="h-4 w-4 mr-2" />
                Disconnect
              </>
            )}
          </Button>
        </div>

        <div className="mt-6 space-y-3 text-sm text-muted-foreground">
          <div className="flex items-start space-x-2">
            <TestTube className="h-4 w-4 mt-0.5 flex-shrink-0" />
            <p>
              <strong>Test Connection:</strong> Verify that the integration is
              working correctly and can communicate with {connector.name}.
            </p>
          </div>

          <div className="flex items-start space-x-2">
            <RefreshCw className="h-4 w-4 mt-0.5 flex-shrink-0" />
            <p>
              <strong>Reconfigure:</strong> Update permissions or reconnect to{" "}
              {connector.name}
              with new settings.
            </p>
          </div>

          <div className="flex items-start space-x-2">
            <Unplug className="h-4 w-4 mt-0.5 flex-shrink-0" />
            <p>
              <strong>Disconnect:</strong> Remove this integration and revoke
              all permissions. This action cannot be undone.
            </p>
          </div>
        </div>

        {connector.type === "slack" && (
          <div className="mt-4 bg-blue-50 border border-blue-200 rounded-lg p-3">
            <div className="flex items-start space-x-2">
              <AlertTriangle className="h-4 w-4 text-blue-600 mt-0.5" />
              <div className="text-sm">
                <p className="text-blue-800 font-medium">
                  Slack Integration Notes
                </p>
                <ul className="text-blue-700 mt-1 space-y-1 list-disc list-inside">
                  <li>Reconfiguring will require workspace admin approval</li>
                  <li>Disconnecting will remove the bot from all channels</li>
                  <li>
                    Test connection verifies bot permissions and channel access
                  </li>
                </ul>
              </div>
            </div>
          </div>
        )}

        {connector.type === "github" && (
          <div className="mt-4 bg-blue-50 border border-blue-200 rounded-lg p-3">
            <div className="flex items-start space-x-2">
              <AlertTriangle className="h-4 w-4 text-blue-600 mt-0.5" />
              <div className="text-sm">
                <p className="text-blue-800 font-medium">
                  GitHub Integration Notes
                </p>
                <ul className="text-blue-700 mt-1 space-y-1 list-disc list-inside">
                  <li>Reconfiguring will update the GitHub App installation</li>
                  <li>
                    Disconnecting will uninstall the app from your repositories
                  </li>
                  <li>
                    Test connection verifies webhook delivery and API access
                  </li>
                </ul>
              </div>
            </div>
          </div>
        )}

        {integration.status === "error" && (
          <div className="mt-4 bg-red-50 border border-red-200 rounded-lg p-3">
            <div className="flex items-start space-x-2">
              <AlertTriangle className="h-4 w-4 text-red-600 mt-0.5" />
              <div className="text-sm">
                <p className="text-red-800 font-medium">Integration Error</p>
                <p className="text-red-700 mt-1">
                  This integration is experiencing issues. Try testing the
                  connection or reconfiguring to resolve the problem.
                </p>
              </div>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
};
