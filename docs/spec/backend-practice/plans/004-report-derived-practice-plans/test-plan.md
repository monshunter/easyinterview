# 004 — Report-derived Practice Plans Test Plan

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-07-06

**关联计划**: [plan](./plan.md) / [checklist](./checklist.md)
**关联 Test Checklist**: [test-checklist](./test-checklist.md)

## 1 测试矩阵

| 行 | 风险 | 测试文件 / 入口 | 预期证据 |
|----|------|----------------|----------|
| T1 | `sourceReportId` contract drift | `openapi/openapi.yaml`, generated Go/TS, `backend/internal/api/practice/*_test.go` | retry / next-round request and response expose `sourceReportId` |
| T2 | report source validation / isolation drift | `backend/internal/practice/create_plan_test.go`, `backend/internal/store/practice/create_plan_test.go`, `backend/cmd/api/practice_http_scenario_test.go` | missing/cross-user/wrong-target source rejected without leaking existence |
| T3 | prohibited source fields return | negative grep over OpenAPI, generated artifacts, fixtures, backend runtime, scenario docs | no current `sourceDebriefId`, `source_debrief_id`, `PracticeGoalDebrief`, `goal='debrief'`, or source-question bypass scenario |
| T4 | stale BDD no-op gate | `test/scenarios/e2e/p0-070-*`, `test/scenarios/e2e/p0-072-*`, backend test search | P0.070 and P0.072 scripts target existing tests |

## 2 Current Verification Commands

```bash
cd backend && go test ./internal/practice ./internal/store/practice ./internal/api/practice ./cmd/api -run 'Derived|Source|TestE2EP0070|TestE2EP0072' -count=1
```

```bash
rg -n "sourceDebriefId|source_debrief_id|PracticeGoalDebrief|goal='debrief'|goal=debrief|debrief-derived|raw_questions|__DEBRIEF" \
  backend openapi frontend/src/api/generated backend/internal/api/generated openapi/fixtures test/scenarios/e2e \
  -g '!**/*_test.go'
```

Expected current hits are limited to negative text in scenario verifiers or product-scope gates, not positive runtime/generated/fixture contract.

```bash
python3 .agent-skills/implement/shared/scripts/validate_context.py \
  --context docs/spec/backend-practice/plans/004-report-derived-practice-plans/context.yaml \
  --docs-root docs \
  --target backend
make docs-check
git diff --check
```

## 3 Scenario Gate Contract

- `E2E.P0.070`: `TestE2EP0070PracticeDerivedPlanCreateReadReplay` proves report-derived `retry_current_round` / `next_round` create/read/replay.
- `E2E.P0.072`: `TestE2EP0072PracticeDerivedSourceValidationIsolationPrivacy` proves missing/cross-user/wrong-target source validation and privacy.
