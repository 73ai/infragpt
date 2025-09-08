import React from "react";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";

export interface ValidationError {
  line: number;
  column: number;
  message: string;
  kind: string;
}

export interface ValidationPanelProps {
  errors: ValidationError[];
  isLoading: boolean;
  onErrorClick?: (error: ValidationError) => void;
}

const getErrorBadgeVariant = (kind: string) => {
  switch (kind.toLowerCase()) {
    case "error":
      return "destructive";
    case "warning":
      return "secondary";
    case "info":
      return "outline";
    default:
      return "default";
  }
};

const ValidationPanel: React.FC<ValidationPanelProps> = ({
  errors,
  isLoading,
  onErrorClick,
}) => {
  const renderLoadingState = () => (
    <div className="space-y-3">
      <div className="flex items-center space-x-2">
        <Skeleton className="h-4 w-16" />
        <Skeleton className="h-4 w-24" />
      </div>
      <Skeleton className="h-12 w-full" />
      <Skeleton className="h-12 w-full" />
      <Skeleton className="h-12 w-full" />
    </div>
  );

  const renderEmptyState = () => (
    <div className="flex flex-col items-center justify-center h-64 text-center">
      <div className="text-green-500 mb-2">
        <svg
          className="w-12 h-12 mx-auto"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
      </div>
      <p className="text-muted-foreground font-medium">
        No validation errors found
      </p>
      <p className="text-sm text-muted-foreground mt-1">
        Your YAML appears to be valid
      </p>
    </div>
  );

  const renderError = (error: ValidationError, index: number) => (
    <div
      key={index}
      className="border rounded-lg p-3 cursor-pointer hover:bg-muted/50 hover:border-primary/50 transition-all duration-200 hover:shadow-sm"
      onClick={() => onErrorClick?.(error)}
      onKeyDown={(e) => e.key === "Enter" && onErrorClick?.(error)}
      tabIndex={0}
      role="button"
      aria-label={`Navigate to error at line ${error.line}, column ${error.column}`}
      title="Click to navigate to error location in editor"
    >
      <div className="flex items-start justify-between mb-2">
        <div className="flex items-center space-x-2">
          <Badge variant="outline" className="text-xs">
            Line {error.line}, Col {error.column}
          </Badge>
          <Badge variant={getErrorBadgeVariant(error.kind)} className="text-xs">
            {error.kind}
          </Badge>
        </div>
        <div className="text-muted-foreground hover:text-foreground transition-colors">
          <svg
            className="w-4 h-4"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 5l7 7-7 7"
            />
          </svg>
        </div>
      </div>
      <p className="text-sm text-foreground leading-relaxed">{error.message}</p>
    </div>
  );

  const renderErrorList = () => (
    <div className="flex-1 flex flex-col min-h-0">
      <div className="flex items-center justify-between mb-3 flex-shrink-0">
        <span className="text-sm font-medium text-muted-foreground">
          {errors.length} error{errors.length !== 1 ? "s" : ""} found
        </span>
      </div>
      <div className="flex-1 overflow-y-auto space-y-2 min-h-0">
        {errors.map((error, index) => renderError(error, index))}
      </div>
    </div>
  );

  return (
    <div className="h-full flex flex-col">
      {isLoading
        ? renderLoadingState()
        : errors.length === 0
          ? renderEmptyState()
          : renderErrorList()}
    </div>
  );
};

export default ValidationPanel;
