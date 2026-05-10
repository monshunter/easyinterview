# Expected Outcome

- Search tab renders natural-language input, Run button, source chips, saved searches, and result filters.
- `searchJobs` receives the typed query and an Idempotency-Key.
- Slow-response fixture keeps the five-step AGENT scanning panel visible until the request settles or is aborted.
- Leaving the Search tab clears query and filter state, aborts in-flight search, and prevents late responses from repopulating results.
- Empty and failed search fixtures surface the correct inline states.
- Result filters are client-side only and do not dispatch another `searchJobs` request.
- Run search and Save current pending actions omit query and label from route params.
- Search-tab Vitest specs all pass and privacy red-line assertions remain green.
