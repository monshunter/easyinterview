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
| T2 | report source validation / isolation drift | practice service/store/API code tests | missing/cross-user/wrong-target source rejected without leaking existence |
| T3 | prohibited source fields return | negative grep over OpenAPI, generated artifacts, fixtures and backend runtime | no current `sourceDebriefId`, `source_debrief_id`, `PracticeGoalDebrief`, `goal='debrief'`, or source-question bypass |
| T4 | client focus authority | OpenAPI/generated/API request negatives | no focus input; report-local dimension focus is server-projected only |
| T5 | source-derived settings | service/store real PostgreSQL matrix | derived request has no copied settings; server reuses source plan and successor duration/identity |
| T8 | legacy focus identifiers | final active runtime/generated/OpenAPI/fixtures negative gate | old competency/turn/action names and code-only prompt focus have zero positive consumers; immutable v0.1/000002 rollback and explicit negative fixtures are allowlisted |

## 2 Verification policy

Focused practice/store/API commands may be used for development feedback. Phase completion is reported only by repository-root `make test`; database integration and contract negative searches remain separate gates.

```bash
rg -n "sourceDebriefId|source_debrief_id|PracticeGoalDebrief|goal='debrief'|goal=debrief|debrief-derived|raw_questions|__DEBRIEF" \
  backend openapi frontend/src/api/generated backend/internal/api/generated openapi/fixtures \
  -g '!**/*_test.go'
```

Expected current hits are limited to negative contract text, not positive runtime/generated/fixture behavior.

The same allowlisted-negative policy applies to `focusCompetencyCodes|focus_competency_codes|retryFocusCompetencyCodes|retry_focus_competency_codes|retryFocusTurnIds|retry_focus_turn_ids|retry_round`; current positive contract uses only `retryFocusDimensionCodes` / `focus_dimension_codes` and `semanticFocus/{{semantic_focus_json}}`. Immutable v0.1/000002 rollback and explicit negative fixtures may retain old prompt literals, but F3's final active v0.2 and backend runtime may not consume them. The gate emits `REPORT_DERIVED_LEGACY_IDENTIFIER_NEGATIVE_PASS` only after F3 emits `PRACTICE_SEMANTIC_FOCUS_PROMPT_V020_PASS`.

```bash
python3 .agent-skills/implement/shared/scripts/validate_context.py \
  --context docs/spec/backend-practice/plans/004-report-derived-practice-plans/context.yaml \
  --docs-root docs \
  --target backend
make docs-check
git diff --check
```
