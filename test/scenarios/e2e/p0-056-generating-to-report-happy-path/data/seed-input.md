# Seed input

- Existing `.test-output/e2e/p0-047-practice-text-loop-complete-and-generating-handoff/completion-backend-evidence.json` with exact `practice-completion-evidence.v1` keys, three PASS tests and the three completion owner markers.
- Isolated queued/generating/ready report rows created inside `TestE2EP0056ReportBackendEvidence`; the test receives the owner artifact path through `PRACTICE_COMPLETION_EVIDENCE_PATH`.
- Direct report fixtures with frozen context, summary, dimension assessments, evidence, actions and legal empty/non-empty retry-focus shapes.
- Redacted scenario correlation only; no raw context or transcript is copied into scenario output.
