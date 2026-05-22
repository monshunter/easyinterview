# Seed Input

- Authenticated runtime using fixture-backed API responses for deterministic UI
  variants, plus a real-mode generated-client gate with
  `VITE_EI_API_MODE=real`.
- JobMatch fixtures:
  - `searchJobs.default`
  - `searchJobs.empty`
  - `searchJobs.failed`
  - `searchJobs.slow-response`
  - `listSavedSearches.default`
  - `listSavedSearches.empty`
  - `listSavedSearches.4xx`
  - `createSavedSearch.default`
  - `createSavedSearch.4xx-validation`
- Unauthenticated runtime using `getMe.unauthenticated` for Run search and Save current pending-action coverage.
- User opens `jd_match`, switches to Search, enters a natural-language query, runs search, toggles result filters, saves the query, and switches tabs during a slow search.
- Backend E2E.P0.094-P0.097 provide the paired live route/persistence/AI
  provenance proof for the same JobMatch operation family.
