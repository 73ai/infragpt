import React, { useEffect, useRef } from 'react';
import { EditorView, basicSetup } from 'codemirror';
import { EditorState } from '@codemirror/state';
import { json } from '@codemirror/lang-json';
import { oneDark } from '@codemirror/theme-one-dark';
import { ViewUpdate } from '@codemirror/view';
import { linter, Diagnostic } from '@codemirror/lint';
import { placeholder } from '@codemirror/view';
import { cn } from '../lib/utils';

interface JSONEditorProps {
  value: string;
  onChange: (value: string) => void;
  onValidation?: (errors: string[] | null) => void;
  placeholder?: string;
  className?: string;
  height?: string;
  readOnly?: boolean;
  theme?: 'light' | 'dark';
}

const jsonLinter = linter((view) => {
  const diagnostics: Diagnostic[] = [];
  const doc = view.state.doc.toString();
  if (doc.trim()) {
    try {
      JSON.parse(doc);
    } catch (e) {
      const errorMessage = e instanceof Error ? e.message : 'Invalid JSON';
      const match = errorMessage.match(/position (\d+)/);
      const position = match ? parseInt(match[1]) : 0;
      diagnostics.push({
        from: position,
        to: Math.min(position + 1, doc.length),
        severity: 'error',
        message: errorMessage,
      });
    }
  }
  return diagnostics;
});

export const JSONEditor: React.FC<JSONEditorProps> = ({
  value,
  onChange,
  onValidation,
  placeholder: placeholderText = 'Enter JSON...',
  className,
  height = '300px',
  readOnly = false,
  theme = 'dark',
}) => {
  const editorRef = useRef<HTMLDivElement>(null);
  const viewRef = useRef<EditorView | null>(null);
  const onChangeRef = useRef(onChange);
  const onValidationRef = useRef(onValidation);

  useEffect(() => {
    onChangeRef.current = onChange;
  }, [onChange]);

  useEffect(() => {
    onValidationRef.current = onValidation;
  }, [onValidation]);

  useEffect(() => {
    if (!editorRef.current) return;

    const updateListener = EditorView.updateListener.of((update: ViewUpdate) => {
      if (update.docChanged) {
        const newValue = update.state.doc.toString();
        onChangeRef.current(newValue);

        if (onValidationRef.current) {
          if (newValue.trim()) {
            try {
              JSON.parse(newValue);
              onValidationRef.current(null);
            } catch (e) {
              const errorMessage = e instanceof Error ? e.message : 'Invalid JSON';
              onValidationRef.current([errorMessage]);
            }
          } else {
            onValidationRef.current(null);
          }
        }
      }
    });

    const extensions = [
      basicSetup,
      json(),
      jsonLinter,
      updateListener,
      EditorView.editable.of(!readOnly),
      EditorState.readOnly.of(readOnly),
      EditorView.lineWrapping,
      EditorView.theme({
        '&': {
          height: height,
          fontSize: '13px',
        },
        '.cm-content': {
          fontFamily: 'ui-monospace, SFMono-Regular, "SF Mono", Consolas, "Liberation Mono", Menlo, monospace',
          padding: '12px',
          minHeight: '100%',
          whiteSpace: 'pre-wrap',
          wordBreak: 'break-all',
        },
        '.cm-focused': {
          outline: 'none',
        },
        '.cm-editor': {
          borderRadius: '6px',
          border: '1px solid',
          borderColor: theme === 'dark' ? '#374151' : '#e5e7eb',
          height: '100%',
          maxWidth: '100%',
          overflow: 'hidden',
        },
        '.cm-editor.cm-focused': {
          borderColor: theme === 'dark' ? '#60a5fa' : '#3b82f6',
          boxShadow: '0 0 0 3px rgba(59, 130, 246, 0.1)',
        },
        '.cm-gutters': {
          backgroundColor: theme === 'dark' ? '#1f2937' : '#f9fafb',
          borderRight: '1px solid',
          borderColor: theme === 'dark' ? '#374151' : '#e5e7eb',
          color: theme === 'dark' ? '#6b7280' : '#9ca3af',
        },
        '.cm-activeLineGutter': {
          backgroundColor: theme === 'dark' ? '#374151' : '#e5e7eb',
        },
        '.cm-line': {
          padding: '0 2px 0 6px',
        },
        '.cm-scroller': {
          fontFamily: 'inherit',
          overflowX: 'hidden',
          overflowY: 'auto',
        },
        '.cm-tooltip': {
          backgroundColor: theme === 'dark' ? '#1f2937' : '#ffffff',
          border: '1px solid',
          borderColor: theme === 'dark' ? '#374151' : '#e5e7eb',
          borderRadius: '6px',
          padding: '4px 8px',
        },
        '.cm-diagnostic': {
          padding: '3px 6px 3px 8px',
          marginLeft: '-1px',
          display: 'block',
          whiteSpace: 'pre-wrap',
        },
        '.cm-diagnostic-error': {
          borderLeft: '3px solid #ef4444',
          backgroundColor: theme === 'dark' ? 'rgba(239, 68, 68, 0.1)' : 'rgba(239, 68, 68, 0.05)',
        },
      }),
    ];

    if (theme === 'dark') {
      extensions.push(oneDark);
    }

    const startState = EditorState.create({
      doc: value,
      extensions,
    });

    const view = new EditorView({
      state: startState,
      parent: editorRef.current,
    });

    viewRef.current = view;

    return () => {
      view.destroy();
      viewRef.current = null;
    };
  }, [height, readOnly, theme]);

  useEffect(() => {
    if (viewRef.current) {
      const currentValue = viewRef.current.state.doc.toString();
      if (currentValue !== value) {
        viewRef.current.dispatch({
          changes: {
            from: 0,
            to: currentValue.length,
            insert: value,
          },
        });
      }
    }
  }, [value]);

  return (
    <div
      ref={editorRef}
      className={cn(
        'json-editor',
        'relative w-full overflow-hidden rounded-md',
        className
      )}
      style={{ height: '350px', width: '100%', maxWidth: '100%' }}
    />
  );
};

export default JSONEditor;