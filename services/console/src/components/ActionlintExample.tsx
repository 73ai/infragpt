import React, { useState } from "react";
import { useActionlint, ActionlintError } from "../hooks/useActionlint";

/**
 * Example component demonstrating how to use the useActionlint hook
 */
export function ActionlintExample() {
  const [yamlContent, setYamlContent] = useState(`name: CI
on:
  push:
    branches: [main]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: echo "Hello World"`);

  const {
    state,
    validateYaml,
    validateImmediate,
    reset,
    clearCache,
    isReady,
    hasErrors,
    hasSystemError,
    cacheStats,
  } = useActionlint({
    debounceMs: 500,
    autoValidate: true,
    enableCache: true,
    cacheTtl: 30000, // 30 seconds for demo
    maxCacheSize: 5,
  });

  const handleContentChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const newContent = e.target.value;
    setYamlContent(newContent);
    validateYaml(newContent);
  };

  const handleValidateNow = () => {
    validateImmediate(yamlContent);
  };

  const handleReset = () => {
    reset();
  };

  const handleClearCache = () => {
    clearCache();
  };

  const renderError = (error: ActionlintError, index: number) => (
    <div
      key={index}
      className="p-3 mb-2 bg-red-50 border border-red-200 rounded-md"
    >
      <div className="flex items-start gap-2">
        <span className="inline-block px-2 py-1 text-xs font-medium bg-red-100 text-red-800 rounded">
          line:{error.line}, col:{error.column}
        </span>
        <span className="inline-block px-2 py-1 text-xs font-medium bg-gray-100 text-gray-800 rounded">
          {error.kind}
        </span>
      </div>
      <p className="mt-2 text-sm text-red-700">{error.message}</p>
    </div>
  );

  return (
    <div className="max-w-4xl mx-auto p-6">
      <h1 className="text-3xl font-bold mb-6">
        Actionlint WASM Integration Example
      </h1>

      {/* Status Indicators */}
      <div className="mb-4 flex gap-2">
        <span
          className={`px-3 py-1 rounded-full text-sm font-medium ${
            state.isInitialized
              ? "bg-green-100 text-green-800"
              : "bg-yellow-100 text-yellow-800"
          }`}
        >
          {state.isInitialized ? "WASM Ready" : "Initializing WASM..."}
        </span>

        {state.isLoading && (
          <span className="px-3 py-1 rounded-full text-sm font-medium bg-blue-100 text-blue-800">
            Validating...
          </span>
        )}
      </div>

      {/* Controls */}
      <div className="mb-4 flex gap-2">
        <button
          onClick={handleValidateNow}
          disabled={!isReady}
          className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 disabled:bg-gray-300 disabled:cursor-not-allowed"
        >
          Validate Now
        </button>
        <button
          onClick={handleReset}
          className="px-4 py-2 bg-gray-500 text-white rounded hover:bg-gray-600"
        >
          Reset
        </button>
        <button
          onClick={handleClearCache}
          className="px-4 py-2 bg-orange-500 text-white rounded hover:bg-orange-600"
        >
          Clear Cache ({cacheStats.size}/{cacheStats.maxSize})
        </button>
      </div>

      {/* YAML Editor */}
      <div className="mb-6">
        <label htmlFor="yaml-editor" className="block text-sm font-medium mb-2">
          GitHub Actions YAML Content:
        </label>
        <textarea
          id="yaml-editor"
          value={yamlContent}
          onChange={handleContentChange}
          className="w-full h-64 p-3 font-mono text-sm border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          placeholder="Enter your GitHub Actions YAML here..."
        />
      </div>

      {/* System Error */}
      {hasSystemError && (
        <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-md">
          <h3 className="text-lg font-semibold text-red-800 mb-2">
            System Error
          </h3>
          <p className="text-red-700">{state.error}</p>
        </div>
      )}

      {/* Validation Results */}
      <div className="mb-6">
        <h3 className="text-lg font-semibold mb-3">Validation Results</h3>

        {!hasErrors && !hasSystemError && !state.isLoading && isReady && (
          <div className="p-4 bg-green-50 border border-green-200 rounded-md">
            <p className="text-green-800 font-medium">
              ✅ No validation errors found!
            </p>
          </div>
        )}

        {hasErrors && (
          <div>
            <p className="text-red-700 font-medium mb-3">
              Found {state.errors.length} validation error
              {state.errors.length !== 1 ? "s" : ""}:
            </p>
            {state.errors.map(renderError)}
          </div>
        )}

        {state.isLoading && (
          <div className="p-4 bg-blue-50 border border-blue-200 rounded-md">
            <p className="text-blue-800">🔄 Validating your YAML content...</p>
          </div>
        )}

        {!isReady && !state.isLoading && (
          <div className="p-4 bg-yellow-50 border border-yellow-200 rounded-md">
            <p className="text-yellow-800">
              ⏳ Initializing validation engine...
            </p>
          </div>
        )}
      </div>

      {/* Hook State Debug Info */}
      <details className="mt-8">
        <summary className="cursor-pointer text-sm font-medium text-gray-600 hover:text-gray-800">
          Debug: Hook State
        </summary>
        <pre className="mt-2 p-3 bg-gray-100 rounded text-xs overflow-auto">
          {JSON.stringify(
            {
              isInitialized: state.isInitialized,
              isLoading: state.isLoading,
              isReady,
              hasErrors,
              hasSystemError,
              errorCount: state.errors.length,
              systemError: state.error,
              lastValidated: state.lastValidated
                ? state.lastValidated.substring(0, 50) + "..."
                : null,
              validatedAt: state.validatedAt?.toISOString(),
              cacheStats,
            },
            null,
            2,
          )}
        </pre>
      </details>
    </div>
  );
}

export default ActionlintExample;
