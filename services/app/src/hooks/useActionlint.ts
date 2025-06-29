import { useState, useEffect, useCallback, useRef } from 'react';

// TypeScript types for actionlint errors
export interface ActionlintError {
  kind: string;
  message: string;
  line: number;
  column: number;
}

export interface ActionlintState {
  isLoading: boolean;
  isInitialized: boolean;
  errors: ActionlintError[];
  error: string | null;
  lastValidated: string | null;
  validatedAt: Date | null;
}

export interface ValidationCache {
  content: string;
  errors: ActionlintError[];
  timestamp: number;
}

export interface UseActionlintOptions {
  /**
   * Debounce delay in milliseconds for validation
   * @default 300
   */
  debounceMs?: number;
  /**
   * Auto-validate on content change
   * @default true
   */
  autoValidate?: boolean;
  /**
   * Path to the WASM file
   * @default '/main.wasm'
   */
  wasmPath?: string;
  /**
   * Path to the wasm_exec.js file
   * @default '/wasm_exec.js'
   */
  wasmExecPath?: string;
  /**
   * Enable caching of validation results
   * @default true
   */
  enableCache?: boolean;
  /**
   * Cache TTL in milliseconds
   * @default 60000 (1 minute)
   */
  cacheTtl?: number;
  /**
   * Maximum number of cached results
   * @default 10
   */
  maxCacheSize?: number;
}

// Extend window interface for WASM functions
declare global {
  interface Window {
    runActionlint?: (src: string) => void;
    Go?: new () => {
      importObject: WebAssembly.Imports;
      run: (instance: WebAssembly.Instance) => Promise<void>;
    };
    onCheckCompleted?: (errors: ActionlintError[]) => void;
    showError?: (message: string) => void;
    dismissLoading?: () => void;
    getYamlSource?: () => string;
  }
}

/**
 * React hook for integrating WASM actionlint validation
 * 
 * This hook provides a clean interface for validating YAML workflows using
 * the actionlint WASM module while handling loading states, errors, and cleanup.
 * 
 * @example
 * ```tsx
 * const { validateYaml, state, reset } = useActionlint({
 *   debounceMs: 500,
 *   autoValidate: true
 * });
 * 
 * // Validate YAML content
 * validateYaml(yamlContent);
 * 
 * // Access validation state
 * if (state.isLoading) {
 *   return <div>Validating...</div>;
 * }
 * 
 * if (state.errors.length > 0) {
 *   return <ErrorList errors={state.errors} />;
 * }
 * ```
 */
