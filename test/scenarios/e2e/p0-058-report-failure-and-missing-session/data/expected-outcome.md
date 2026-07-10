# Expected outcome

- `report.failureState.errorCode.AI_PROVIDER_TIMEOUT` copy renders for the failed status path with retry + back CTAs.
- ReportMissingSessionState renders without invoking `getFeedbackReport`.
- Cross-user 404 routes the dashboard to ReportFailureState with `data-not-found="true"` and the dedicated REPORT_NOT_FOUND copy.
- The poll hook reaches `timeout` after max attempts and maps HTTP 404 to `REPORT_NOT_FOUND`; this runner does not claim GeneratingScreen UI evidence.
- All six focused file markers, the real-mode bootstrap contract and typed i18n keys pass.
