import { useLexicalComposerContext } from '@lexical/react/LexicalComposerContext';
import { mergeRegister } from '@lexical/utils';
import {
  $isListNode,
  INSERT_ORDERED_LIST_COMMAND,
  INSERT_UNORDERED_LIST_COMMAND,
  REMOVE_LIST_COMMAND,
} from '@lexical/list';
import {
  $getSelection,
  $isRangeSelection,
  CAN_REDO_COMMAND,
  CAN_UNDO_COMMAND,
  COMMAND_PRIORITY_LOW,
  FORMAT_ELEMENT_COMMAND,
  FORMAT_TEXT_COMMAND,
  REDO_COMMAND,
  SELECTION_CHANGE_COMMAND,
  UNDO_COMMAND,
} from 'lexical';
import {
  AlignCenter,
  AlignJustify,
  AlignLeft,
  AlignRight,
  Bold,
  ItalicIcon,
  List,
  ListOrdered,
  Redo,
  Strikethrough,
  Underline,
  Undo,
} from 'lucide-react';
import { useCallback, useEffect, useRef, useState } from 'react';
import type { ListType } from '@lexical/list';
import { cn } from '@/lib/utils';

export const ToolbarPlugin = () => {
  const [editor] = useLexicalComposerContext();
  const toolbarRef = useRef(null);
  const [canUndo, setCanUndo] = useState(false);
  const [canRedo, setCanRedo] = useState(false);
  const [isBold, setIsBold] = useState(false);
  const [isItalic, setIsItalic] = useState(false);
  const [isUnderline, setIsUnderline] = useState(false);
  const [isStrikethrough, setIsStrikethrough] = useState(false);
  const [listType, setListType] = useState<ListType | null>(null);

  const $updateToolbar = useCallback(() => {
    const selection = $getSelection();
    if ($isRangeSelection(selection)) {
      setIsBold(selection.hasFormat('bold'));
      setIsItalic(selection.hasFormat('italic'));
      setIsUnderline(selection.hasFormat('underline'));
      setIsStrikethrough(selection.hasFormat('strikethrough'));

      const anchorNode = selection.anchor.getNode();
      const topLevelElement = anchorNode.getTopLevelElementOrThrow();
      setListType($isListNode(topLevelElement) ? topLevelElement.getListType() : null);
    }
  }, []);

  useEffect(() => {
    return mergeRegister(
      editor.registerUpdateListener(({ editorState }) => {
        editorState.read(
          () => {
            $updateToolbar();
          },
          { editor },
        );
      }),
      editor.registerCommand(
        SELECTION_CHANGE_COMMAND,
        (_payload, _newEditor) => {
          $updateToolbar();
          return false;
        },
        COMMAND_PRIORITY_LOW,
      ),
      editor.registerCommand(
        CAN_UNDO_COMMAND,
        (payload) => {
          setCanUndo(payload);
          return false;
        },
        COMMAND_PRIORITY_LOW,
      ),
      editor.registerCommand(
        CAN_REDO_COMMAND,
        (payload) => {
          setCanRedo(payload);
          return false;
        },
        COMMAND_PRIORITY_LOW,
      ),
    );
  }, [editor, $updateToolbar]);

  const iconClassName = 'w-5 h-5 text-slate-500';
  const iconButtonClassName = 'toolbar-item spaced hover:bg-gray-100 p-1 rounded-sm dark:hover:bg-slate-600';
  const activeIconButtonClassName = '';
  const activeIconClassName = 'text-slate-900 dark:text-slate-100';
  const dividerClassName = 'h-[16px] w-[1px] min-w-[1px] bg-slate-200 dark:bg-slate-600';

  return (
    <div
      className="flex flex-wrap items-center gap-2 rounded-tl-md rounded-tr-md bg-white p-2 dark:bg-slate-700"
      ref={toolbarRef}
    >
      <button
        disabled={!canUndo}
        onClick={() => {
          editor.dispatchCommand(UNDO_COMMAND, undefined);
        }}
        className={cn(iconButtonClassName, canUndo ? activeIconButtonClassName : '')}
        aria-label="Undo"
        type="button"
      >
        <Undo className={cn(iconClassName, canUndo ? activeIconClassName : '')} />
      </button>
      <button
        disabled={!canRedo}
        onClick={() => {
          editor.dispatchCommand(REDO_COMMAND, undefined);
        }}
        className={cn(iconButtonClassName, canRedo ? activeIconButtonClassName : '')}
        aria-label="Redo"
        type="button"
      >
        <Redo className={cn(iconClassName, canRedo ? activeIconClassName : '')} />
      </button>
      <div className={dividerClassName} />
      <button
        onClick={() => {
          editor.dispatchCommand(
            listType === 'bullet' ? REMOVE_LIST_COMMAND : INSERT_UNORDERED_LIST_COMMAND,
            undefined,
          );
        }}
        className={cn(iconButtonClassName, listType === 'bullet' ? activeIconButtonClassName : '')}
        aria-label="Bulleted List"
        aria-pressed={listType === 'bullet'}
        type="button"
      >
        <List className={cn(iconClassName, listType === 'bullet' ? activeIconClassName : '')} />
      </button>
      <button
        onClick={() => {
          editor.dispatchCommand(listType === 'number' ? REMOVE_LIST_COMMAND : INSERT_ORDERED_LIST_COMMAND, undefined);
        }}
        className={cn(iconButtonClassName, listType === 'number' ? activeIconButtonClassName : '')}
        aria-label="Numbered List"
        aria-pressed={listType === 'number'}
        type="button"
      >
        <ListOrdered className={cn(iconClassName, listType === 'number' ? activeIconClassName : '')} />
      </button>
      <div className={dividerClassName} />
      <button
        onClick={() => {
          editor.dispatchCommand(FORMAT_TEXT_COMMAND, 'bold');
        }}
        className={cn(iconButtonClassName, isBold ? activeIconButtonClassName : '')}
        aria-label="Format Bold"
        type="button"
      >
        <Bold className={cn(iconClassName, isBold ? activeIconClassName : '')} />
      </button>
      <button
        onClick={() => {
          editor.dispatchCommand(FORMAT_TEXT_COMMAND, 'italic');
        }}
        className={cn(iconButtonClassName, isItalic ? activeIconButtonClassName : '')}
        aria-label="Format Italics"
        type="button"
      >
        <ItalicIcon className={cn(iconClassName, isItalic ? activeIconClassName : '')} />
      </button>
      <button
        onClick={() => {
          editor.dispatchCommand(FORMAT_TEXT_COMMAND, 'underline');
        }}
        className={cn(iconButtonClassName, isUnderline ? activeIconButtonClassName : '')}
        aria-label="Format Underline"
        type="button"
      >
        <Underline className={cn(iconClassName, isUnderline ? activeIconClassName : '')} />
      </button>
      <button
        onClick={() => {
          editor.dispatchCommand(FORMAT_TEXT_COMMAND, 'strikethrough');
        }}
        className={cn(iconButtonClassName, isStrikethrough ? activeIconButtonClassName : '')}
        aria-label="Format Strikethrough"
        type="button"
      >
        <Strikethrough className={cn(iconClassName, isStrikethrough ? activeIconClassName : '')} />
      </button>
      <div className={dividerClassName} />
      <button
        onClick={() => {
          editor.dispatchCommand(FORMAT_ELEMENT_COMMAND, 'left');
        }}
        className={iconButtonClassName}
        aria-label="Left Align"
        type="button"
      >
        <AlignLeft className={iconClassName} />
      </button>
      <button
        onClick={() => {
          editor.dispatchCommand(FORMAT_ELEMENT_COMMAND, 'center');
        }}
        className={iconButtonClassName}
        aria-label="Center Align"
        type="button"
      >
        <AlignCenter className={iconClassName} />
      </button>
      <button
        onClick={() => {
          editor.dispatchCommand(FORMAT_ELEMENT_COMMAND, 'right');
        }}
        className={iconButtonClassName}
        aria-label="Right Align"
        type="button"
      >
        <AlignRight className={iconClassName} />
      </button>
      <button
        onClick={() => {
          editor.dispatchCommand(FORMAT_ELEMENT_COMMAND, 'justify');
        }}
        className={iconButtonClassName}
        aria-label="Justify Align"
        type="button"
      >
        <AlignJustify className={iconClassName} />
      </button>{' '}
    </div>
  );
};
