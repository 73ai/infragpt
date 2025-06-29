import { SidebarTrigger } from "@/components/ui/sidebar";
import { Button } from '@/components/ui/button';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import ValidationPanel, { ValidationError } from '@/components/ValidationPanel';
import YamlEditor, { YamlEditorRef } from '@/components/YamlEditor';
import AddCommandModal from '@/components/AddCommandModal';
import { useState, useCallback, useEffect, useRef } from 'react';
import { useActionlint, ActionlintError } from '@/hooks/useActionlint';

const CreateSkillPage = () => {
  const { validateYaml, state, isReady } = useActionlint({
    debounceMs: 500,
    autoValidate: true,
    wasmPath: '/main.wasm',
    wasmExecPath: '/wasm_exec.js'
  });
  const yamlEditorRef = useRef<YamlEditorRef>(null);
  const [yamlContent, setYamlContent] = useState(`# Test with invalid GitHub Actions YAML
name: Test
on: push
jobs:
  test:
    runs-on: invalid-runner
    steps:
    - uses: nonexistent/action@v999
    - name: Invalid step
      run: echo "test"
      uses: actions/checkout@v4`);
  
  const [isModalOpen, setIsModalOpen] = useState(false);

  // Auto-validate YAML content when it changes
  useEffect(() => {
    console.log('[DEBUG] CreateSkillPage useEffect - isReady:', isReady, 'content length:', yamlContent.length);
    if (yamlContent.trim()) {
      if (isReady) {
        console.log('[DEBUG] Calling validateYaml from CreateSkillPage');
        validateYaml(yamlContent);
      } else {
        console.log('[DEBUG] WASM not ready yet, validation will be deferred');
      }
    }
  }, [yamlContent, validateYaml, isReady]);

  // Trigger validation when WASM becomes ready
  useEffect(() => {
    if (isReady && yamlContent.trim()) {
      console.log('[DEBUG] WASM became ready, triggering initial validation');
      validateYaml(yamlContent);
    }
  }, [isReady]);

  const handleYamlChange = useCallback((newContent: string) => {
    setYamlContent(newContent);
  }, []);

  const handleAddCommand = () => {
    setIsModalOpen(true);
  };

  const handleCommandInsert = (commandYaml: string) => {
    // Find the insertion point - look for the jobs section or commands section
    const lines = yamlContent.split('\n');
    let insertIndex = -1;
    
    // Look for existing jobs section
    for (let i = 0; i < lines.length; i++) {
      if (lines[i].trim().startsWith('jobs:')) {
        // Find the end of the current jobs section to append new command
        for (let j = i + 1; j < lines.length; j++) {
          if (lines[j].trim() === '' || !lines[j].startsWith(' ')) {
            insertIndex = j;
            break;
          }
        }
        break;
      }
    }
    
    // If no jobs section found or insert point not found, append at the end
    if (insertIndex === -1) {
      insertIndex = lines.length;
    }
    
    // Insert the command YAML with proper spacing
    const commandLines = commandYaml.split('\n');
    const newLines = [
      ...lines.slice(0, insertIndex),
      '',
      '  # New command added',
      ...commandLines,
      ...lines.slice(insertIndex)
    ];
    
    setYamlContent(newLines.join('\n'));
  };

  const handleErrorClick = (error: ValidationError) => {
    console.log('Error clicked:', error);
    if (yamlEditorRef.current) {
      // Navigate to the error location in the editor
      // Convert to 0-based column index for CodeMirror (error.column is 1-based)
      yamlEditorRef.current.setCursor(error.line, Math.max(0, error.column - 1));
      yamlEditorRef.current.focus();
    }
  };

  // Convert ActionlintError to ValidationError for the ValidationPanel
  const convertedErrors: ValidationError[] = state.errors.map((error: ActionlintError) => ({
    line: error.line,
    column: error.column,
    message: error.message,
    kind: error.kind
  }));
  
  console.log('[DEBUG] CreateSkillPage state:', { 
    isReady, 
    isLoading: state.isLoading, 
    errorsCount: state.errors.length,
    error: state.error,
    convertedErrorsCount: convertedErrors.length
  });

  return (
    <div className="min-h-screen flex flex-col">
      {/* Header */}
      <div className="border-b">
        <div className="flex h-16 items-center px-4 gap-4 justify-between">
          <div className="flex items-center gap-4">
            <SidebarTrigger />
            <h1 className="text-xl font-semibold">Create a New skill</h1>
          </div>
          <div className="flex gap-2">
            <Button onClick={handleAddCommand}>
              Add Command
            </Button>
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
      <div className="flex-1 flex flex-col lg:flex-row" style={{ height: 'calc(100vh - 64px)' }}>
        {/* Left Panel - YAML Editor (60%) */}
        <div className="w-full lg:w-[60%] border-r-0 lg:border-r border-b lg:border-b-0 flex flex-col">
          <div className="flex-1 p-3 overflow-hidden">
                <YamlEditor
                  ref={yamlEditorRef}
                  value={yamlContent}
                  onChange={handleYamlChange}
                  errors={convertedErrors.map(error => ({
                    line: error.line,
                    message: error.message
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
                <p className="text-sm text-destructive font-medium">System Error:</p>
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