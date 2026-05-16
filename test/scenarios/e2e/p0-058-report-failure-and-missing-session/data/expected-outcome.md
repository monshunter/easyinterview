# Expected outcome

- `report.failureState.errorCode.AI_PROVIDER_TIMEOUT` copy renders for the failed status path with retry + back CTAs.
- ReportMissingSessionState renders without invoking `getFeedbackReport`.
- Cross-user 404 routes the dashboard to ReportFailureState with `data-not-found="true"` and the dedicated REPORT_NOT_FOUND copy.
- GeneratingScreen `timeout` state surfaces GeneratingErrorState with the retry CTA available.
- `errorCode` carried on URL params is a B1 enum string; no raw provider error message leaks.
