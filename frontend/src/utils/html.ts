import dompurify from 'dompurify';

export const sanitizeHTML = (html: string) => {
  return dompurify.sanitize(html);
};

export const isHtmlContentEmpty = (html: string | null | undefined): boolean => {
  if (!html) return true;
  const textOnly = html
    .replace(/<[^>]*>/g, '')
    .replace(/&nbsp;/g, ' ')
    .trim();
  return textOnly.length === 0;
};
