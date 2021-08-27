module.exports = {
  extends: [
    "eslint:recommended",
    "plugin:vue/vue3-recommended", 
    "@vue/typescript",
    "plugin:prettier/recommended",
    "prettier",
  ],
  env: {
    node: true,
  },
  rules: {
    "no-unused-vars": ["error", { argsIgnorePattern: "^_" }],
  },
};
