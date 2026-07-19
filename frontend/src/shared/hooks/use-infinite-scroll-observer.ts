import { useCallback, useEffect, useEffectEvent, useState } from 'react';
import type { RefCallback, RefObject } from 'react';

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
  const [target, setTarget] = useState<T | null>(null);

  const targetRef = useCallback<RefCallback<T>>((node) => {
    setTarget(node);
  }, []);

  const handleIntersect = useEffectEvent(() => {
    onLoadMore();
  });

  useEffect(() => {
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
  }, [rootMargin, rootRef, target, threshold]);

  return targetRef;
};
