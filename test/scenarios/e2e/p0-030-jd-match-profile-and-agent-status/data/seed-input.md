# Seed Input

- Authenticated runtime using fixture-backed API responses.
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
