import dompurify from 'dompurify';

export const richTextListClassName = '[&_li]:my-1 [&_ol]:list-decimal [&_ol]:pl-6 [&_ul]:list-disc [&_ul]:pl-6';

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
