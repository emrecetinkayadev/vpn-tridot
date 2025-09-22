import globals from "globals";
import tseslint from "typescript-eslint";

export default tseslint.config(
  {
    files: ["**/*.ts", "**/*.tsx"],
    languageOptions: {
      parser: tseslint.parser,
      parserOptions: {
        project: true,
        ecmaVersion: "latest",
        sourceType: "module",
      },
      globals: {
        ...globals.browser,
        ...globals.node,
      },
    },
    rules: {
      "@typescript-eslint/consistent-type-imports": "warn",
      "@typescript-eslint/no-unused-vars": ["warn", { argsIgnorePattern: "^_" }],
      "no-console": ["warn", { allow: ["warn", "error"] }],
    },
  }
);
