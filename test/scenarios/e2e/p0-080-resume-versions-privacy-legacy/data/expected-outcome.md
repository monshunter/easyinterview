# E2E.P0.080 Expected Outcome

- `TestOutboxPrivacyForTailorCompletedEvent` proves `resume.tailor.completed` payloads expose only `tailorRunId`, `resumeId`, `targetJobId`, `mode`, and `status`.
- `TestAiTaskRunsPrivacyForTailorDrainer` proves typed `ai_task_runs` rows do not persist raw prompt, raw model response, match summary, or suggested bullet content.
- `TestAuditPrivacyForTailorDrainer` proves tailor drainer audit metadata does not persist prompt or response bodies.
- Live store and cmd/api drainer tests pass for ready and failure paths.
- Negative grep checks report zero retired `inline` / `rewrite` / `mirror` and Mistakes / Growth / Drill vocabulary matches in backend resume runtime code.
