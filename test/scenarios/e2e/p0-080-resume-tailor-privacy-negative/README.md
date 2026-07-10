# E2E.P0.080 resume tailor privacy negative

## 1. Purpose

Validate the backend-resume D-20 privacy and out-of-scope vocabulary regression gate
after flat resume tailor async jobs, task runs, and outbox events are wired.

## 2. Requirements

- backend-resume C-13
- `docs/spec/backend-resume/plans/002-tailor-runs-and-save-v1/bdd-plan.md`

## 3. Given / When / Then

Given E2E.P0.074 through E2E.P0.079 have already covered the live API and
persistence paths for flat resume reads, update, duplicate, tailor dispatch,
ready/failure drainer paths, and out-of-scope suggestion decisions.

When the privacy regression runner replays the focused tailor drainer, live store, and cmd/api gates and performs zero-reference searches over `backend/internal/resume/`.

Then completed outbox payloads contain only IDs, mode, and status; `ai_task_runs` and audit metadata do not contain prompt bodies, raw model responses, resume/JD text, match summaries, or suggested bullet values; and out-of-scope `inline` / `rewrite` / `mirror` plus Mistakes / Growth / Drill vocabulary remains absent from backend resume runtime code.

## 4. Scripts

- `scripts/setup.sh`: prepares output directories and copies seed / expected outcome notes into `.test-output`.
- `scripts/trigger.sh`: runs privacy-focused Go gates, live store outbox assertions, cmd/api drainer scenarios, and out-of-scope vocabulary negative searches.
- `scripts/verify.sh`: rejects skipped or no-op gates, checks required runner markers and PASS evidence, and repeats privacy / out-of-scope vocabulary negative searches.
- `scripts/cleanup.sh`: records cleanup completion while preserving logs under `.test-output/`.

## 5. Evidence

Scenario evidence is written to `.test-output/e2e/p0-080-resume-tailor-privacy-negative/`:

- `setup.log`
- `trigger.log`
- `verify.log`
- `cleanup.log`
- `seed-input.md`
- `expected-outcome.md`

## 6. Isolation

- Environment: shared local scenario environment.
- Parallel safe: No.
- Cleanup is idempotent and preserves evidence logs.

## 7. Offline Limits

This scenario is a regression and privacy negative gate. It depends on focused backend tests and live store integration assertions rather than a deployed frontend, because frontend real-backend consumption is owned by the follow-on `frontend-resume-workshop` plan.
