/**
 * Tests for useActionlint hook
 *
 * Note: These are example tests showing how to test the hook.
 * Full testing would require setting up WASM mocks or integration tests.
 */

import { renderHook, act } from "@testing-library/react";
import { useActionlint } from "../useActionlint";

// Mock the WASM-related globals
const mockWasm = {
  runActionlint: jest.fn(),
  onCheckCompleted: jest.fn(),
  showError: jest.fn(),
  dismissLoading: jest.fn(),
  Go: jest.fn().mockImplementation(() => ({
    importObject: {},
    run: jest.fn().mockResolvedValue(undefined),
  })),
};

// Mock fetch for WASM loading
global.fetch = jest.fn();

describe("useActionlint", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    // Reset window globals
    Object.assign(window, mockWasm);

    // Mock fetch responses
    (global.fetch as jest.Mock)
      .mockResolvedValueOnce({
        arrayBuffer: () => Promise.resolve(new ArrayBuffer(0)),
      })
      .mockResolvedValueOnce({
        text: () => Promise.resolve("mock wasm_exec.js content"),
      });
  });

  afterEach(() => {
    // Clean up globals
    (window as unknown as { runActionlint?: unknown }).runActionlint = undefined;
    (window as unknown as { onCheckCompleted?: unknown }).onCheckCompleted = undefined;
    (window as unknown as { showError?: unknown }).showError = undefined;
    (window as unknown as { dismissLoading?: unknown }).dismissLoading = undefined;
    (window as unknown as { Go?: unknown }).Go = undefined;
  });

  describe("initialization", () => {
    it("should initialize with default state", () => {
      const { result } = renderHook(() => useActionlint());

      expect(result.current.state).toEqual({
        isLoading: false,
        isInitialized: false,
        errors: [],
        error: null,
        lastValidated: null,
        validatedAt: null,
      });
      expect(result.current.isReady).toBe(false);
      expect(result.current.hasErrors).toBe(false);
      expect(result.current.hasSystemError).toBe(false);
    });

    it("should accept custom options", () => {
      const options = {
        debounceMs: 500,
        enableCache: false,
        maxCacheSize: 20,
      };

      const { result } = renderHook(() => useActionlint(options));

      expect(result.current.cacheStats).toEqual({
        size: 0,
        maxSize: 20,
        enabled: false,
      });
    });
  });

  describe("validation", () => {
    it("should handle validation when not initialized", () => {
      const { result } = renderHook(() => useActionlint());

      act(() => {
        result.current.validateYaml("test content");
      });

      expect(result.current.state.error).toContain(
        "WASM module not initialized",
      );
    });

    it("should validate immediately when requested", () => {
      const { result } = renderHook(() => useActionlint());

      // Mock as initialized
      act(() => {
        result.current.state.isInitialized = true;
        window.runActionlint = jest.fn();
      });

      act(() => {
        result.current.validateImmediate("test content");
      });

      expect(window.runActionlint).toHaveBeenCalledWith("test content");
    });
  });

  describe("state management", () => {
    it("should reset state correctly", () => {
      const { result } = renderHook(() => useActionlint());

      // Set some state
      act(() => {
        result.current.state.errors = [
          { kind: "test", message: "test error", line: 1, column: 1 },
        ];
        result.current.state.error = "test error";
      });

      act(() => {
        result.current.reset();
      });

      expect(result.current.state.errors).toEqual([]);
      expect(result.current.state.error).toBeNull();
      expect(result.current.state.lastValidated).toBeNull();
      expect(result.current.state.validatedAt).toBeNull();
    });

    it("should clear cache correctly", () => {
      const { result } = renderHook(() => useActionlint({ enableCache: true }));

      act(() => {
        result.current.clearCache();
      });

      expect(result.current.cacheStats.size).toBe(0);
    });
  });

  describe("error handling", () => {
    it("should handle system errors", () => {
      const { result } = renderHook(() => useActionlint());

      act(() => {
        result.current.state.error = "System error occurred";
      });

      expect(result.current.hasSystemError).toBe(true);
      expect(result.current.hasErrors).toBe(false);
    });

    it("should handle validation errors", () => {
      const { result } = renderHook(() => useActionlint());

      const mockErrors = [
        { kind: "syntax", message: "Invalid syntax", line: 1, column: 5 },
      ];

      act(() => {
        result.current.state.errors = mockErrors;
      });

      expect(result.current.hasErrors).toBe(true);
      expect(result.current.hasSystemError).toBe(false);
    });
  });

  describe("caching", () => {
    it("should provide cache statistics", () => {
      const { result } = renderHook(() =>
        useActionlint({
          enableCache: true,
          maxCacheSize: 5,
        }),
      );

      expect(result.current.cacheStats).toEqual({
        size: 0,
        maxSize: 5,
        enabled: true,
      });
    });

    it("should disable caching when requested", () => {
      const { result } = renderHook(() =>
        useActionlint({ enableCache: false }),
      );

      expect(result.current.cacheStats.enabled).toBe(false);
    });
  });
});

// Integration test example (would require actual WASM setup)
describe.skip("useActionlint integration", () => {
  it("should validate real YAML content", async () => {
    const { result } = renderHook(() => useActionlint());

    // Wait for initialization
    await act(async () => {
      // This would require actual WASM setup
      await new Promise((resolve) => setTimeout(resolve, 1000));
    });

    expect(result.current.isReady).toBe(true);

    const validYaml = `
name: CI
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: echo "Hello World"
    `;

    act(() => {
      result.current.validateImmediate(validYaml);
    });

    // Wait for validation
    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 100));
    });

    expect(result.current.hasErrors).toBe(false);
  });

  it("should detect errors in invalid YAML", async () => {
    const { result } = renderHook(() => useActionlint());

    // Wait for initialization
    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 1000));
    });

    const invalidYaml = `
name: CI
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5  # This might trigger an error
      - run: echo \${{ github.invalid_context }}  # Invalid context
    `;

    act(() => {
      result.current.validateImmediate(invalidYaml);
    });

    // Wait for validation
    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 100));
    });

    expect(result.current.hasErrors).toBe(true);
    expect(result.current.state.errors.length).toBeGreaterThan(0);
  });
});
