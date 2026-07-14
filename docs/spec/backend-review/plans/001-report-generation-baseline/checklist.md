# 001 — Grounded Conversation Report Generation Checklist

> **版本**: 2.24
> **状态**: active
> **更新日期**: 2026-07-15

**关联计划**: [plan](./plan.md)

## Phase 1-5: Conversation-level baseline（历史已完成）

- [x] Conversation-level contract、generate/read/replay、privacy baseline and numeric-score removal completed.

## Phase 6: Frozen context and direct contract

- [x] Consume OpenAPI、migration、completion snapshot and prompt-registry owner contracts; review reads only frozen context and terminal messages.
- [x] Persist and expose the direct report shape losslessly while stripping private anchors from API.
- [x] Delete committed `input-*.json`; do not reconstruct default-sized boundary material.

## Phase 7: Validator and action-local recovery

- [x] Focused owner tests cover schema/wire、24/64 language、focus/action and fail-closed invariants.
- [x] Focused owner tests cover invocation-local initial+3、10s/20s/40s、dynamic full revalidation、cancellation and independent invocation reset.
- [x] Store/runner integration tests preserve stale-worker fencing while keeping async attempts infrastructure-only.
- [x] Removed identifiers and hidden score paths have zero positive active-code hits.

## Phase 8: Reliability and real UI separation

- [x] Prompt/eval owners validate generation/judge reliability, typed retry/content rejection and redacted evidence as independent code/eval gates.
- [x] BDD-Gate: `BDD.REPORT.GENERATE.001` 由 [BDD checklist](./bdd-checklist.md) 关联 frozen-context generation/repair/persistence/replay owner behavior tests。
- [x] E2E-HANDOFF: P0.099 是唯一 real report/generating frontend/backend/provider/API/DB + exact-six visual owner；本轮未运行，状态仍为 `Ready`。
- [x] Provider/eval output is not an E2E scenario and is not a P0.099 prerequisite.

## Phase 9: Persistence and privacy closeout

- [x] Ready/failure persistence keeps jobID + claimed attempts fencing and zero stale-worker domain side effects.
- [x] Report/job/outbox/audit/log/metric surfaces retain no raw prompt/output or complete candidate context.
- [x] Development focused tests and required race/PostgreSQL tests pass; phase completion uses root `make test` for whole backend/frontend unit regression.

## Phase 10: Canonical-round report overview

- [x] Minimal closed wire、canonical order、independent current/latest selection and fail-closed identity/context tests pass.
- [x] ReportsScreen is the only list consumer; Parse/Report/Generating have zero list calls and no global/history center is introduced.
- [ ] Generated/fixture handoff、root `make test` and scoped stale pagination/pointer negative search close the phase.

## Phase 11: Injected report input guard

- [x] A4 owns one typed default/override/invalid/cross-field contract suite.
- [x] Review uses one small injected admitted/overflow provider call/no-call test；the historical 62,397-byte symptom is not reconstructed and no default-size material is created.
- [x] A3 loader/coverage gates require all six active profiles `max_tokens >= 16384` and keep report context at 1,000,000; no byte/token capacity formula is used.
- [x] BDD-N/A: configuration wiring does not create a user workflow; no scenario or real large material is used.

## Phase 12: Report-owned conversation read

- [ ] 12.1 RED: store/handler tests require owned report lookup, existing unique session relation, strict `seq_no ASC`, four report statuses and closed message projection.
- [ ] 12.2 FAILURE/PRIVACY-GATE: missing/cross-user hidden 404；report/session/user/target mismatch、empty/duplicate/non-increasing sequence、unknown role/additional locator fail closed with no partial transcript or raw log/audit/metric body.
- [ ] 12.3 GREEN: implement generated `getReportConversation` handler/store with zero AI/write/pagination/new table; do not call `getPracticeSession` or reorder corruption into apparent success.
- [ ] 12.4 REMOVAL-GATE: current OpenAPI/generated/router/handler/fixture/mock/frontend positive surface has zero `listPracticeSessions`; accepted history/decision and exact negative declarations are classified, not blanket-excluded.
- [ ] 12.5 BDD-Gate: `BDD.REPORT.CONVERSATION.API.001` passes owner tests; E2E.P0.099 receives real API/DB binding handoff without changing exact-six screenshots.
- [ ] 12.6 COMPLETION-GATE: focused Go tests, root `make test`, OpenAPI/fixture/codegen/mock, docs/context/index/diff and migration-zero-change audit pass.

## Closeout

- [x] Root `make test` is the independent complete frontend/backend unit regression gate; code tests are never wrapped as E2E.
- [x] P0.099、docs/index/context/diff and deleted-scenario negative checks are reported separately.
- [ ] BDD-Gate: 在当前真实环境显式运行 `E2E.P0.099` 并完成 exact-six no-OCR audit；本轮未执行。
