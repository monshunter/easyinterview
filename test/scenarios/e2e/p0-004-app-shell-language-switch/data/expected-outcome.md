# E2E.P0.004 Expected Outcome

- Default Chinese shell remains available before switching.
- After selecting English from the dropdown:
  - `topbar-lang-toggle` is a button that opens `topbar-lang-menu`, aligned with `ui-design/src/app.jsx`.
  - `topbar-lang-option-en` selects English from the menu.
  - `topbar-nav-home` shows `Home`.
  - `topbar-nav-workspace` shows `Mock Interview`.
  - `topbar-nav-resume_versions` shows `Resume`.
  - `topbar-login` shows `Sign in`.
  - Auth/settings/placeholder shell labels are English.
- Stable route/test IDs remain unchanged; no localized route names are introduced.
- Generated client requests made after locale toggle include
  `Accept-Language: en`.
- Legacy entries and prototype data imports remain absent from scenario output.
