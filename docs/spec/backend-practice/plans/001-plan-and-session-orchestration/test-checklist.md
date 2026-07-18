# 001 Plan and Session Orchestration Test Checklist

> **版本**: 2.6
> **状态**: active
> **更新日期**: 2026-07-18

**关联 Test Plan**: [test-plan](./test-plan.md)

## Phase 1: Contract tests
- [x] Phase 1 contract/migration/prompt negative and generation tests pass.

## Phase 2: Plan tests
- [x] Phase 2 plan validation/store/idempotency/isolation tests pass.

## Phase 3: Start tests
- [x] Phase 3 opening/failure/repair/replay/privacy tests pass.

## Phase 4: Read tests
- [x] Phase 4 ordered/empty/missing/cross-user/list tests pass.

## Phase 5: Gate set
- [x] Phase 5 仓库根 `make test` 完成前后端全量单测回归；codegen/migrate/docs/diff 作为独立 gates。

## Phase 6: Resume grounding
- [x] Store precedence and complete snapshot tail-marker tests pass.<!-- verified: 2026-07-12 method=go-test -->
- [x] Empty resume context returns typed validation and records zero AI/opening reply.<!-- verified: 2026-07-12 method=go-test -->
- [x] Persisted-resume/candidate-user evidence, clarification, assistant-history non-evidence lint and offline eval pass.<!-- verified: 2026-07-12 method=unittest+prompt-lint+eval-offline result=27/27-pass -->
- [x] System-role policy, JSON untrusted-context escaping and persona-style-only payload/lint/eval gates pass.<!-- verified: 2026-07-12 method=go+pytest+eval includes=TestBackendPracticeConversationPromptPreflight -->

## Phase 7: Round identity and plan selection

- [x] Contract/generated round fields and paired persistence tests pass.<!-- verified: 2026-07-12 method=unit+integration -->
- [x] Baseline/retry/next transactional selection and idempotency tests pass.<!-- verified: 2026-07-12 method=real-postgres-integration -->
- [x] TargetJob-bound resume/provenance plus same-duration/non-contiguous/int32/type/legacy/all-complete/invalid-source tests pass.<!-- verified: 2026-07-12 method=unit+real-postgres-integration markers="target-resume-binding-and-provenance,canonical-round-type-case-sensitive,non-contiguous-successor" -->

## Phase 8: Session-list removal

- [x] Retired list operation/route/handler/fixture/mock/generated/frontend positive-surface negatives pass.
  <!-- verified: 2026-07-15 method=focused-contract tests="openapi inventory + fixture registry + Go mockruntime" -->
- [x] Start-session and scoped get-session preservation tests pass.
  <!-- verified: 2026-07-15 method=fixture+go+source result=PASS -->
- [x] ReportConversation handoff and no-migration audit pass.
  <!-- verified: 2026-07-15 method=backend+frontend+source+diff result=PASS -->
- [x] Root `make test` plus OpenAPI/fixture/codegen/mock/docs/context gates pass.
  <!-- verified: 2026-07-15 method=post-refreeze-full-regression result=PASS -->

## Phase 9: Active-session start recovery

- [x] RED evidence captures the pre-fix same-plan conflict without weakening existing unique constraints.
  <!-- verified: 2026-07-18 method=go-test-red error="practice session conflict" constraint="idx_practice_sessions_one_active_per_plan retained" -->
- [x] Running and queued recovery return the original session, finalize a fresh key exactly, and call no opening AI/commit side effects.
  <!-- verified: 2026-07-18 method=focused-unit+real-postgres result=PASS -->
- [x] Concurrency, same-key mismatch/pending, fresh-start and different user/plan preservation tests pass.
  <!-- verified: 2026-07-18 method=focused-unit+all-practice-integration result=PASS -->
- [x] Existing affected session passes Chrome UI plus PostgreSQL zero-duplicate before/after verification.
  <!-- verified: 2026-07-18 method=chrome+postgres existingSession=019f751a-b64b-7e01-b607-3c99372beff7 counts="sessions=1 messages=1 events=1 outbox=1 audit=1 aiTasks=0" -->
- [x] Root regression and contract/docs/diff gates pass.
  <!-- verified: 2026-07-18 method=root-gates evidence="make test/build/lint/codegen-check/docs-check/openapi-diff/validate-fixtures and git diff --check PASS" -->
