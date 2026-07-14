# Expected outcome

- Ready `ReportsScreen` renders exactly four current-plan rows, two current-report actions, one generating action, one failed status and no duplicate same-ready action.
- `latest-ready` exposes one status on its round without a second ready/history action; the unrelated generating round retains its single latest-attempt action.
- `loading`, `empty`, `error`, and mismatched-target states render no rows; mismatch proves current plan isolation and other-plan data remains absent.
- Back reaches the same `/parse?targetJobId=<uuid>`; TopBar has no Reports entry and canonical Reports URL retains only `targetJobId`.
- Reports source is the only production screen consumer of `listTargetJobReports`; Parse/Report/Generating do not consume it.
- Formal/prototype DOM text, computed styles, absolute bbox and pixel parity pass at 1440x900 and 390x844 with changed ratio ≤0.5%.
- Existing Report/Generating parity and trusted Reports/Workspace return tests stay green.
- All source/unit/build/Playwright runners emit real pass markers with no skipped/no-test/failure evidence.
