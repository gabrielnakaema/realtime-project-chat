import { Button } from '../button';
import { LoadingSpinner } from '../loading';
import { TextEditor } from '../../shared/components/text-editor';
import type { TaskComment } from '@/types/task';
import { isHtmlContentEmpty } from '@/utils/html';

interface ReplyComposerProps {
  comment: TaskComment;
  composerKey: number;
  draft: string;
  isSubmitting: boolean;
  onCancel: () => void;
  onDraftChange: (value: string) => void;
  onSubmit: () => void;
}

export const ReplyComposer = ({
  comment,
  composerKey,
  draft,
  isSubmitting,
  onCancel,
  onDraftChange,
  onSubmit,
}: ReplyComposerProps) => {
  return (
    <div className="border-border bg-muted mt-4 rounded-xl border p-3">
      <TextEditor
        key={`${comment.id}-${composerKey}`}
        initialValue=""
        onChange={onDraftChange}
        placeholder={`Reply to ${comment.user.name}...`}
      />
      <div className="mt-3 flex items-center justify-end gap-2">
        <Button type="button" variant="secondary" onClick={onCancel}>
          Cancel
        </Button>
        <Button type="button" onClick={onSubmit} disabled={isSubmitting || isHtmlContentEmpty(draft)}>
          {isSubmitting ? <LoadingSpinner size="1rem" /> : 'Reply'}
        </Button>
      </div>
    </div>
  );
};
