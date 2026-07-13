# 004 — Report-derived Practice Plans Test Plan

> **版本**: 1.8
> **状态**: completed
> **更新日期**: 2026-07-12

**关联计划**: [plan](./plan.md) / [checklist](./checklist.md)
**关联 Test Checklist**: [test-checklist](./test-checklist.md)

## 1 测试矩阵

| 行 | 风险 | 测试文件 / 入口 | 预期证据 |
|----|------|----------------|----------|
| T1 | `sourceReportId` contract drift | `openapi/openapi.yaml`, generated Go/TS, `backend/internal/api/practice/*_test.go` | retry / next-round request and response expose `sourceReportId` |
| T2 | report source validation / isolation drift | `backend/internal/practice/create_plan_test.go`, `backend/internal/store/practice/create_plan_test.go`, `backend/cmd/api/practice_http_scenario_test.go` | missing/cross-user/wrong-target source rejected without leaking existence |
| T3 | prohibited source fields return | negative grep over OpenAPI, generated artifacts, fixtures, backend runtime, scenario docs | no current `sourceDebriefId`, `source_debrief_id`, `PracticeGoalDebrief`, `goal='debrief'`, or source-question bypass scenario |
| T4 | client focus authority | OpenAPI/generated/API request negatives | no focus input; report-local dimension focus is server-projected only |
| T5 | source-derived settings | service/store real PostgreSQL matrix | derived request has no copied settings; server reuses source plan and successor duration/identity |
| T6 | retry/next semantics | service/store/API + F3 v0.2 exact-version/active pair + P0.070 | retry accepts empty generic focus or projects issue-backed non-empty codes and resolves code→label+issues into `semanticFocus`; F3 v0.2 uses `{{semantic_focus_json}}`; next successor has empty focus; IK replay exact |
| T7 | stale BDD no-op gate | `test/scenarios/e2e/p0-070-*`, `test/scenarios/e2e/p0-072-*`, backend test search | P0.070 and P0.072 scripts target existing tests |
| T8 | legacy focus identifiers | final active runtime/generated/OpenAPI/fixtures/scenario negative gate | old competency/turn/action names and code-only prompt focus have zero positive consumers; immutable v0.1/000002 rollback and explicit negative fixtures are allowlisted |

## 2 Current Verification Commands

```bash
cd backend && go test ./internal/practice ./internal/store/practice ./internal/api/practice ./cmd/api -run 'Derived|Source|TestE2EP0070|TestE2EP0072' -count=1
cd backend && go test ./internal/practice -run '^TestPracticeChatV020CandidateUsesSemanticFocus$' -count=1
```

```bash
rg -n "sourceDebriefId|source_debrief_id|PracticeGoalDebrief|goal='debrief'|goal=debrief|debrief-derived|raw_questions|__DEBRIEF" \
  backend openapi frontend/src/api/generated backend/internal/api/generated openapi/fixtures test/scenarios/e2e \
  -g '!**/*_test.go'
```

Expected current hits are limited to negative text in scenario verifiers or product-scope gates, not positive runtime/generated/fixture contract.

The same allowlisted-negative policy applies to `focusCompetencyCodes|focus_competency_codes|retryFocusCompetencyCodes|retry_focus_competency_codes|retryFocusTurnIds|retry_focus_turn_ids|retry_round`; current positive contract uses only `retryFocusDimensionCodes` / `focus_dimension_codes` and `semanticFocus/{{semantic_focus_json}}`. Immutable v0.1/000002 rollback and explicit negative fixtures may retain old prompt literals, but F3's final active v0.2 and backend runtime may not consume them. The gate emits `REPORT_DERIVED_LEGACY_IDENTIFIER_NEGATIVE_PASS` only after F3 emits `PRACTICE_SEMANTIC_FOCUS_PROMPT_V020_PASS`.

```bash
python3 .agent-skills/implement/shared/scripts/validate_context.py \
  --context docs/spec/backend-practice/plans/004-report-derived-practice-plans/context.yaml \
  --docs-root docs \
  --target backend
make docs-check
git diff --check
```

## 3 Scenario Gate Contract

- `E2E.P0.070`: `TestE2EP0070PracticeDerivedPlanCreateReadReplay` proves report-derived generic-empty-focus and issue-backed-focus retry plus `next_round` create/read/replay.
- `E2E.P0.072`: `TestE2EP0072PracticeDerivedSourceValidationIsolationPrivacy` proves missing/cross-user/wrong-target source validation and privacy.
