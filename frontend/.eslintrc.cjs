// Frontend ESLint config — minimal placeholder owned by D1 frontend-shell.
//
// B1 shared-conventions-codified does NOT enforce error-code casing or
// boundary via ESLint (ESLint is not a hard install at this stage). Local
// enforcement runs through scripts/lint/error_codes.py, invoked by
// `make lint`. D1 may extend this config; it MUST NOT relax the boundary
// rule (no ERROR_CODES literal outside frontend/src/lib/conventions/errors.ts).
//
// When ESLint is installed by D1 the recommended starter rule set is below;
// keep the no-restricted-syntax pattern that matches the Python script so
// both gates report the same violation.
/** @type {import('eslint').Linter.Config} */
module.exports = {
  root: true,
  env: { browser: true, es2022: true, node: true },
  parserOptions: { ecmaVersion: 2022, sourceType: 'module' },
  rules: {
    // Boundary placeholder. The pattern must not appear in any module other
    // than frontend/src/lib/conventions/errors.ts; scripts/lint/error_codes.py
    // is the authoritative gate today.
    // 'no-restricted-syntax': [
    //   'error',
    //   {
    //     selector: "VariableDeclarator[id.name='ERROR_CODES']",
    //     message:
    //       'ERROR_CODES literals are owned by lib/conventions/errors.ts (B1 D-5).',
    //   },
    // ],
  },
};
