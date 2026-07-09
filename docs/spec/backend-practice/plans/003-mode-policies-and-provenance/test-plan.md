# 003 Test Plan

> **版本**: 1.5
> **状态**: completed
> **更新日期**: 2026-07-09

**关联 Plan**: [plan](./plan.md)
**关联 Test Checklist**: [test-checklist](./test-checklist.md)

## 1 测试矩阵

| Test area | Required assertions | Representative tests / commands |
|-----------|---------------------|----------------------------------|
| Mode dispatch | assisted and legacy strict produce `show_hint` AI path; goal is orthogonal | `cd backend && go test ./internal/practice -run 'Test.*Hint.*LegacyStrict|TestServiceAppliesHintAIForAssisted' -count=1` |
| Assisted hint persistence | success writes `practice_turns.hint_text`; replay returns original hint snapshot; no turn/outbox/audit side effects | `cd backend && go test ./internal/store/practice ./internal/practice -run 'Test.*Hint.*' -count=1` |
| AI/task-run contract | `hint_generate` is accepted by migration/A3 writer; F3/A3 success and failure write typed task-run evidence | `cd backend && go test ./internal/ai/aiclient ./internal/ai/registry ./internal/migrations -count=1` |
| Provenance wire shape | AssistantAction provenance JSON has exactly six B2 fields and no runtime fields | `cd backend && go test ./internal/practice -run 'TestAssistantActionProvenance' -count=1` |
| Graceful degrade | F3/A3/parser failures return session_wait, keep session running, and do not leak details | `cd backend && go test ./internal/practice ./cmd/api -run 'TestApplyHintAIGracefulDegradeMatrix|TestE2EP0051' -count=1` |
| Privacy / redaction | no question/answer/hint/prompt/response/secret in logs, metrics, audit, events, typed task-run payloads | `cd backend && go test ./internal/practice ./cmd/api -run 'Privacy|Redact|TestE2EP0051' -count=1` |
| Runtime boundary lint | removed mode/goal/route vocabulary absent from runtime/generated/scenario surfaces except explicit negative gates | `python3 scripts/lint/backend_practice_non_current.py --repo-root . --phase all` |
| BDD HTTP scenarios | P0.048-P0.051 all pass through `cmd/api` route/middleware/service/store harness | `cd backend && go test ./cmd/api -run 'TestE2EP0048|TestE2EP0049|TestE2EP0050|TestE2EP0051' -count=1` |

## 2 Closeout Gates

- `make validate-fixtures`
- `make codegen-check`
- `python3 scripts/lint/backend_practice_non_current.py --repo-root . --phase all`
- `python3 -m pytest scripts/lint/backend_practice_non_current_test.py -q`
- `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/backend-practice/plans/003-mode-policies-and-provenance/context.yaml --target backend`
- `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
- `make docs-check`
- `git diff --check`

## 3 Test Ownership

- Unit/service/store/API tests live under `backend/internal/practice`, `backend/internal/store/practice`, `backend/internal/api/practice` and `backend/cmd/api`.
- Contract and task-run support tests live under `backend/internal/migrations`, `backend/internal/ai/aiclient` and `backend/internal/ai/registry`.
- Boundary lint lives under `scripts/lint/backend_practice_non_current.py` and its pytest coverage.
- Scenario-level behavior for this plan is Go HTTP scenario coverage, not separate shell scenario directories.
