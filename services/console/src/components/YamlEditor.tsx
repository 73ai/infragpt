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
import React, {
  useEffect,
  useRef,
  useMemo,
  useImperativeHandle,
  forwardRef,
} from "react";
import {
  EditorView,
  keymap,
  lineNumbers,
  highlightActiveLineGutter,
  highlightSpecialChars,
  drawSelection,
  dropCursor,
  rectangularSelection,
  crosshairCursor,
} from "@codemirror/view";
import { EditorState, Extension } from "@codemirror/state";
import { yaml } from "@codemirror/lang-yaml";
import { oneDark } from "@codemirror/theme-one-dark";
import { linter, lintGutter } from "@codemirror/lint";
import {
  defaultKeymap,
  history,
  historyKeymap,
  indentWithTab,
} from "@codemirror/commands";
import { highlightSelectionMatches, searchKeymap } from "@codemirror/search";
import {
  autocompletion,
  completionKeymap,
  closeBrackets,
  closeBracketsKeymap,
} from "@codemirror/autocomplete";
import {
  foldGutter,
  indentOnInput,
  indentUnit,
  bracketMatching,
  foldKeymap,
  syntaxHighlighting,
  defaultHighlightStyle,
} from "@codemirror/language";
import * as yamlParser from "js-yaml";

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

export interface YamlEditorRef {
  setCursor: (line: number, column?: number) => void;
  focus: () => void;
}

const YamlEditor = forwardRef<YamlEditorRef, YamlEditorProps>(
  (
    {
      value,
      onChange,
      errors = [],
      className = "",
      placeholder = "Enter YAML configuration...",
    },
    ref,
  ) => {
    const editorRef = useRef<HTMLDivElement>(null);
    const viewRef = useRef<EditorView | null>(null);
    const onChangeRef = useRef(onChange);
    const errorsRef = useRef(errors);

    // Keep refs updated without causing re-renders
    useEffect(() => {
      onChangeRef.current = onChange;
    }, [onChange]);

    useEffect(() => {
      errorsRef.current = errors;
    }, [errors]);

    // Expose imperative methods via ref
    useImperativeHandle(
      ref,
      () => ({
        setCursor: (line: number, column: number = 0) => {
          if (!viewRef.current) return;

          const doc = viewRef.current.state.doc;
          if (line < 1 || line > doc.lines) return;

          try {
            const lineInfo = doc.line(line);
            const pos =
              lineInfo.from + Math.min(Math.max(0, column), lineInfo.length);

            viewRef.current.dispatch({
              selection: { anchor: pos, head: pos },
              scrollIntoView: true,
            });
          } catch (error) {
            console.warn("Failed to set cursor position:", error);
          }
        },
        focus: () => {
          if (viewRef.current) {
            viewRef.current.focus();
          }
        },
      }),
      [],
    );

    // Always use dark theme - no detection needed

    // Create stable linter function
    const yamlLinter = useMemo(() => {
      return linter((view) => {
        const doc = view.state.doc;
        const text = doc.toString();
        const diagnostics: { from: number; to: number; severity: string; message: string }[] = [];

        if (!text.trim()) {
          return diagnostics;
        }

        // YAML syntax validation
        try {
          yamlParser.load(text);
        } catch (error: unknown) {
          const yamlError = error as { mark?: { line: number; column: number }; reason?: string };
          if (yamlError.mark) {
            const line = yamlError.mark.line + 1; // Convert to 1-based line numbers
            const column = yamlError.mark.column;

            if (line <= doc.lines) {
              const lineInfo = doc.line(line);
              const pos = lineInfo.from + Math.min(column, lineInfo.length);

              diagnostics.push({
                from: pos,
                to: Math.min(pos + 1, lineInfo.to),
                severity: "error",
                message: yamlError.reason || "YAML syntax error",
              });
            }
          } else {
            // Generic error for the entire document
            diagnostics.push({
              from: 0,
              to: Math.min(doc.length, 1),
              severity: "error",
              message: error.message || "YAML parsing error",
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
              severity: "error",
              message: error.message,
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
        indentUnit.of("  "), // Use 2 spaces for indentation

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
          "&": {
            fontSize: "14px",
            fontFamily:
              'ui-monospace, SFMono-Regular, "SF Mono", Consolas, "Liberation Mono", Menlo, monospace',
            height: "100%",
          },
          ".cm-content": {
            padding: "8px",
            height: "100%",
            caretColor: "#fff",
          },
          ".cm-focused": {
            outline: "none",
          },
          ".cm-editor": {
            borderRadius: "6px",
            height: "100%",
          },
          ".cm-scroller": {
            scrollbarWidth: "thin",
            height: "100%",
          },
          ".cm-gutters": {
            backgroundColor: "transparent",
            border: "none",
            paddingLeft: "4px",
          },
          ".cm-lineNumbers": {
            color: "hsl(215 20.2% 65.1%)",
            fontSize: "12px",
            minWidth: "32px",
          },
          ".cm-lineNumbers .cm-gutterElement": {
            padding: "0 8px 0 4px",
          },
          ".cm-lint-marker-error": {
            backgroundColor: "rgb(239 68 68)",
          },
          ".cm-foldGutter .cm-gutterElement": {
            padding: "0 4px",
          },
          // Always use dark theme - no light theme overrides needed
        }),

        // Line wrapping and tab configuration
        EditorView.lineWrapping,
        EditorState.tabSize.of(2),
      ];

      // Always add dark theme
      extensions.push(oneDark);

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
    }, [value, yamlLinter]); // Only create editor once

    // No theme change handler needed - always dark theme

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
          effects: [],
        });
      }
    }, [value, yamlLinter, errors]);

    return (
      <div className={`yaml-editor ${className} h-full flex flex-col`}>
        <div
          ref={editorRef}
          className="flex-1 w-full"
          style={{
            fontSize: "14px",
          }}
        />
        {!value && (
          <div className="absolute top-[8px] left-[52px] pointer-events-none text-muted-foreground text-sm">
            {placeholder}
          </div>
        )}
      </div>
    );
  },
);

YamlEditor.displayName = "YamlEditor";

export default YamlEditor;