export function useActionlint(options: UseActionlintOptions = {}) {
  const {
    debounceMs = 300,
    autoValidate = true,
    wasmPath = '/main.wasm',
    wasmExecPath = '/wasm_exec.js',
    enableCache = true,
    cacheTtl = 60000,
    maxCacheSize = 10
  } = options;

  const [state, setState] = useState<ActionlintState>({
    isLoading: false,
    isInitialized: false,
    errors: [],
    error: null,
    lastValidated: null,
    validatedAt: null
  });

  // Use refs to store debounce timer and current validation content
  const debounceTimerRef = useRef<NodeJS.Timeout | null>(null);
  const lastValidationContentRef = useRef<string>('');
  const wasmInstanceRef = useRef<WebAssembly.Instance | null>(null);
  const goRef = useRef<any>(null);
  const cacheRef = useRef<Map<string, ValidationCache>>(new Map());

  // Cache management functions
  const getCachedResult = useCallback((content: string): ActionlintError[] | null => {
    if (!enableCache) return null;
    
    const cached = cacheRef.current.get(content);
    if (!cached) return null;
    
    const now = Date.now();
    if (now - cached.timestamp > cacheTtl) {
      cacheRef.current.delete(content);
      return null;
    }
    
    return cached.errors;
  }, [enableCache, cacheTtl]);

  const setCachedResult = useCallback((content: string, errors: ActionlintError[]) => {
    if (!enableCache) return;
    
    const cache = cacheRef.current;
    
    // Clean up old entries if cache is full
    if (cache.size >= maxCacheSize) {
      const oldestKey = cache.keys().next().value;
      if (oldestKey) {
        cache.delete(oldestKey);
      }
    }
    
    cache.set(content, {
      content,
      errors,
      timestamp: Date.now()
    });
  }, [enableCache, maxCacheSize]);

  const clearCache = useCallback(() => {
    cacheRef.current.clear();
  }, []);

  // Initialize WASM module
  const initializeWasm = useCallback(async () => {
    try {
      console.log('[DEBUG] Starting WASM initialization with paths:', { wasmPath, wasmExecPath });
      setState(prev => ({ ...prev, isLoading: true, error: null }));

      // Load wasm_exec.js if not already loaded
      if (!window.Go) {
        await new Promise<void>((resolve, reject) => {
          const script = document.createElement('script');
          script.src = wasmExecPath;
          script.onload = () => resolve();
          script.onerror = () => reject(new Error(`Failed to load ${wasmExecPath}`));
          document.head.appendChild(script);
        });
      }

      if (!window.Go) {
        throw new Error('Go class not available after loading wasm_exec.js');
      }

      // Create Go instance
      const go = new window.Go();
      goRef.current = go;

      // CRITICAL FIX: Set up getYamlSource function BEFORE running WASM
      // This function is called by the WASM module during initialization
      window.getYamlSource = () => {
        console.log('[DEBUG] getYamlSource called during WASM initialization');
        // Return default YAML content for initial validation
        // The WASM module calls this during its main() function
        const content = lastValidationContentRef.current || `name: Default
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4`;
        console.log('[DEBUG] getYamlSource returning:', content.substring(0, 100) + '...');
        return content;
      };

      // Set up callback for validation results
      window.onCheckCompleted = (errors: ActionlintError[]) => {
        console.log('[DEBUG] onCheckCompleted called with errors:', errors);
        const content = lastValidationContentRef.current;
        console.log('[DEBUG] Current content for validation:', content ? content.substring(0, 100) + '...' : 'null');
        
        // Cache the result
        setCachedResult(content, errors);
        
        console.log('[DEBUG] Setting state with errors count:', errors.length);
        setState(prev => ({
          ...prev,
          isLoading: false,
          errors,
          error: null,
          lastValidated: content,
          validatedAt: new Date()
        }));
      };

      // Set up error callback
      window.showError = (message: string) => {
        console.log('[DEBUG] showError called with message:', message);
        setState(prev => ({
          ...prev,
          isLoading: false,
          error: message,
          errors: []
        }));
      };

      // Set up loading dismissal callback
      window.dismissLoading = () => {
        console.log('[DEBUG] dismissLoading called');
        setState(prev => ({ ...prev, isLoading: false }));
      };

      // Load and instantiate WASM module
      let result: WebAssembly.WebAssemblyInstantiatedSource;
      
      if (typeof WebAssembly.instantiateStreaming === 'function') {
        // Use streaming if available (not available in Safari)
        result = await WebAssembly.instantiateStreaming(
          fetch(wasmPath),
          go.importObject
        );
      } else {
        // Fallback for Safari and other browsers without streaming support
        const response = await fetch(wasmPath);
        const wasmBuffer = await response.arrayBuffer();
        result = await WebAssembly.instantiate(wasmBuffer, go.importObject);
      }

      wasmInstanceRef.current = result.instance;

      // Run the Go WASM module
      // This will call getYamlSource, onCheckCompleted, and dismissLoading
      console.log('[DEBUG] Starting WASM go.run()');
      await go.run(result.instance);
      console.log('[DEBUG] WASM go.run() completed');

      setState(prev => ({
        ...prev,
        isLoading: false,
        isInitialized: true,
        error: null
      }));
      
      console.log('[DEBUG] WASM initialization complete, module ready for validation');

    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to initialize WASM module';
      setState(prev => ({
        ...prev,
        isLoading: false,
        isInitialized: false,
        error: errorMessage
      }));
    }
  }, [wasmPath, wasmExecPath, setCachedResult]);

  // Initialize WASM on mount
  useEffect(() => {
    initializeWasm();

    // Cleanup on unmount
    return () => {
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current);
      }
      
      // Clean up global callbacks
      if (typeof window !== 'undefined') {
        delete window.onCheckCompleted;
        delete window.showError;
        delete window.dismissLoading;
        delete window.getYamlSource;
      }
      
      // Note: We don't clean up the WASM instance as it's managed by Go runtime
      // and cleanup could interfere with other parts of the application
    };
  }, [initializeWasm]);

  // Validate YAML content
  const validateYaml = useCallback((content: string, immediate = false) => {
    console.log('[DEBUG] validateYaml called with content length:', content.length);
    console.log('[DEBUG] state.isInitialized:', state.isInitialized);
    console.log('[DEBUG] window.runActionlint available:', !!window.runActionlint);
    
    if (!state.isInitialized || !window.runActionlint) {
      console.log('[DEBUG] WASM not ready, showing error');
      setState(prev => ({
        ...prev,
        error: 'WASM module not initialized yet. Please wait and try again.'
      }));
      return;
    }

    // Check cache first
    const cachedErrors = getCachedResult(content);
    if (cachedErrors !== null) {
      setState(prev => ({
        ...prev,
        isLoading: false,
        errors: cachedErrors,
        error: null,
        lastValidated: content,
        validatedAt: new Date()
      }));
      return;
    }

    // Store the content for potential later validation
    lastValidationContentRef.current = content;
    console.log('[DEBUG] Stored content for validation:', content.substring(0, 100) + '...');

    // Clear existing debounce timer
    if (debounceTimerRef.current) {
      clearTimeout(debounceTimerRef.current);
      debounceTimerRef.current = null;
    }

    const performValidation = () => {
      // Check if content has changed since scheduling validation
      if (lastValidationContentRef.current !== content) {
        return;
      }

      // Check cache again in case it was populated during debounce
      const cachedErrors = getCachedResult(content);
      if (cachedErrors !== null) {
        setState(prev => ({
          ...prev,
          isLoading: false,
          errors: cachedErrors,
          error: null,
          lastValidated: content,
          validatedAt: new Date()
        }));
        return;
      }

      setState(prev => ({
        ...prev,
        isLoading: true,
        error: null
      }));

      try {
        // Call the WASM validation function
        console.log('[DEBUG] Calling window.runActionlint with content:', content.substring(0, 100) + '...');
        window.runActionlint!(content);
        console.log('[DEBUG] window.runActionlint call completed');
      } catch (err) {
        console.log('[DEBUG] Error in window.runActionlint:', err);
        const errorMessage = err instanceof Error ? err.message : 'Validation failed';
        setState(prev => ({
          ...prev,
          isLoading: false,
          error: errorMessage,
          errors: [],
          lastValidated: null,
          validatedAt: null
        }));
      }
    };

    if (immediate) {
      performValidation();
    } else {
      debounceTimerRef.current = setTimeout(performValidation, debounceMs);
    }
  }, [state.isInitialized, debounceMs, getCachedResult]);

  // Validate immediately (skip debounce)
  const validateImmediate = useCallback((content: string) => {
    validateYaml(content, true);
  }, [validateYaml]);

  // Reset validation state
  const reset = useCallback(() => {
    if (debounceTimerRef.current) {
      clearTimeout(debounceTimerRef.current);
      debounceTimerRef.current = null;
    }
    
    setState(prev => ({
      ...prev,
      isLoading: false,
      errors: [],
      error: null,
      lastValidated: null,
      validatedAt: null
    }));
  }, []);

  // Re-initialize WASM module
  const reinitialize = useCallback(() => {
    setState(prev => ({
      ...prev,
      isInitialized: false,
      errors: [],
      error: null
    }));
    initializeWasm();
  }, [initializeWasm]);

  return {
    /**
     * Current validation state
     */
    state,
    
    /**
     * Validate YAML content with debouncing
     * @param content - YAML content to validate
     * @param immediate - Skip debouncing if true
     */
    validateYaml,
    
    /**
     * Validate YAML content immediately (no debouncing)
     * @param content - YAML content to validate
     */
    validateImmediate,
    
    /**
     * Reset validation state and clear any pending validations
     */
    reset,
    
    /**
     * Re-initialize the WASM module
     */
    reinitialize,
    
    /**
     * Clear the validation cache
     */
    clearCache,
    
    /**
     * Check if the hook is ready for validation
     */
    isReady: state.isInitialized && !state.isLoading,
    
    /**
     * Check if there are validation errors
     */
    hasErrors: state.errors.length > 0,
    
    /**
     * Check if there's a system error (not validation errors)
     */
    hasSystemError: state.error !== null,
    
    /**
     * Get cache statistics
     */
    cacheStats: {
      size: cacheRef.current.size,
      maxSize: maxCacheSize,
      enabled: enableCache
    }
  };
}

export default useActionlint;