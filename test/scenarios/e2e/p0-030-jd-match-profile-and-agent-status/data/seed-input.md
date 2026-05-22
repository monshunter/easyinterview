# Seed Input

- Authenticated runtime using fixture-backed API responses for deterministic UI
  variants, plus a real-mode generated-client gate with
  `VITE_EI_API_MODE=real`.
- Auth fixtures:
  - `getMe.authenticated`
  - `getMe.unauthenticated`
  - `getRuntimeConfig.default`
- JobMatch fixtures:
  - `getJobMatchProfile.default`
  - `getJobMatchProfile.partial-profile`
  - `getJobMatchProfile.unauthenticated`
  - `getAgentScanStatus.idle`
  - `getAgentScanStatus.scanning`
  - `getAgentScanStatus.error`
  - `getAgentScanStatus.next-scan-soon`
- User opens `jd_match`, observes the shell, switches between tabs, and triggers unauthenticated side-effect actions.
- Backend E2E.P0.094-P0.097 provide the paired live route/persistence/AI
  provenance proof for the same JobMatch operation family.
