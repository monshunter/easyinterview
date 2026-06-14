# Expected Outcome

- count defaults to 6 and rejects values outside 1..10
- sessionId loads same-user completed practice session summary for the selected target job and injects mock-report signals into the AI prompt
- cross-user or wrong-target sessionId is rejected as a debrief prerequisite failure
- real prompt markers such as `{{mock_report_summary}}` are replaced before the AI call
- success response includes suggested questions with source
- AI failures write failed/timeout task run rows
- audit metadata contains only ids, language and suggestion_count
