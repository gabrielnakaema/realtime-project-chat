import { useEffect, useEffectEvent, useRef } from 'react';
import type { RefObject } from 'react';

interface UseInfiniteScrollObserverOptions {
  onLoadMore: () => void;
  rootMargin?: string;
  rootRef?: RefObject<Element | null>;
  threshold?: number | number[];
}

export const useInfiniteScrollObserver = <T extends Element = HTMLDivElement>({
  onLoadMore,
  rootMargin = '200px',
  rootRef,
  threshold = 0.1,
}: UseInfiniteScrollObserverOptions) => {
  const targetRef = useRef<T>(null);

  const handleIntersect = useEffectEvent(() => {
    onLoadMore();
  });

  useEffect(() => {
    const target = targetRef.current;
    if (!target) return;

    const root = rootRef?.current ?? null;

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) {
          handleIntersect();
        }
      },
      {
        root,
        rootMargin,
        threshold,
      },
    );

    observer.observe(target);

    return () => observer.disconnect();
  }, [rootMargin, rootRef, threshold]);

  return targetRef;
};
