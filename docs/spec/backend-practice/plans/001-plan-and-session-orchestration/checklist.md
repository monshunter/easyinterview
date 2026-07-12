# 001 — Plan and Session Orchestration Checklist

> **版本**: 2.5
> **状态**: completed
> **更新日期**: 2026-07-12

**关联计划**: [plan](./plan.md)

## Phase 1: Contract and baseline rebase

- [x] 1.1 RED: add OpenAPI/shared/migration/prompt inventory tests for message conversation and stale-question rejection.
  <!-- verified: 2026-07-12 method=red python3 -m unittest scripts.lint.practice_conversation_contract_test failed on all four pre-change contract surfaces as expected -->
- [x] 1.2 GREEN: replace PracticeTurn/event-answer/question-review schemas with PracticeMessage/sendPracticeMessage/conversation-report schemas; run codegen/fixture gates.
  <!-- verified: 2026-07-12 method=contract make codegen; openapi inventory 20 tests; events inventory 28 tests with 13 current events; lint-openapi; lint-events; validate-fixtures 37 fixtures; git diff --check -->
- [x] 1.3 GREEN: rewrite baseline SQL/enum sources/seeds to 21 app tables with `practice_messages`, 11 shared enums and 6 prompt coordinates; run migration/prompt/rubric/profile/eval gates.
  <!-- verified: 2026-07-12 method=contract migration lint; migrations/codegen/eval Go tests; 11-enum codegen; 6 prompt/rubric/profile lints; eval-offline 24/24; cross-layer contract 4/4 -->
## Phase 2: PracticePlan simplification

- [x] 2.1 RED-GREEN: remove question budget/mode/hints across request, domain, store and fixtures; preserve baseline and derived-plan validation.
- [x] 2.2 RED-GREEN: update frontend start helpers and real-mode contract tests to send only current plan fields.
- [x] 2.3 BDD-Gate: P0.022/P0.070/P0.072 pass for create/read/derived/isolation.
  <!-- verified: 2026-07-12 method=handler-domain-store-scenarios evidence="P0.022 executes real CreatePracticePlan handler plus domain/SQL gates; P0.070/P0.072 execute derived-source behavior and isolation" -->

## Phase 3: Session start with opening message

- [x] 3.1 RED: starter tests require `practice.session.chat` and one opening assistant message, with no currentTurn/first_question.
- [x] 3.2 GREEN: implement three-stage start reservation/generation/persistence and one-repair language/schema validation.
- [x] 3.3 RED-GREEN: timeout/config/invalid-output/idempotency replay produce no duplicate session/message/outbox or canned text.
- [x] 3.4 BDD-Gate: P0.023/P0.024/P0.025 start, failure and replay scenarios pass.
  <!-- verified: 2026-07-12 method=focused-scenarios evidence="P0.023 opening message, P0.024 AI failure, P0.025 replay/pending retry/mismatch/isolation all execute named tests and reject no-test output" -->

## Phase 4: Session read model

- [x] 4.1 RED-GREEN: get/list session store/API return ordered messages and no turn fields.
- [x] 4.2 RED-GREEN: empty/cross-user/missing/deleted context tests pass.

## Phase 5: Privacy and closeout

- [x] 5.1 RED-GREEN: raw message redaction tests cover event/outbox/audit/log/metric/task-run surfaces.
- [x] 5.2 BDD-Gate: P0.026 privacy/observability passes.
  <!-- verified: 2026-07-12 method=privacy-observability-scenario evidence="lifecycle-only outbox, plaintext redaction, metric allowlist and one conversation-level report call pass" -->
- [x] 5.3 Run focused/full backend, codegen, fixture, migration, prompt/eval, context/docs/index and diff gates.

## Phase 6: Complete resume grounding for session start

