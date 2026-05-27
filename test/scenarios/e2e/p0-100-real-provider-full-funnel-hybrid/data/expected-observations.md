# Expected Observations

> Owner: `e2e-scenarios-p0/002-manual-uat-real-provider-full-funnel`
> Scenario: `E2E.P0.100`

Use these observations as manual review prompts. They are not fixture outputs and must not be copied into assertions as expected AI text.

## Parse

- The target job title should read as a backend / platform engineering role, not frontend or data-only.
- Core requirements should include distributed systems, Go or equivalent backend language, observability, cloud infrastructure, and reliability ownership.
- Seniority should be senior / staff-adjacent; junior or internship seniority is a failure signal.

## Workspace

- Resume and JD context should both be visible or recoverable before starting practice.
- The practice entry should not require mock seed ids or fixture-only route parameters.
- The page should remain usable in both Chinese and English UI language settings.

## Practice

- The first question should reference backend architecture, reliability, scale, tradeoffs, or incident handling.
- Follow-up should react to the submitted answer sample instead of repeating the same first question.
- The UI should not expose raw prompt text, provider response payloads, or session cookie values.

## Report

- The report should summarize strengths, gaps, evidence, and next actions related to the JD and answer.
- Provider/model evidence should be recorded only as a redacted summary in `.test-output/e2e/p0-100-real-provider-full-funnel-hybrid/evidence.md`.
- The report should not display unrelated legacy modules such as mistakes, growth, drill, standalone voice, or `mode=debrief`.

## Next Round

- Starting the next round should create a distinct practice plan/session tied to the same target context.
- The next round should preserve the training context without copying previous answer text into URL, localStorage, console logs, or screenshots.
