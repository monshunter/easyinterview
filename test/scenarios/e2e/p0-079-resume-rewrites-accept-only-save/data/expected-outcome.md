# E2E.P0.079 Expected Outcome

- Non-current `resume-versions` and suggestion decision routes return 404 and are
  absent from generated route catalog.
- `updateResume`, `duplicateResume`, and `requestResumeTailor` fixture parity
  passes.
- Rewrites suggestions are ephemeral UI bullets: accept marks local rows only,
  with no reject/status route.
- Save persists accepted rewrites through `updateResume` (overwrite) or
  `duplicateResume` (save as new).
