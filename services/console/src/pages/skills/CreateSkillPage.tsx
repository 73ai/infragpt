import { SidebarTrigger } from "@/components/ui/sidebar";
import { Button } from "@/components/ui/button";
import ValidationPanel, { ValidationError } from "@/components/ValidationPanel";
import YamlEditor, { YamlEditorRef } from "@/components/YamlEditor";
import AddCommandModal from "@/components/AddCommandModal";
import { useState, useCallback, useEffect, useRef } from "react";
import { useActionlint, ActionlintError } from "@/hooks/useActionlint";

const CreateSkillPage = () => {
  const { validateYaml, state, isReady } = useActionlint({
    debounceMs: 500,
    autoValidate: true,
    wasmPath: "/main.wasm",
    wasmExecPath: "/wasm_exec.js",
  });
  const yamlEditorRef = useRef<YamlEditorRef>(null);
  const [yamlContent, setYamlContent] =
    useState(`# GitHub Actions Workflow Template
name: CI/CD Pipeline
on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - name: Checkout code
      uses: actions/checkout@v4`);

  const [isModalOpen, setIsModalOpen] = useState(false);

  useEffect(() => {
    if (yamlContent.trim()) {
      if (isReady) {
        validateYaml(yamlContent);
      }
    }
  }, [yamlContent, validateYaml, isReady]);

  useEffect(() => {
    if (isReady && yamlContent.trim()) {
      validateYaml(yamlContent);
    }
  }, [isReady, validateYaml, yamlContent]);

  const handleYamlChange = useCallback((newContent: string) => {
    setYamlContent(newContent);
  }, []);

  const handleAddCommand = () => {
    setIsModalOpen(true);
  };

  const handleCommandInsert = (commandYaml: string) => {
    const lines = yamlContent.split("\n");
    let insertIndex = -1;

    const commandLines = commandYaml
      .split("\n")
      .filter((line) => line.trim() !== "");

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];
      const trimmed = line.trim();

      if (
        trimmed.includes("- name:") &&
        (trimmed.includes("Checkout code") ||
          trimmed.toLowerCase().includes("checkout"))
      ) {
        for (let j = i + 1; j < lines.length; j++) {
          const nextLine = lines[j];
          const nextTrimmed = nextLine.trim();

          if (
            nextTrimmed.includes("- name:") ||
            (!nextTrimmed &&
              j + 1 < lines.length &&
              !lines[j + 1].startsWith("  ")) ||
            j === lines.length - 1
          ) {
            insertIndex = j === lines.length - 1 ? lines.length : j;
            break;
          }
        }
        break;
      }
    }

    if (insertIndex === -1) {
      for (let i = 0; i < lines.length; i++) {
        const line = lines[i];
        if (line.trim().startsWith("steps:")) {
          insertIndex = i + 1;
          break;
        }
      }
    }

    if (insertIndex === -1) {
      insertIndex = lines.length;
    }

    const newLines = [
      ...lines.slice(0, insertIndex),
      ...commandLines,
      ...lines.slice(insertIndex),
    ];

    setYamlContent(newLines.join("\n"));
  };

  const handleErrorClick = (error: ValidationError) => {
    if (yamlEditorRef.current) {
      yamlEditorRef.current.setCursor(
        error.line,
        Math.max(0, error.column - 1),
      );
      yamlEditorRef.current.focus();
    }
  };

  const convertedErrors: ValidationError[] = state.errors.map(
    (error: ActionlintError) => ({
      line: error.line,
      column: error.column,
      message: error.message,
      kind: error.kind,
    }),
  );

  return (
    <div className="min-h-screen flex flex-col">
      {/* Header */}
      <div className="border-b">
        <div className="flex h-16 items-center px-4 gap-4 justify-between">
          <div className="flex items-center gap-4">
            <SidebarTrigger />
            <h1 className="text-xl font-semibold">Create a New Skill</h1>
          </div>
          <div className="flex gap-2">
            <Button onClick={handleAddCommand}>Add Command</Button>
            {!isReady && (
              <div className="flex items-center text-sm text-muted-foreground">
                <div className="animate-spin h-4 w-4 border-2 border-primary border-t-transparent rounded-full mr-2" />
                Initializing validator...
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Main Content - True 60/40 Split Layout */}
      <div
        className="flex-1 flex flex-col lg:flex-row"
        style={{ height: "calc(100vh - 64px)" }}
      >
        {/* Left Panel - YAML Editor (60%) */}
        <div className="w-full lg:w-[60%] border-r-0 lg:border-r border-b lg:border-b-0 flex flex-col">
          <div className="flex-1 overflow-hidden">
            <YamlEditor
              ref={yamlEditorRef}
              value={yamlContent}
              onChange={handleYamlChange}
              errors={convertedErrors.map((error) => ({
                line: error.line,
                message: error.message,
              }))}
              placeholder="Enter your GitHub Actions workflow YAML here..."
              className="h-full"
            />
          </div>
        </div>

        {/* Right Panel - Validation Results (40%) */}
        <div className="w-full lg:w-[40%] flex flex-col">
          <div className="flex-1 p-3 overflow-hidden">
            <ValidationPanel
              errors={convertedErrors}
              isLoading={state.isLoading}
              onErrorClick={handleErrorClick}
            />
            {state.error && (
              <div className="mt-4 p-3 bg-destructive/10 border border-destructive/20 rounded-lg">
                <p className="text-sm text-destructive font-medium">
                  System Error:
                </p>
                <p className="text-sm text-destructive mt-1">{state.error}</p>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Add Command Modal */}
      <AddCommandModal
        open={isModalOpen}
        onOpenChange={setIsModalOpen}
        onAddCommand={handleCommandInsert}
      />
    </div>
  );
};

export default CreateSkillPage;
