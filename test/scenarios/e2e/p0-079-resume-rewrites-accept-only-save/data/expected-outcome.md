# E2E.P0.079 Expected Outcome

- Non-current `resume-versions` and suggestion decision routes return 404 and are
  absent from generated route catalog.
- `updateResume`, `duplicateResume`, and `requestResumeTailor` fixture parity
  passes.
- Resume detail is read-only: no Rewrites/Edit tab, export, copy, original modal,
  tailor request, or detail save call is exposed.
- Legacy Rewrites/Edit route params do not materialize removed frontend surfaces.
