# 001 — Plan and Session Orchestration Checklist

> **版本**: 2.10
> **状态**: active
> **更新日期**: 2026-07-21

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

- [x] 8.1 RED: OpenAPI inventory/generated/router/handler/fixture/mock/source tests fail while `GET /practice/sessions` / `listPracticeSessions` remains a current positive surface.
  <!-- verified: 2026-07-15 method=preexisting-red-mutation+focused-green evidence="OpenAPI mutation rejects restored GET/list operation; Python registry 4 tests and Go mockruntime removed-list test PASS" -->
- [x] 8.2 GREEN: remove list operation, generated method/server interface, mux route, handler/service/store path, fixture and mock registry entry without redirect/deprecated/empty compatibility.
  <!-- verified: 2026-07-15 method=focused-source+contract evidence="production OpenAPI/generated/backend/frontend list surface zero; legacy fixture absent; registry and mockruntime PASS" -->
- [x] 8.3 PRESERVATION-GATE: `POST /practice/sessions` start and `GET /practice/sessions/{sessionId}` live recovery remain generated, routed, fixture-backed and behavior-tested.
  <!-- verified: 2026-07-15 method=fixture+domain+generated evidence="practice fixture scenarios 2 tests PASS; backend start/get focused tests PASS; exact operations and generated methods preserved" -->
- [x] 8.4 HANDOFF-GATE: completed transcript is reachable only through backend-review `getReportConversation(reportId)` over the existing relation; Workspace/Reports/Practice have no session-list consumer and no migration/table is introduced.
  <!-- verified: 2026-07-15 method=backend+frontend+source+diff evidence="review/store/api focused tests PASS; ReportConversationScreen 17 tests PASS; production consumer and SQL relation/order audit PASS; migrations unchanged from main" -->
- [x] 8.5 BDD-N/A: no current user-visible session-list flow exists; substitute gates are OpenAPI exact inventory/diff, fixture/codegen/mock parity, focused handler/source negatives and root `make test`.
  <!-- verified: 2026-07-15 method=behavior-id+scenario-negative+owner-tests evidence="no session-list BDD/E2E asset; retained BDD.PRACTICE.PLAN.001 start/get tests PASS" -->
- [x] 8.6 COMPLETION-GATE: root `make test`, build, OpenAPI/fixture/codegen/mock, docs/context/index/diff and scoped zero-reference gates pass before restoring completed status.
  <!-- verified: 2026-07-15 method=full-post-refreeze-regression evidence="root Python 551/4493, Go all, frontend 125/993; build PASS; OpenAPI diff/lint/37 fixtures/codegen/mock/consumers/docs/context/diff/zero-reference PASS" -->

## Phase 9: Recover an existing active session on repeated start

- [x] 9.1 RED: service/store tests reproduce same-user/plan active-session conflict and assert zero AI/new session/opening/lifecycle/outbox/audit side effects.
  <!-- verified: 2026-07-18 method=go-test-red evidence="service contract initially failed on missing recovery boundary; real PostgreSQL returned practice session conflict for an existing running session" -->
- [x] 9.2 GREEN: different start keys serialize by user/plan; new key binds an admitted queued/running session, waits queued to reach running, and persists the exact recovered response.
  <!-- verified: 2026-07-18 method=go-test+postgres-integration tests="TestStartPracticeSessionRecoversRunningSessionWithoutOpeningSideEffects,TestStartPracticeSessionWaitsForQueuedRecoveryBeforeFinalizing,TestSQLRepositoryIntegration_StartRecoversSamePlanActiveSession" marker="active-session-start-recovery=PASS" -->
- [x] 9.3 PRESERVATION-GATE: same-key replay/mismatch/pending semantics, fresh start opening generation, cross-user/plan isolation and active-session unique index all pass.
  <!-- verified: 2026-07-18 method=focused-unit+all-practice-integration evidence="domain/store/api packages PASS; all store integration PASS; active unique index unchanged; OpenAPI zero breaking/additive findings" -->
- [x] 9.4 BDD-Gate: `BDD.PRACTICE.PLAN.001` owner behavior/integration tests prove repeated start recovers the same activity without duplicate facts.
  <!-- verified: 2026-07-18 method=owner-behavior+postgres-integration marker="active-session-start-recovery=PASS" -->
