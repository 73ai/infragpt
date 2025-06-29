/**
 * YamlEditor Component
 * 
 * A modern YAML editor built with CodeMirror 6 that provides:
 * - Syntax highlighting for YAML
 * - Real-time syntax validation
 * - Error highlighting with gutter indicators
 * - Dark/light theme support that follows the app's theme
 * - Line numbers and auto-indentation with 2-space tabs
 * - Customizable error reporting from external validators
 * 
 * Props:
 * - value: The YAML content string
 * - onChange: Callback when content changes
 * - errors: Array of validation errors with line numbers and messages
 * - className: Additional CSS classes
 * - placeholder: Placeholder text when empty
 */
import React, { useEffect, useRef, useState, useMemo } from 'react';
import { EditorView, keymap, lineNumbers, highlightActiveLineGutter, highlightSpecialChars, drawSelection, dropCursor, rectangularSelection, crosshairCursor } from '@codemirror/view';
import { EditorState, Extension } from '@codemirror/state';
import { yaml } from '@codemirror/lang-yaml';
import { oneDark } from '@codemirror/theme-one-dark';
import { linter, lintGutter } from '@codemirror/lint';
import { defaultKeymap, history, historyKeymap, indentWithTab } from '@codemirror/commands';
import { highlightSelectionMatches, searchKeymap } from '@codemirror/search';
import { autocompletion, completionKeymap, closeBrackets, closeBracketsKeymap } from '@codemirror/autocomplete';
import { foldGutter, indentOnInput, indentUnit, bracketMatching, foldKeymap, syntaxHighlighting, defaultHighlightStyle } from '@codemirror/language';
import * as yamlParser from 'js-yaml';

interface YamlEditorProps {
  value: string;
  onChange: (value: string) => void;
  errors?: Array<{
    line: number;
    message: string;
  }>;
  className?: string;
  placeholder?: string;
}

