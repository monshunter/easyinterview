# 001 — Grounded Conversation Report Generation Checklist

> **版本**: 2.32
> **状态**: active
> **更新日期**: 2026-07-18

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
- [x] E2E-HANDOFF: P0.099 是唯一 real report/generating frontend/backend/provider/API/DB + exact-six visual owner；current run `e2e-p0-099-20260715T021319Z-57232` 已通过 Chrome exact-six/no-OCR 与 live API/PostgreSQL binding。
- [x] Provider/eval output is not an E2E scenario and is not a P0.099 prerequisite.

## Phase 9: Persistence and privacy closeout

- [x] Ready/failure persistence keeps jobID + claimed attempts fencing and zero stale-worker domain side effects.
- [x] Report/job/outbox/audit/log/metric surfaces retain no raw prompt/output or complete candidate context.
- [x] Development focused tests and required race/PostgreSQL tests pass; phase completion uses root `make test` for whole backend/frontend unit regression.

## Phase 10: Canonical-round report overview

- [x] Minimal closed wire、canonical order、independent current/latest selection and fail-closed identity/context tests pass.
- [x] ReportsScreen is the only list consumer; Parse/Report/Generating have zero list calls and no global/history center is introduced.
- [x] Generated/fixture handoff、root `make test` and scoped stale pagination/pointer negative search close the phase.
  <!-- verified: 2026-07-15 method=root-test+contract-search evidence="make test PASS; 37-operation generated/fixture handoff and ReportsScreen-only current/latest projection remain closed" -->

## Phase 11: Injected report input guard

- [x] A4 owns one typed default/override/invalid/cross-field contract suite.
- [x] Review uses one small injected admitted/overflow provider call/no-call test；the historical 62,397-byte symptom is not reconstructed and no default-size material is created.
- [x] A3 loader/coverage gates require all six active profiles `max_tokens >= 16384` and keep report context at 1,000,000; no byte/token capacity formula is used.
- [x] BDD-N/A: configuration wiring does not create a user workflow; no scenario or real large material is used.

## Phase 12: Report-owned conversation read

- [x] 12.1 RED: store/handler tests require owned report lookup, existing unique session relation, strict `seq_no ASC`, four report statuses and closed message projection.
  <!-- verified: 2026-07-15 method=go-test red=missing-read-model-and-handler green=report-conversation-store-and-handler -->
- [x] 12.2 FAILURE/PRIVACY-GATE: owned report 的空 `messages` 数组返回 200；missing/cross-user hidden 404；report/session/user/target mismatch、empty identity、blank content、missing createdAt、duplicate/non-increasing sequence、unknown role/additional locator fail closed with no partial transcript or raw log/audit/metric body；成功、业务错误与 auth 拒绝均为 `private, no-store`。
  <!-- verified: 2026-07-15 method=go-test cases=empty-200-hidden-404-identity-malformed-reportId-no-read-blank-content-created-at-order-role-closed-projection,mux-auth-no-store bug=BUG-0173 -->
- [x] 12.3 GREEN: implement generated `getReportConversation` handler/store with zero AI/write/pagination/new table; do not call `getPracticeSession` or reorder corruption into apparent success; set no-store before session middleware.
  <!-- verified: 2026-07-15 method=go-test source-negative=side-effect-ai-session-fallback-pagination,route-pre-auth-no-store -->
- [x] 12.4 REMOVAL-GATE: current OpenAPI/generated/router/handler/fixture/mock/frontend positive surface has zero `listPracticeSessions`; accepted history/decision and exact negative declarations are classified, not blanket-excluded.
  <!-- verified: 2026-07-15 method=scoped-negative-search evidence="only explicit removal tests, accepted history/decision text, and baseline oracle logic remain" -->
- [x] 12.5 BDD-Gate: `BDD.REPORT.CONVERSATION.API.001` passes owner tests; E2E.P0.099 receives real API/DB binding handoff without changing exact-six screenshots.
  <!-- verified: 2026-07-15 method=domain-behavior bddChecklist=complete -->
- [x] 12.6 COMPLETION-GATE: focused Go tests, root `make test`, OpenAPI/fixture/codegen/mock, docs/context/index/diff and migration-zero-change audit pass.
  <!-- verified: 2026-07-15 method=full-code-gates evidence="focused Go PASS; make test PASS; 149 contract tests PASS; codegen-openapi second-run hashes unchanged; docs/context/index/diff PASS; no migration change" -->

## Phase 13: Atomic failed report regeneration

- [x] 13.1 RED: handler/service/store/idempotency tests fail until same-ID `202 + ReportWithJob`, replay and typed rejection contracts exist.<!-- verified: 2026-07-16 method=backend-regenerate-red evidence="focused compilable RED proves missing operation-specific idempotency behavior, handler/domain/store contracts, same-ID response, cancellation-safe completion, job-before-report reset, typed zero-write rejection and real-PostgreSQL concurrency gate; diff-check PASS" -->
- [x] 13.2 GREEN: transaction locks job then report, resets all ready-only fields, enqueues one fresh stable-dedupe job and writes bounded audit without migration/outbox.<!-- verified: 2026-07-16 method=backend-regenerate-green evidence="focused idempotency/review/store/api packages PASS; same-report advisory serialization, active-job-first lock, ready-only reset, fresh queued job and bounded user audit are covered by sqlmock atomicity/rollback tests" -->
- [x] 13.3 CONCURRENCY/PRIVACY: active old job, simultaneous different keys, cross-user, non-failed and oversize paths are race-safe, zero-write and raw-content-free.<!-- verified: 2026-07-16 method=real-postgresql-integration evidence="current local PostgreSQL ran both regeneration integration tests without skip: simultaneous different keys produced exactly one active job while preserving frozen input/transcript; active-old-job, non-failed, oversize and cross-user subtests all PASS with zero writes" -->
- [x] 13.4 BDD-Gate: `BDD.REPORT.REGENERATE.001`, focused owner tests and root `make test` pass；checklist evidence records exact commands.<!-- verified: 2026-07-16 method=focused+root-regression evidence="go test internal/review, internal/store/review and internal/api/reports PASS; real PostgreSQL regeneration integration PASS without skip; make test Python 584/4583 subtests, Go all packages, frontend 126/1026 PASS" -->

