# Expected outcome

- `report.*` and `generating.*` key sets are byte-identical across zh and en.
- AI_* enum errorCodes each have a dedicated i18n string distinct from UNKNOWN.
- REPORT_NOT_FOUND has independent failureState.notFound.{eyebrow,title,desc} copy.
- Implementation files under report/ and generating/ contain none of the retired vocabulary terms; the Python lint script and TypeScript negative-grep both pass.
- Frontend build succeeds before visual parity execution.
- Playwright pixel-parity specs `generating.spec.ts` + `report.spec.ts` execute and pass for desktop and mobile viewports.
