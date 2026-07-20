import { handleError } from './handle-error';

export const copyToClipboard = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text);
  } catch (error) {
    handleError('Failed to copy to clipboard');
  }
};
