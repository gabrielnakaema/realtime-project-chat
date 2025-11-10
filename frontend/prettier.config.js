//  @ts-check

/** @type {import('prettier').Config} */
const config = {
  semi: true,
  singleQuote: true,
  trailingComma: 'all',
  tabWidth: 2,
  endOfLine: 'lf',
  printWidth: 120,
  plugins: ['prettier-plugin-tailwindcss'],
};

export default config;
