# Expected outcome

- `report.*` and `generating.*` key sets are byte-identical across zh and en.
- AI_* enum errorCodes each have a dedicated i18n string distinct from UNKNOWN.
- REPORT_NOT_FOUND has independent failureState.notFound.{eyebrow,title,desc} copy.
- Implementation files under report/ and generating/ contain none of the out-of-scope vocabulary terms; the Python lint script and TypeScript negative-grep both pass.
- Frontend build succeeds before visual parity execution.
- The owner/browser preflight checks the active spec, six plan artifacts, browser sources and scenario claims; it rejects claims not backed by the P0.059 runner and confirms both Playwright specs execute real screenshot calls.
- Playwright `generating.spec.ts` passes desktop main, missing-report and mobile-overflow states; `report.spec.ts` passes desktop dashboard, missing-session, failed and mobile-overflow states.
- Every covered browser state produces a non-empty in-memory screenshot; no image-comparison files are written.