- [x] 9.5 RUNTIME-GATE: redeploy backend; Chrome skill starts an existing affected plan from a formal UI entry and reaches the original session while PostgreSQL before/after counts remain unchanged.
  <!-- verified: 2026-07-18 method=chrome+postgres existingSession=019f751a-b64b-7e01-b607-3c99372beff7 evidence="formal workspace start reached the same running session; sessions/messages/events/outbox/audit/aiTasks remained 1/1/1/1/1/0; succeeded idempotency records increased from 1 to 2" -->
- [x] 9.6 COMPLETION-GATE: focused tests, integration tests, root `make test`, build/OpenAPI/codegen/fixture/docs/context/index/diff gates pass before closeout.
  <!-- verified: 2026-07-18 method=root-gates evidence="focused Go + PostgreSQL integration PASS; make test PASS (Python 584/4583, Go all packages, frontend 127 files/1035 tests); make build/lint/codegen-check/docs-check/openapi-diff/validate-fixtures/git-diff-check PASS" -->
- [x] 9.7 RED: prove recovery finalization must lock the session row before reading/finalizing, and a queued reservation whose original starter never advances cannot poll indefinitely.
  <!-- verified: 2026-07-19 method=go-test-red evidence="ServiceOptions lacked a recovery deadline/waiter; SQL mocks rejected missing FOR UPDATE and missing queued/user fences in CommitSessionStart/FailSessionStart" -->
- [x] 9.8 GREEN: add session-row locking, bounded 100ms-to-1s polling with a 35-second queued recovery boundary, retryable orphan failure, and `status='queued'` fences on original commit/failure so late workers roll back all opening facts.
  <!-- verified: 2026-07-19 method=focused-go+postgres-integration tests="TestStartPracticeSessionExpiresQueuedRecoveryAfterBoundedWait,TestSQLRepositoryCommitSessionStartRecoveryLocksSessionBeforeFinalizing,TestSQLRepositoryIntegration_StartRecoversSamePlanActiveSession" evidence="bounded backoff and retryable timeout PASS; completion ordered first returns conflict; late original commit rolls back all opening facts" -->
- [x] 9.9 PRESERVATION/BDD-GATE: running/queued recovery, same-key replay/mismatch, caller cancellation, completion ordering, fresh start and zero duplicate opening/AI/lifecycle/outbox/audit behavior pass.
  <!-- verified: 2026-07-19 method=owner-unit+all-practice-postgres-integration evidence="api/domain/store packages PASS; all practice integration PASS; bounded orphan recovery, same-key, cancellation, completion order, fresh start and zero duplicate facts preserved" -->
- [x] 9.10 COMPLETION-GATE: focused unit/store tests, PostgreSQL integration, root `make test`, build/OpenAPI/codegen/fixture/docs/context/index/diff gates pass before remediation closeout.
  <!-- verified: 2026-07-19 method=full-closeout evidence="focused + all practice PostgreSQL integration PASS; make test PASS (Python 584/4583, Go all, frontend 127/1035); build/lint/codegen-check/openapi-diff/validate-fixtures/docs-check/context/index/git-diff-check PASS; OpenAPI diff 0/0/0" note="Makefile has no git-diff-check target, so git diff --check ran directly" -->

## Phase 10: Ground interviewer employer identity

- [x] 10.1 RED: prompt/contract tests reproduce TargetJob-versus-Resume company conflict and fail while the policy permits the interviewer to claim the Resume employer.
  <!-- verified: 2026-07-21 method=pytest-red test=test_practice_chat_v030_locks_interviewer_employer_identity_sources result="FileNotFoundError: v0.3.0.md absent" -->
- [x] 10.2 GREEN: `practice.session.chat` resolves the new immutable prompt/rubric pair; TargetJob/round is the only hiring-side identity source, Resume companies remain candidate-only, and anonymous targets produce no invented company name.
  <!-- verified: 2026-07-21 method=pytest+go-test+lint evidence="v0.3 exact prompt/rubric GREEN and active; role_identity weight=0.4; v0.2 remains exact-readable as the inactive rollback coordinate; prompt/rubric lints and registry package pass" -->
- [x] 10.3 EVAL-GATE: identity-specific strong/weak/assistant-history cases pass prompt/rubric lint and offline eval without weakening existing resume-grounding cases.
  <!-- verified: 2026-07-21 method=tdd-red-green evidence="RED: exact-32 and practice v0.3 identity-suite tests failed against the 28-case suite with seven unpinned cases, no role_identity scores and four missing identity cases. GREEN: prompt/rubric lint clean; evalkit/registry focused tests pass; make eval-offline reports drift-check 32 cases/9 prompts, offline grading 32 and Promptfoo 32 passed/0 failed/0 errors." -->