- [x] 6.1 RED: start store/service tests require `parsed_text_snapshot → original_text → structured_profile` precedence, full long-input tail marker in AI payload, and zero AI call when all resume content is empty.<!-- verified: 2026-07-12 method=go-test-red tests=TestSQLRepositoryReserveSessionStartPrefersCompleteResumeSourceSnapshot,TestStartPracticeSessionFailsClosedWithoutResumeContextAndSkipsAI -->
- [x] 6.2 GREEN: start reservation exposes one complete `ResumeContext`, removes the `resume context unavailable` fallback, and returns typed `VALIDATION_FAILED` before prompt resolve/AI when empty.<!-- verified: 2026-07-12 method=go-test packages=internal/practice,internal/store/practice -->
- [x] 6.3 RED-GREEN: prompt lint/eval requires persisted resume or candidate-authored `user` evidence, forbids invented project/company/technology claims, treats `assistant` history as continuity-only, and asks for clarification when project names are absent; sync hash/seed/resolved artifacts.<!-- verified: 2026-07-12 method=unittest+prompt-lint+eval-offline result=27/27-pass cases="real-resume-grounding-no-invented-project,assistant-hallucination-is-not-candidate-fact" -->
- [x] 6.3a RED-GREEN: payload tests and prompt lint/eval prove immutable policy is a `system` message; JD/resume/round/persona/history are JSON-encoded untrusted user data; injection-like tags cannot escape into policy; persona controls style only.<!-- verified: 2026-07-12 method=go+pytest+eval tests="TestPracticeChatPayloadDoesNotLetLanguageBreakSystemPolicyBoundary,TestStartPracticeSessionCreatesOpeningAssistantMessage,TestBackendPracticeConversationPromptPreflight" gates="registry preflight,practice conversation contract,27-case offline eval" -->
- [x] 6.4 E2E.P0.023/P0.024 trigger/verify require tail-marker and no-context fail-closed named tests, with no skip/no-op.<!-- verified: 2026-07-12 method=scenario both=PASS -->
- [x] 6.5 BDD-Gate: P0.023/P0.024 pass with complete snapshot grounding and zero-AI empty-context evidence.<!-- verified: 2026-07-12 method=scenario bddChecklist=complete -->

## Phase 7: Persisted canonical round identity and plan selection

- [x] 7.1 RED: generated/domain/API tests require request `roundId`, paired plan response identity, and no request `roundSequence`.<!-- verified: 2026-07-12 method=generated+handler+domain-tests -->
- [x] 7.2 RED-GREEN: create-plan SQL/domain atomically derives canonical round pair and validates baseline/retry/next against distinct completed session facts.<!-- verified: 2026-07-12 method=real-postgres-integration test=TestSQLRepositoryIntegration_CreatePlanProjectsCanonicalRoundLedger -->
- [x] 7.2a RED-GREEN: start reservation resolves exact round name/type/focus from the persisted pair; prompt round context is not `interviewerPersona`, and legacy/mismatch identity cannot start a new session.<!-- verified: 2026-07-12 method=unit+real-postgres marker=canonical-round-prompt-context=PASS -->
- [x] 7.2b RED-GREEN: request resume, source plan/report and completion facts must match `target_jobs.resume_id`; missing provenance and same-user wrong-resume inputs fail closed.<!-- verified: 2026-07-12 method=real-postgres markers="target-resume-binding-and-provenance" -->
- [x] 7.3 RED-GREEN: canonical summary requires non-empty provenance, positive int32 strictly increasing/unique sequence and lowercase allowlisted type; `1,2,4` is valid and selects existing successor `4`; overflow/case drift, same-duration ambiguity, mismatched round/budget, all-complete, missing/legacy source identity and IK mismatch fail closed without inserting a plan.<!-- verified: 2026-07-12 method=unit+real-postgres markers="canonical-round-type-case-sensitive,non-contiguous-successor,equal-duration-next-round,stale-source-and-round-budget-mismatch,all-rounds-complete-fail-closed" -->
- [x] 7.4 BDD-Gate: P0.022/P0.070/P0.072 execute round identity create/read/replay/source validation against real PostgreSQL, not sqlmock-only evidence.<!-- verified: 2026-07-12 method=scenario-run result=PASS -->
- [x] 7.5 Run OpenAPI/generated, migration, focused/full backend, `DATABASE_URL` integration, context/docs/index/diff and business-persistence negative gates.<!-- verified: 2026-07-12 evidence="codegen idempotent; OpenAPI/fixtures/diff clean; isolated migrate up-down-up v17 clean; full Go/make test; P0.098 real DB+browser; contexts/docs/index/diff clean" -->

## 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-12 | 2.5 | 补齐 assistant history 不得成为候选人事实的 RED/GREEN prompt 与负向 eval gate。 |
| 2026-07-12 | 2.4 | 补齐 TargetJob 绑定 resume/provenance/type/int32 目录约束，并增加 system policy 与 JSON 不可信上下文分层 gate。 |
| 2026-07-12 | 2.3 | 原地重开 Phase 7，按方案 A 持久化规范化轮次身份并由完成台账校验当前/复练/下一轮。 |
| 2026-07-12 | 2.2 | 原地重开 Phase 6，修复 start 只读空 structured_profile 导致真实简历丢失和无证据提问。 |
| 2026-07-12 | 2.1 | 经用户批准，将依赖真实 backend handler 的 P0.022 从 contract rebase 后移到 Phase 2，禁止以 fixture-only 证据替代。 |