const YamlEditor: React.FC<YamlEditorProps> = ({
  value,
  onChange,
  errors = [],
  className = '',
  placeholder = 'Enter YAML configuration...'
}) => {
  const editorRef = useRef<HTMLDivElement>(null);
  const viewRef = useRef<EditorView | null>(null);
  const [isDark, setIsDark] = useState(false);
  const onChangeRef = useRef(onChange);
  const errorsRef = useRef(errors);

  // Keep refs updated without causing re-renders
  useEffect(() => {
    onChangeRef.current = onChange;
  }, [onChange]);

  useEffect(() => {
    errorsRef.current = errors;
  }, [errors]);

  // Detect dark mode
  useEffect(() => {
    const checkDarkMode = () => {
      const isDarkMode = document.documentElement.classList.contains('dark') ||
        window.matchMedia('(prefers-color-scheme: dark)').matches;
      setIsDark(isDarkMode);
    };

    checkDarkMode();
    
    // Watch for theme changes
    const observer = new MutationObserver(checkDarkMode);
    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['class']
    });

    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    mediaQuery.addEventListener('change', checkDarkMode);

    return () => {
      observer.disconnect();
      mediaQuery.removeEventListener('change', checkDarkMode);
    };
  }, []);

  // Create stable linter function
  const yamlLinter = useMemo(() => {
    return linter((view) => {
      const doc = view.state.doc;
      const text = doc.toString();
      const diagnostics: any[] = [];

      if (!text.trim()) {
        return diagnostics;
      }

      // YAML syntax validation
      try {
        yamlParser.load(text);
      } catch (error: any) {
        if (error.mark) {
          const line = error.mark.line + 1; // Convert to 1-based line numbers
          const column = error.mark.column;
          
          if (line <= doc.lines) {
            const lineInfo = doc.line(line);
            const pos = lineInfo.from + Math.min(column, lineInfo.length);
            
            diagnostics.push({
              from: pos,
              to: Math.min(pos + 1, lineInfo.to),
              severity: 'error',
              message: error.reason || 'YAML syntax error'
            });
          }
        } else {
          // Generic error for the entire document
          diagnostics.push({
            from: 0,
            to: Math.min(doc.length, 1),
            severity: 'error',
            message: error.message || 'YAML parsing error'
          });
        }
      }

      // Add external errors passed as props
      errorsRef.current.forEach((error) => {
        if (error.line > 0 && error.line <= doc.lines) {
          const lineInfo = doc.line(error.line);
          diagnostics.push({
            from: lineInfo.from,
            to: lineInfo.to,
            severity: 'error',
            message: error.message
          });
        }
      });

      return diagnostics;
    });
  }, []);

  // Create editor only once
  useEffect(() => {
    if (!editorRef.current) return;

    const extensions: Extension[] = [
      // Basic editor features
      lineNumbers(),
      foldGutter(),
      drawSelection(),
      dropCursor(),
      EditorState.allowMultipleSelections.of(true),
      indentOnInput(),
      bracketMatching(),
      closeBrackets(),
      autocompletion(),
      rectangularSelection(),
      crosshairCursor(),
      highlightActiveLineGutter(),
      highlightSelectionMatches(),
      highlightSpecialChars(),
      
      // History and undo/redo support
      history(),
      
      // Language support
      yaml(),
      syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
      indentUnit.of('  '), // Use 2 spaces for indentation
      
      // Linting
      lintGutter(),
      yamlLinter,
      
      // Key bindings - Order matters! Most specific first
      keymap.of([
        ...closeBracketsKeymap,
        ...completionKeymap,
        ...foldKeymap,
        ...searchKeymap,
        ...historyKeymap,
        indentWithTab,
        ...defaultKeymap,
      ]),
      
      // Content change handler
      EditorView.updateListener.of((update) => {
        if (update.docChanged) {
          const newValue = update.state.doc.toString();
          onChangeRef.current(newValue);
        }
      }),
      
      // Theme and styling
      EditorView.theme({
        '&': {
          fontSize: '14px',
          fontFamily: 'ui-monospace, SFMono-Regular, "SF Mono", Consolas, "Liberation Mono", Menlo, monospace',
        },
        '.cm-content': {
          padding: '12px',
          minHeight: '400px',
          caretColor: isDark ? '#fff' : '#000',
        },
        '.cm-focused': {
          outline: 'none',
        },
        '.cm-editor': {
          borderRadius: '6px',
        },
        '.cm-scroller': {
          scrollbarWidth: 'thin',
        },
        '.cm-gutters': {
          backgroundColor: 'transparent',
          border: 'none',
          paddingLeft: '4px',
        },
        '.cm-lineNumbers': {
          color: isDark ? 'hsl(215 20.2% 65.1%)' : 'hsl(215.4 16.3% 46.9%)',
          fontSize: '12px',
          minWidth: '32px',
        },
        '.cm-lineNumbers .cm-gutterElement': {
          padding: '0 8px 0 4px',
        },
        '.cm-lint-marker-error': {
          backgroundColor: 'rgb(239 68 68)',
        },
        '.cm-foldGutter .cm-gutterElement': {
          padding: '0 4px',
        },
        // Custom styles for light theme
        ...(!isDark && {
          '.cm-content': {
            backgroundColor: 'hsl(0 0% 100%)',
            color: 'hsl(222.2 84% 4.9%)',
          },
          '.cm-editor': {
            backgroundColor: 'hsl(0 0% 100%)',
            border: '1px solid hsl(214.3 31.8% 91.4%)',
          },
          '.cm-gutters': {
            backgroundColor: 'hsl(210 40% 96.1%)',
            borderRight: '1px solid hsl(214.3 31.8% 91.4%)',
          },
        }),
      }),
      
      // Line wrapping and tab configuration
      EditorView.lineWrapping,
      EditorState.tabSize.of(2),
    ];

    // Add dark theme if needed
    if (isDark) {
      extensions.push(oneDark);
    }

    const startState = EditorState.create({
      doc: value,
      extensions,
    });

    // Create new view
    viewRef.current = new EditorView({
      state: startState,
      parent: editorRef.current,
    });

    return () => {
      if (viewRef.current) {
        viewRef.current.destroy();
        viewRef.current = null;
      }
    };
  }, []); // Only create editor once

  // Handle theme changes separately
  useEffect(() => {
    if (!viewRef.current) return;
    
    // Force a theme update by recreating the editor when theme changes
    const currentContent = viewRef.current.state.doc.toString();
    const currentSelection = viewRef.current.state.selection.main;
    
    // Destroy current editor
    viewRef.current.destroy();
    
    // Recreate with new theme
    const extensions: Extension[] = [
      // Basic editor features
      lineNumbers(),
      foldGutter(),
      drawSelection(),
      dropCursor(),
      EditorState.allowMultipleSelections.of(true),
      indentOnInput(),
      bracketMatching(),
      closeBrackets(),
      autocompletion(),
      rectangularSelection(),
      crosshairCursor(),
      highlightActiveLineGutter(),
      highlightSelectionMatches(),
      highlightSpecialChars(),
      
      // History and undo/redo support
      history(),
      
      // Language support
      yaml(),
      syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
      indentUnit.of('  '),
      
      // Linting
      lintGutter(),
      yamlLinter,
      
      // Key bindings
      keymap.of([
        ...closeBracketsKeymap,
        ...completionKeymap,
        ...foldKeymap,
        ...searchKeymap,
        ...historyKeymap,
        indentWithTab,
        ...defaultKeymap,
      ]),
      
      // Content change handler
      EditorView.updateListener.of((update) => {
        if (update.docChanged) {
          const newValue = update.state.doc.toString();
          onChangeRef.current(newValue);
        }
      }),
      
      // Theme and styling
      EditorView.theme({
        '&': {
          fontSize: '14px',
          fontFamily: 'ui-monospace, SFMono-Regular, "SF Mono", Consolas, "Liberation Mono", Menlo, monospace',
        },
        '.cm-content': {
          padding: '12px',
          minHeight: '400px',
          caretColor: isDark ? '#fff' : '#000',
        },
        '.cm-focused': {
          outline: 'none',
        },
        '.cm-editor': {
          borderRadius: '6px',
        },
        '.cm-scroller': {
          scrollbarWidth: 'thin',
        },
        '.cm-gutters': {
          backgroundColor: 'transparent',
          border: 'none',
          paddingLeft: '4px',
        },
        '.cm-lineNumbers': {
          color: isDark ? 'hsl(215 20.2% 65.1%)' : 'hsl(215.4 16.3% 46.9%)',
          fontSize: '12px',
          minWidth: '32px',
        },
        '.cm-lineNumbers .cm-gutterElement': {
          padding: '0 8px 0 4px',
        },
        '.cm-lint-marker-error': {
          backgroundColor: 'rgb(239 68 68)',
        },
        '.cm-foldGutter .cm-gutterElement': {
          padding: '0 4px',
        },
        // Custom styles for light theme
        ...(!isDark && {
          '.cm-content': {
            backgroundColor: 'hsl(0 0% 100%)',
            color: 'hsl(222.2 84% 4.9%)',
          },
          '.cm-editor': {
            backgroundColor: 'hsl(0 0% 100%)',
            border: '1px solid hsl(214.3 31.8% 91.4%)',
          },
          '.cm-gutters': {
            backgroundColor: 'hsl(210 40% 96.1%)',
            borderRight: '1px solid hsl(214.3 31.8% 91.4%)',
          },
        }),
      }),
      
      // Line wrapping and tab configuration
      EditorView.lineWrapping,
      EditorState.tabSize.of(2),
    ];

    // Add dark theme if needed
    if (isDark) {
      extensions.push(oneDark);
    }

    const newState = EditorState.create({
      doc: currentContent,
      extensions,
      selection: currentSelection,
    });

    viewRef.current = new EditorView({
      state: newState,
      parent: editorRef.current!,
    });
  }, [isDark, yamlLinter]);

  // Update editor content when value prop changes externally
  useEffect(() => {
    if (viewRef.current && viewRef.current.state.doc.toString() !== value) {
      const currentDoc = viewRef.current.state.doc.toString();
      
      // Only update if the value actually differs to avoid unnecessary updates
      if (currentDoc !== value) {
        viewRef.current.dispatch({
          changes: {
            from: 0,
            to: viewRef.current.state.doc.length,
            insert: value,
          },
        });
      }
    }
  }, [value]);

  // Force linter to re-run when errors change
  useEffect(() => {
    if (viewRef.current) {
      // Trigger a lint update by forcing a document update
      viewRef.current.dispatch({
        effects: []
      });
    }
  }, [errors]);

  return (
    <div className={`yaml-editor ${className}`}>
      <div
        ref={editorRef}
        className="min-h-[400px] w-full"
        style={{
          fontSize: '14px',
        }}
      />
      {!value && (
        <div className="absolute top-[12px] left-[52px] pointer-events-none text-muted-foreground text-sm">
          {placeholder}
        </div>
      )}
    </div>
  );
};

export default YamlEditor;