- [x] 10.4 BDD-Gate: `BDD.PRACTICE.PLAN.002` proves opening/reply identity separation through owner behavior tests and current real-provider acceptance when available; no mock-backed browser flow is claimed as E2E.
  <!-- verified: 2026-07-21 marker=PRACTICE_INTERVIEWER_IDENTITY_BEHAVIOR_PASS method="owner-contract+offline-eval+real-provider" evidence="exact v0.3 prompt/registry identity source tests pass; four identity eval classes pass; DeepSeek v4 flash produced 5/5 valid v0.3 completions across three anonymous-target runs, one named-target run and one assistant-history correction with zero Resume-employer identity claims; raw payloads deleted; live judge JSON parsing failed and is not claimed PASS; no browser/E2E claim" -->
- [x] 10.5 COMPLETION-GATE: exact registry coordinates, migration up/down/up, focused Practice/F3 gates, root `make test`, docs/context/index/diff and post-pass reconcile pass before closeout.
  <!-- verified: 2026-07-21 method=full-closeout evidence="exact v0.2/v0.3 registry and practice-only activation tests PASS; disposable PostgreSQL 22->23->22->23 PASS; make eval-offline 32/32; make test PASS Python 626/4628, Go all, frontend 137/1126; make build/lint/docs-check, three context validators, sync-doc-index and git diff --check PASS; BUG-0197 and delivery retrospective recorded" -->

## 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-21 | 2.10 | 原地追加 Phase 10：锁定 TargetJob 招聘方与 Resume 候选人雇主的身份边界，并增加版本化 prompt、评测与 BDD 验收。 |
| 2026-07-19 | 2.9 | 原地追加 Phase 9 P1 修复：恢复最终化行锁、queued 有界等待与 retryable orphan 收敛、原启动提交 fencing。 |
| 2026-07-18 | 2.8 | 新增 Phase 9：同 user/plan 活动会话恢复、plan-scoped 并发、精确新 key 最终化与 Chrome 真实验收。 |
| 2026-07-15 | 2.7 | 新增 Phase 8：删除 listPracticeSessions 全部正向 surface，保留 start/get live operations，并交接 report-owned conversation read。 |
| 2026-07-12 | 2.5 | 补齐 assistant history 不得成为候选人事实的 RED/GREEN prompt 与负向 eval gate。 |
| 2026-07-12 | 2.4 | 补齐 TargetJob 绑定 resume/provenance/type/int32 目录约束，并增加 system policy 与 JSON 不可信上下文分层 gate。 |
| 2026-07-12 | 2.3 | 原地重开 Phase 7，按方案 A 持久化规范化轮次身份并由完成台账校验当前/复练/下一轮。 |
| 2026-07-12 | 2.2 | 原地重开 Phase 6，修复 start 只读空 structured_profile 导致真实简历丢失和无证据提问。 |

## BDD Gate

- [x] BDD-Gate: `BDD.PRACTICE.PLAN.001` 由 [BDD checklist](./bdd-checklist.md) 关联现有 plan/session owner behavior tests；不创建或声明真实 E2E PASS。
- [x] BDD-Gate Phase 8: `listPracticeSessions` 无当前用户行为流，不新增 BDD/E2E；按 [BDD checklist](./bdd-checklist.md) 回归保留的 start/get 行为并执行代码层替代 gate。
  <!-- verified: 2026-07-15 method=owner-behavior+exact-audit+root-contract evidence="retained start/get behavior PASS; no new scenario; 15-finding old-baseline audit, mock parity and root tests PASS; no E2E claimed" -->
- [x] BDD-Gate Phase 10: `BDD.PRACTICE.PLAN.002` 按 [BDD checklist](./bdd-checklist.md) 验证面试官只代表 TargetJob 招聘方，匿名目标公司不猜名，Resume 公司不被代入面试官身份。
  <!-- verified: 2026-07-21 method=owner-behavior+offline-eval+real-provider evidence="v0.3 owner contracts and four identity case classes PASS; DeepSeek completion acceptance 5/5 with zero Resume-employer identity claims; root make test and independent F3/Practice gates PASS; live judge parse failure remains explicitly unavailable; no browser/E2E claim" -->
