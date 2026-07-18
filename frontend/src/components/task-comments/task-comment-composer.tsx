import { SendHorizontal } from 'lucide-react';
import { useState } from 'react';
import { Button } from '../button';
import { LoadingSpinner } from '../loading';
import { TextEditor } from '../../shared/components/text-editor';
import type { SubmitTaskComment } from './use-task-comments';
import { isHtmlContentEmpty } from '@/utils/html';

interface TaskCommentComposerProps {
  isSubmitting: boolean;
  onSubmit: SubmitTaskComment;
}

export const TaskCommentComposer = ({ isSubmitting, onSubmit }: TaskCommentComposerProps) => {
  const [draft, setDraft] = useState('');
  const [composerKey, setComposerKey] = useState(0);

  const handleSubmit = async () => {
    if (isHtmlContentEmpty(draft)) {
      return;
    }

    await onSubmit({ content: draft, parentCommentId: null });
    setDraft('');
    setComposerKey((current) => current + 1);
  };

  return (
    <>
      <TextEditor
        key={composerKey}
        initialValue=""
        onChange={setDraft}
        placeholder="Share context, decisions, or blockers..."
      />
      <div className="mt-3 flex items-center justify-between gap-3">
        <Button
          type="button"
          onClick={handleSubmit}
          disabled={isSubmitting || isHtmlContentEmpty(draft)}
          className="ml-auto shrink-0"
        >
          {isSubmitting ? <LoadingSpinner size="1rem" /> : <SendHorizontal className="h-4 w-4" />}
          Comment
        </Button>
      </div>
    </>
  );
};
