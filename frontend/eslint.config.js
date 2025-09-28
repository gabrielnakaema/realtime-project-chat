//  @ts-check

import { tanstackConfig } from '@tanstack/eslint-config';
import pluginQuery from '@tanstack/eslint-plugin-query';
import reactHooks from 'eslint-plugin-react-hooks';

export default [
  ...tanstackConfig,
  ...pluginQuery.configs['flat/recommended'],
  reactHooks.configs['recommended-latest'],
  {
    rules: {
      '@typescript-eslint/array-type': 'off',
    },
  },
];
