import { useEffect, useRef } from 'react';
import styles from './use-target-comment-highlight.module.css';

interface UseTargetCommentHighlightParams {
  open: boolean;
  targetCommentId?: string;
}

export const useTargetCommentHighlight = ({ open, targetCommentId }: UseTargetCommentHighlightParams) => {
  const highlightedElementRef = useRef<HTMLElement | null>(null);

  useEffect(() => {
    const prev = highlightedElementRef.current;
    if (prev) {
      prev.classList.remove(styles.highlight);
      highlightedElementRef.current = null;
    }

    if (!open || !targetCommentId) return;

    const targetElement = document.getElementById(`task-comment-${targetCommentId}`);
    if (!targetElement) return;

    targetElement.scrollIntoView({ behavior: 'smooth', block: 'center' });

    void targetElement.offsetWidth;
    targetElement.classList.add(styles.highlight);
    highlightedElementRef.current = targetElement;

    return () => {
      targetElement.classList.remove(styles.highlight);
      highlightedElementRef.current = null;
    };
  }, [open, targetCommentId]);
};
