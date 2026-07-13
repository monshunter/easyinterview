# Expected outcome

- `report.*` and `generating.*` UI key sets stay exact across zh/en; typed backend enums never render as raw user copy.
- Prototype and formal pages use identical fixtures and deterministic locale/timezone/Date/DPR/font/motion controls at 1440×900 and 390×844.
- Normalized DOM text, selected computed styles, viewport-relative absolute bbox and responsive column/scroll-width assertions pass for zh needs-practice, en well-prepared and generating.
- `pixelmatch` runs with threshold 0.1; every compared root has changed-pixel ratio ≤0.5%.
- On parity failure, prototype/formal/diff images are attached; on success, no diagnostic image set is retained.
- Frontend build, owner preflight, both browser specs, i18n gate and stale-contract negatives all pass with no skipped/no-test/failure marker.
