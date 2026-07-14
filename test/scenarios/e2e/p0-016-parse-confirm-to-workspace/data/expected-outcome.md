# Expected Outcome

- Parse renders only queued/processing progress and history-replaces ready to `/workspace?targetJobId=<uuid>`; ready cards and direct detail links never show Parse animation.
- Workspace detail displays one `parse-reports-entry` DOM anchor in the shared content header and reaches `/reports?targetJobId=<uuid>`.
- Workspace detail does not call `listTargetJobReports`, does not embed `parse-reports` / report rows, and neither Parse nor Workspace accepts `section=reports`.
- Formal TopBar and prototype TopBar each keep exactly three primary entries and contain no report navigation item.
- Formal/prototype entry and done/current/pending round-state text, computed styles, absolute bounding boxes and screenshots match at 1440x900 and 390x844; all three round states keep distinct backgrounds and borders under default, dark and custom-accent themes.
- The readonly detail continues to show saved backend rounds and the bound resume without edit/picker controls.
- Start still reaches `/practice` with trusted target/resume/plan/session context and makes no `updateTargetJob` request.
- Source, Vitest, frontend build and six focused Playwright project cases all report PASS with no no-test/failure marker.
