# Seed Input

- Frontend surfaces: mounted `DebriefScreen`, debrief i18n namespace in zh/en, debrief CSS theme tokens, responsive breakpoints, dev mock fixture registry, and debrief privacy boundary.
- Runner inputs:
  - `frontend-real-backend-gate.sh`
  - `debriefI18nCoverage.test.ts`
  - `privacyBoundary.test.ts`
  - `devMockClient.test.ts`
  - frontend build output
  - `tests/pixel-parity/debrief.spec.ts`
  - `frontend_debrief_legacy.py --phase 8.12`
- Scenario tree negative grep scope: P0.065 through P0.069 scenario directories.
