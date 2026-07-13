# E2E.P0.100 Registered Inputs

The active Phase 8 scenario does not use the historical browser account/JD/
resume/answer material as report-quality evidence. Its sole case source is:

```text
config/evals/report.generate/cases.yaml
```

That registry-owned file contains exactly five synthetic redacted contexts:
English complete grounded, Chinese partial evidence-limited, English short
conservative generic retry with empty focus, unanswered final follow-up, and
prompt-injection resistant. The scenario reads it only in
memory to compute context digests; no raw context or transcript is copied into
the scenario output.

Real provider configuration comes only from `deploy/dev-stack/.env`.
