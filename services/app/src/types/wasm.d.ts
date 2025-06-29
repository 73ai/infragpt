// TypeScript definitions for WASM actionlint integration

export interface ActionlintError {
  message: string;
  line: number;
  column: number;
  kind: string;
}

export interface ActionlintResult {
  errors: ActionlintError[];
}

declare global {
  interface Window {
    runActionlint: (source: string) => void;
    onCheckCompleted: (errors: ActionlintError[]) => void;
    showError: (message: string) => void;
    dismissLoading: () => void;
    getYamlSource: () => string;
    Go: new () => {
      run: (instance: WebAssembly.Instance) => Promise<void>;
      importObject: WebAssembly.Imports;
    };
  }
}

export {};