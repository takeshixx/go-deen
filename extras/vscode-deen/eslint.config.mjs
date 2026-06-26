import js from "@eslint/js";
import tseslint from "typescript-eslint";

export default tseslint.config(
  js.configs.recommended,
  ...tseslint.configs.recommended,
  {
    files: ["src/**/*.ts"],
    rules: {
      semi: ["error", "always"],
      eqeqeq: "warn",
      curly: "warn",
      "no-eval": "warn",
      "no-throw-literal": "warn",
    },
  },
);
