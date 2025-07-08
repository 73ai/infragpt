/**
 * TypeScript declarations for actionlint WASM integration
 */

declare global {
  interface Window {
    /**
     * Main actionlint validation function exposed by WASM
     * @param src - YAML source code to validate
     */
    runActionlint?: (src: string) => void;
    
    /**
     * Callback function called when validation is completed
     * @param errors - Array of validation errors found
     */
    onCheckCompleted?: (errors: ActionlintError[]) => void;
    
    /**
     * Callback function called when an error occurs during validation
     * @param message - Error message
     */
    showError?: (message: string) => void;
    
    /**
     * Callback function called to dismiss loading state
     */
    dismissLoading?: () => void;
    
    /**
     * Go class for WASM runtime
     */
    Go?: new () => {
      importObject: WebAssembly.Imports;
      run: (instance: WebAssembly.Instance) => Promise<void>;
    };
  }
}

/**
 * Represents a validation error found by actionlint
 */
export interface ActionlintError {
  /** Error type/category */
  kind: string;
  /** Human-readable error message */
  message: string;
  /** Line number where error occurs (1-based) */
  line: number;
  /** Column number where error occurs (1-based) */
  column: number;
}

/**
 * State object for the actionlint validation hook
 */
export interface ActionlintState {
  /** Whether validation is currently in progress */
  isLoading: boolean;
  /** Whether the WASM module has been initialized */
  isInitialized: boolean;
  /** Array of validation errors from the last validation */
  errors: ActionlintError[];
  /** System error message (not validation errors) */
  error: string | null;
  /** The content that was last validated */
  lastValidated: string | null;
  /** Timestamp of last validation */
  validatedAt: Date | null;
}

/**
 * Cached validation result
 */
export interface ValidationCache {
  /** The YAML content that was validated */
  content: string;
  /** Validation errors found for this content */
  errors: ActionlintError[];
  /** Timestamp when this result was cached */
  timestamp: number;
}

/**
 * Configuration options for the useActionlint hook
 */
export interface UseActionlintOptions {
  /** Debounce delay in milliseconds for validation (default: 300) */
  debounceMs?: number;
  /** Auto-validate on content change (default: true) */
  autoValidate?: boolean;
  /** Path to the WASM file (default: '/main.wasm') */
  wasmPath?: string;
  /** Path to the wasm_exec.js file (default: '/wasm_exec.js') */
  wasmExecPath?: string;
  /** Enable caching of validation results (default: true) */
  enableCache?: boolean;
  /** Cache TTL in milliseconds (default: 60000) */
  cacheTtl?: number;
  /** Maximum number of cached results (default: 10) */
  maxCacheSize?: number;
}

/**
 * Cache statistics returned by the hook
 */
export interface CacheStats {
  /** Current number of cached results */
  size: number;
  /** Maximum number of results that can be cached */
  maxSize: number;
  /** Whether caching is enabled */
  enabled: boolean;
}

/**
 * Return type of the useActionlint hook
 */
export interface UseActionlintReturn {
  /** Current validation state */
  state: ActionlintState;
  /** Validate YAML content with debouncing */
  validateYaml: (content: string, immediate?: boolean) => void;
  /** Validate YAML content immediately (no debouncing) */
  validateImmediate: (content: string) => void;
  /** Reset validation state and clear any pending validations */
  reset: () => void;
  /** Re-initialize the WASM module */
  reinitialize: () => void;
  /** Clear the validation cache */
  clearCache: () => void;
  /** Check if the hook is ready for validation */
  isReady: boolean;
  /** Check if there are validation errors */
  hasErrors: boolean;
  /** Check if there's a system error (not validation errors) */
  hasSystemError: boolean;
  /** Cache statistics */
  cacheStats: CacheStats;
}

export {};