## Phase 14: Actionable semantic repair

- [x] 14.1 FORENSICS/RED: current raw pairs and DB metadata prove six provider/schema-success calls, candidate user seqNos `2/4/6`, terminal assistant seqNo `7`, and repeated `not_user_message`; focused serializer/runtime tests must fail before guidance exists.<!-- verified: 2026-07-16 method=raw-db-forensics evidence="provider stop/schema-ok on 6 calls; repair coordinates repeatedly identify assistant anchor; redeploy interruption is secondary" -->
- [x] 14.2 GREEN/PRIVACY: derive sorted candidate-user seqNo allowlist only from validated server/eval coordinates；map every reachable validation code into explicit structural/anchor/evidence/readiness/action/text intent, combine all present families, validate restricted path grammar + code/path compatibility + positive unique strictly ascending seqNos + non-empty anchor allowlist, fail closed on unsafe/unknown coordinates, escape untrusted marker characters, preserve the same JSON semantics and the same escaped untrusted user message across initial/repair attempts, and reject previous-output/raw/untrusted-role promotion.<!-- verified: 2026-07-16 method=backend-report-repair-green evidence="runtime and evalkit derive trusted sorted user seqNos; terminal unanswered assistant regression exposes only [2,4,6]; concrete family-rule, unsafe-coordinate and literal-marker tests pass in internal/review and cmd/evalkit" -->
- [x] 14.3 REGRESSION: focused backend review/evalkit tests, root `make test`, build and context/docs gates pass.<!-- verified: 2026-07-16 method=full-regression evidence="focused review/evalkit repair tests PASS; make test Python 584/4583 subtests, Go all packages, frontend 126/1026 PASS; make build, context validators, docs/index and git diff --check PASS" -->
- [x] 14.4 REAL: redeploy backend and regenerate report `019f6a70-0b24-7c7b-a1d3-1456503a2421`; same-ID report reaches ready and DEBUG evidence proves valid user anchors without exposing content.<!-- verified: 2026-07-16 method=real-provider-same-report evidence="same report reached ready; fresh job succeeded; three stop/schema-valid calls used only user anchors [4,6]; missing_evidence repair requests carried exact-dimension support/removal intent and the third response passed full validation" -->

## Phase 15: Terminal unanswered assistant assessment projection

- [x] 15.1 RED: `reportCompletePayload` focused test proves the current provider payload still contains a trailing unanswered assistant message; paired/non-terminal assistant turns, terminal user answers, canonical ordering and source-slice immutability are asserted.
  <!-- verified: 2026-07-18 method=tdd-red command="cd backend && go test ./internal/review -run TestReportCompletePayloadExcludesOnlyTrailingUnansweredAssistant -count=1" result="failed because terminal assistant seqNo 3 remained in both canonical provider payload cases; terminal-user and paired-turn boundary passed" -->
- [x] 15.2 GREEN: derive a fresh provider assessment slice that removes exactly one trailing assistant message and nothing else; repair user-seq allowlists use the same projection, while persistence and `getReportConversation` remain unchanged.
  <!-- verified: 2026-07-18 method=tdd-green commands="focused terminal-assistant test; full internal/review package" result="pass; paired turns, terminal user, canonical ordering and source immutability preserved" -->
- [x] 15.3 REGRESSION: focused review tests, full review package, root `make test`, context/docs/index and diff gates pass.
  <!-- verified: 2026-07-18 method=full-regression evidence="focused and internal/review PASS; make test Python 584/4583 subtests, Go all packages, frontend 126 files/1027 tests PASS; context/docs/index/diff PASS" -->
- [x] 15.4 BDD-Gate: `BDD.REPORT.GENERATE.001` proves a new real Chrome report contains no assessment/action derived from the unanswered terminal topic while report conversation retains that final assistant message; desktop + 390×844 screenshots and console evidence are recorded.
  <!-- verified: 2026-07-18 method=real-chrome-db evidence="report 019f74c2-096e-7407-bccd-11c01fb59c40 ready; DB transcript count=3 terminal role=assistant; report topic checks false; desktop/mobile report+conversation screenshots; 390x844 overflowX=false; Chrome console zero error/warn; .test-output/e2e/chrome-full-regression-postfix-20260718/consolidated-coverage.md" -->

## Closeout

- [x] Root `make test` is the independent complete frontend/backend unit regression gate; code tests are never wrapped as E2E.
- [x] P0.099、docs/index/context/diff and deleted-scenario negative checks are reported separately.
- [ ] BDD-Gate: 在当前真实环境显式运行 `E2E.P0.099` 并完成 exact-six no-OCR audit；本轮未执行。
