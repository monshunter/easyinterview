# Seed input

- Trusted current plan: deterministic TargetJob with `targetJobId`, company/title and four canonical rounds matching the static prototype.
- Ready overview: empty round, current report plus newer failed attempt, generating-only round, and current/latest same-ready round.
- State variants: `loading`, `empty`, `error`, `latest-ready`, and target-ID `mismatch` for current-plan isolation.
- Report and Generating fixtures retain reportId-only detail/generation regression coverage and trusted Reports Back behavior.
- Deterministic controls: 1440x900 and 390x844, UTC, fixed clock/locale, loaded fonts, disabled motion, normalized DOM, computed style, absolute bbox and pixelmatch threshold 0.1.
- Negative source vocabulary: full history / Report Center plus every production `listTargetJobReports` consumer outside `frontend/src/app/screens/reports/ReportsScreen.tsx`.
