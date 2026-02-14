import { $generateHtmlFromNodes, $generateNodesFromDOM } from '@lexical/html';
import { AutoFocusPlugin } from '@lexical/react/LexicalAutoFocusPlugin';
import { LexicalComposer } from '@lexical/react/LexicalComposer';
import { ContentEditable } from '@lexical/react/LexicalContentEditable';
import { LexicalErrorBoundary } from '@lexical/react/LexicalErrorBoundary';
import { HistoryPlugin } from '@lexical/react/LexicalHistoryPlugin';
import { OnChangePlugin } from '@lexical/react/LexicalOnChangePlugin';
import { RichTextPlugin } from '@lexical/react/LexicalRichTextPlugin';
import { $getRoot, $insertNodes, ParagraphNode, TextNode } from 'lexical';

import { ToolbarPlugin } from './toolbar-plugin';

import { constructImportMap, exportMap } from './utils';
import type { EditorState, LexicalEditor } from 'lexical';

const editorConfig = {
  html: {
    export: exportMap,
    import: constructImportMap(),
  },
  namespace: 'TextEditor',
  nodes: [ParagraphNode, TextNode],
  onError(error: Error) {
    throw error;
  },
};

interface TextEditorProps {
  initialValue: string;
  onChange: (value: string) => void;
  placeholder?: string;
  label?: string;
  id?: string;
  error?: string;
}

export const TextEditor = ({ initialValue, onChange, label, id, error, placeholder }: TextEditorProps) => {
  const handleChange = (_editorState: EditorState, editor: LexicalEditor) => {
    editor.read(() => {
      const html = $generateHtmlFromNodes(editor);
      onChange(html);
    });
  };

  const initialConfig = {
    ...editorConfig,
    editorState: initialValue
      ? (editor: LexicalEditor) => {
          const parser = new DOMParser();
          const dom = parser.parseFromString(initialValue, 'text/html');
          const nodes = $generateNodesFromDOM(editor, dom);
          $getRoot().select();
          $insertNodes(nodes);
        }
      : undefined,
  };

  return (
    <LexicalComposer initialConfig={initialConfig}>
      <div className="w-full space-y-2">
        {label && (
          <label htmlFor={id} className="block text-sm font-medium text-slate-700 dark:text-slate-300">
            {label}
          </label>
        )}
        <div className="flex w-full flex-col">
          <ToolbarPlugin />
          <div className="relative w-full rounded-b-md border-t border-t-gray-50 bg-white dark:border-t-slate-600 dark:bg-slate-700 dark:text-slate-100">
            <RichTextPlugin
              contentEditable={
                <ContentEditable
                  className="relative min-h-40 p-4 font-sans text-base text-gray-700 outline-none dark:text-slate-100"
                  aria-placeholder={placeholder ?? ''}
                  placeholder={
                    <div className="absolute top-4 left-4 font-sans text-base text-slate-400 dark:text-slate-500">
                      {placeholder}
                    </div>
                  }
                  id={id}
                />
              }
              ErrorBoundary={LexicalErrorBoundary}
            />
            <OnChangePlugin onChange={handleChange} />
            <HistoryPlugin />
            <AutoFocusPlugin />
          </div>
          {error && <p className="text-sm text-red-500">{error}</p>}
        </div>
      </div>
    </LexicalComposer>
  );
};
