# Seed Input

- Auth state: frontend fixture-backed app runtime.
- API data source: `openapi/fixtures/TargetJobs/listTargetJobs.json`.
- Fixture scenarios:
  - `empty`: zero TargetJobs for the empty home state.
  - `one-job`: one TargetJob for recent card rendering.
  - `twelve-plus`: 15 TargetJobs for sort-by-`updatedAt desc` and 12-card cap.
- Real backend overlay: `VITE_EI_API_MODE=real` with `VITE_EI_API_BASE_URL=http://localhost:8080/api/v1` for generated-client TargetJob operation routing.
- UI states: zh/en i18n, warm/dark/customAccent theme variants, and mobile layout.
