import dompurify from 'dompurify';

export const sanitizeHTML = (html: string) => {
  const purified = dompurify.sanitize(html);

  return purified;
};
