# 001 — Plan and Session Orchestration Checklist

> **版本**: 2.7
> **状态**: active
> **更新日期**: 2026-07-15

**关联计划**: [plan](./plan.md)

## Phase 1: Contract and baseline rebase

- [x] 1.1 RED: add OpenAPI/shared/migration/prompt inventory tests for message conversation and stale-question rejection.
  <!-- verified: 2026-07-12 method=red python3 -m unittest scripts.lint.practice_conversation_contract_test failed on all four pre-change contract surfaces as expected -->
- [x] 1.2 GREEN: replace PracticeTurn/event-answer/question-review schemas with PracticeMessage/sendPracticeMessage/conversation-report schemas; run codegen/fixture gates.
  <!-- verified: 2026-07-12 method=contract make codegen; openapi inventory 20 tests; events inventory 28 tests with 13 current events; lint-openapi; lint-events; validate-fixtures 37 fixtures; git diff --check -->
- [x] 1.3 GREEN: rewrite baseline SQL/enum sources/seeds to 21 app tables with `practice_messages`, 11 shared enums and 6 prompt coordinates; run migration/prompt/rubric/profile/eval gates.
## Phase 2: PracticePlan simplification

- [x] 2.1 RED-GREEN: remove question budget/mode/hints across request, domain, store and fixtures; preserve baseline and derived-plan validation.
- [x] 2.2 RED-GREEN: update frontend start helpers and real-mode contract tests to send only current plan fields.

## Phase 3: Session start with opening message

- [x] 3.1 RED: starter tests require `practice.session.chat` and one opening assistant message, with no currentTurn/first_question.
- [x] 3.2 GREEN: implement three-stage start reservation/generation/persistence and one-repair language/schema validation.
- [x] 3.3 RED-GREEN: timeout/config/invalid-output/idempotency replay produce no duplicate session/message/outbox or canned text.

## Phase 4: Session read model

- [x] 4.1 RED-GREEN: get/list session store/API return ordered messages and no turn fields.
- [x] 4.2 RED-GREEN: empty/cross-user/missing/deleted context tests pass.

## Phase 5: Privacy and closeout

- [x] 5.1 RED-GREEN: raw message redaction tests cover event/outbox/audit/log/metric/task-run surfaces.
  <!-- verified: 2026-07-12 method=privacy-observability-scenario evidence="lifecycle-only outbox, plaintext redaction, metric allowlist and one conversation-level report call pass" -->
- [x] 5.3 仓库根 `make test` 完成前后端全量单测回归；codegen、fixture、migration、prompt/eval、context/docs/index 与 diff 作为独立 gates。

## Phase 6: Complete resume grounding for session start

- [x] 6.1 RED: start store/service tests require `parsed_text_snapshot → original_text → structured_profile` precedence, full long-input tail marker in AI payload, and zero AI call when all resume content is empty.<!-- verified: 2026-07-12 method=go-test-red tests=TestSQLRepositoryReserveSessionStartPrefersCompleteResumeSourceSnapshot,TestStartPracticeSessionFailsClosedWithoutResumeContextAndSkipsAI -->
- [x] 6.2 GREEN: start reservation exposes one complete `ResumeContext`, removes the `resume context unavailable` fallback, and returns typed `VALIDATION_FAILED` before prompt resolve/AI when empty.<!-- verified: 2026-07-12 method=go-test packages=internal/practice,internal/store/practice -->
- [x] 6.3 RED-GREEN: prompt lint/eval requires persisted resume or candidate-authored `user` evidence, forbids invented project/company/technology claims, treats `assistant` history as continuity-only, and asks for clarification when project names are absent; sync hash/seed/resolved artifacts.<!-- verified: 2026-07-12 method=unittest+prompt-lint+eval-offline result=27/27-pass cases="real-resume-grounding-no-invented-project,assistant-hallucination-is-not-candidate-fact" -->
- [x] 6.3a RED-GREEN: payload tests and prompt lint/eval prove immutable policy is a `system` message; JD/resume/round/persona/history are JSON-encoded untrusted user data; injection-like tags cannot escape into policy; persona controls style only.<!-- verified: 2026-07-12 method=go+pytest+eval tests="TestPracticeChatPayloadDoesNotLetLanguageBreakSystemPolicyBoundary,TestStartPracticeSessionCreatesOpeningAssistantMessage,TestBackendPracticeConversationPromptPreflight" gates="registry preflight,practice conversation contract,27-case offline eval" -->

## Phase 7: Persisted canonical round identity and plan selection

