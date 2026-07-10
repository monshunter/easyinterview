# Seed Input

- Auth state: frontend fixture-backed app runtime.
- API data source: `openapi/fixtures/TargetJobs/listTargetJobs.json`.
- Fixture scenarios:
  - `empty`: zero TargetJobs for the empty home state.
  - `one-job`: one TargetJob for recent card rendering.
  - `twelve-plus`: 15 TargetJobs for sort-by-`updatedAt desc`, 3-card cap, and More CTA.
- Real-mode generated-client input: `VITE_EI_API_MODE=real` with `VITE_EI_API_BASE_URL=http://localhost:8080/api/v1`; the test uses stub fetch and makes no network request.
- UI states: default/empty/one-job/twelve-plus TargetJob variants, current resume variants and English i18n.
