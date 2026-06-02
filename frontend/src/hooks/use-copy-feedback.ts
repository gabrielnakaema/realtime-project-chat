import { useEffect, useState } from 'react';

export const useCopyFeedback = <T extends string>() => {
  const [copiedValue, setCopiedValue] = useState<T | null>(null);

  useEffect(() => {
    if (!copiedValue) {
      return;
    }

    const timeoutId = window.setTimeout(() => {
      setCopiedValue(null);
    }, 2000);

    return () => window.clearTimeout(timeoutId);
  }, [copiedValue]);

  return {
    copiedValue,
    markCopied: (value: T) => setCopiedValue(value),
    resetCopiedValue: () => setCopiedValue(null),
  };
};
