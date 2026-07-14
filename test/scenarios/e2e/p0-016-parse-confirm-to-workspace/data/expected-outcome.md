# Expected Outcome

- Parse displays one `parse-reports-entry` in the plan-detail content header and reaches `/reports?targetJobId=<uuid>`.
- Parse does not call `listTargetJobReports`, does not embed `parse-reports` / round rows, and does not accept `section=reports`.
- Formal TopBar and prototype TopBar each keep exactly three primary entries and contain no report navigation item.
- Formal/prototype entry text, computed styles, absolute bounding boxes and screenshots match at 1440x900 and 390x844 within changed-pixel ratio ≤0.5%.
- The readonly detail continues to show saved backend rounds and the bound resume without edit/picker controls.
- Start still reaches `/practice` with trusted target/resume/plan/session context and makes no `updateTargetJob` request.
- Source, Vitest, frontend build and six focused Playwright project cases all report PASS with no no-test/failure marker.
