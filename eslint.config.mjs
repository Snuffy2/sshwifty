// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

import globals from "globals";
import path from "node:path";
import { fileURLToPath } from "node:url";
import js from "@eslint/js";
import { FlatCompat } from "@eslint/eslintrc";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const compat = new FlatCompat({
  baseDirectory: __dirname,
  recommendedConfig: js.configs.recommended,
  allConfig: js.configs.all,
});

export default [
  ...compat.extends(
    "plugin:vue/recommended",
    "eslint:recommended",
    "prettier",
    "plugin:prettier/recommended",
  ),
  {
    languageOptions: {
      globals: {
        ...globals.node,
        ...globals.browser,
        $nuxt: true,
      },
      ecmaVersion: 13,
      sourceType: "module",
    },
    rules: {
      "vue/component-name-in-template-casing": ["error", "PascalCase"],
      "vue/multi-word-component-names": "off",
      "vue/no-v-html": "off",
      "no-console": "off",
      "no-debugger": "off",
      "no-unused-vars": [
        "warn",
        {
          args: "after-used",
          argsIgnorePattern: "^_",
          varsIgnorePattern: "^_",
          caughtErrors: "none",
        },
      ],
    },
  },
  {
    files: ["**/*_test.js"],
    languageOptions: {
      globals: {
        ...globals.vitest,
      },
    },
  },
];