- [x] 7.1 RED: generated/domain/API tests require request `roundId`, paired plan response identity, and no request `roundSequence`.<!-- verified: 2026-07-12 method=generated+handler+domain-tests -->
- [x] 7.2 RED-GREEN: create-plan SQL/domain atomically derives canonical round pair and validates baseline/retry/next against distinct completed session facts.<!-- verified: 2026-07-12 method=real-postgres-integration test=TestSQLRepositoryIntegration_CreatePlanProjectsCanonicalRoundLedger -->
- [x] 7.2a RED-GREEN: start reservation resolves exact round name/type/focus from the persisted pair; prompt round context is not `interviewerPersona`, and legacy/mismatch identity cannot start a new session.<!-- verified: 2026-07-12 method=unit+real-postgres marker=canonical-round-prompt-context=PASS -->
- [x] 7.2b RED-GREEN: request resume, source plan/report and completion facts must match `target_jobs.resume_id`; missing provenance and same-user wrong-resume inputs fail closed.<!-- verified: 2026-07-12 method=real-postgres markers="target-resume-binding-and-provenance" -->
- [x] 7.3 RED-GREEN: canonical summary requires non-empty provenance, positive int32 strictly increasing/unique sequence and lowercase allowlisted type; `1,2,4` is valid and selects existing successor `4`; overflow/case drift, same-duration ambiguity, mismatched round/budget, all-complete, missing/legacy source identity and IK mismatch fail closed without inserting a plan.<!-- verified: 2026-07-12 method=unit+real-postgres markers="canonical-round-type-case-sensitive,non-contiguous-successor,equal-duration-next-round,stale-source-and-round-budget-mismatch,all-rounds-complete-fail-closed" -->
- [x] 7.5 Repository-root `make test` provides frontend/backend unit regression；OpenAPI/generated, migration, `DATABASE_URL` integration, context/docs/index/diff and business-persistence negative checks remain separate gates.

## Phase 8: Remove public session listing

- [ ] 8.1 RED: OpenAPI inventory/generated/router/handler/fixture/mock/source tests fail while `GET /practice/sessions` / `listPracticeSessions` remains a current positive surface.
- [ ] 8.2 GREEN: remove list operation, generated method/server interface, mux route, handler/service/store path, fixture and mock registry entry without redirect/deprecated/empty compatibility.
- [ ] 8.3 PRESERVATION-GATE: `POST /practice/sessions` start and `GET /practice/sessions/{sessionId}` live recovery remain generated, routed, fixture-backed and behavior-tested.
- [ ] 8.4 HANDOFF-GATE: completed transcript is reachable only through backend-review `getReportConversation(reportId)` over the existing relation; Workspace/Reports/Practice have no session-list consumer and no migration/table is introduced.
- [ ] 8.5 BDD-N/A: no current user-visible session-list flow exists; substitute gates are OpenAPI exact inventory/diff, fixture/codegen/mock parity, focused handler/source negatives and root `make test`.
- [ ] 8.6 COMPLETION-GATE: root `make test`, build, OpenAPI/fixture/codegen/mock, docs/context/index/diff and scoped zero-reference gates pass before restoring completed status.

## 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-15 | 2.7 | 新增 Phase 8：删除 listPracticeSessions 全部正向 surface，保留 start/get live operations，并交接 report-owned conversation read。 |
| 2026-07-12 | 2.5 | 补齐 assistant history 不得成为候选人事实的 RED/GREEN prompt 与负向 eval gate。 |
| 2026-07-12 | 2.4 | 补齐 TargetJob 绑定 resume/provenance/type/int32 目录约束，并增加 system policy 与 JSON 不可信上下文分层 gate。 |
| 2026-07-12 | 2.3 | 原地重开 Phase 7，按方案 A 持久化规范化轮次身份并由完成台账校验当前/复练/下一轮。 |
| 2026-07-12 | 2.2 | 原地重开 Phase 6，修复 start 只读空 structured_profile 导致真实简历丢失和无证据提问。 |

## BDD Gate

- [x] BDD-Gate: `BDD.PRACTICE.PLAN.001` 由 [BDD checklist](./bdd-checklist.md) 关联现有 plan/session owner behavior tests；不创建或声明真实 E2E PASS。
- [ ] BDD-Gate Phase 8: `listPracticeSessions` 无当前用户行为流，不新增 BDD/E2E；按 [BDD checklist](./bdd-checklist.md) 回归保留的 start/get 行为并执行代码层替代 gate。
