# Grounded Conversation Report Test Checklist

> **版本**: 2.19
> **状态**: active
> **更新日期**: 2026-07-14

**关联 Test Plan**: [test-plan](./test-plan.md)

## Phase 1-5: Historical baseline

- [x] Historical conversation-level regression tests remain present.

## Phase 6: Frozen context and contract

- [x] Phase 6 ADR/merge-base OpenAPI and B4/F3/002 owner markers are consumed; review-only frozen-load/public-projection/trust-boundary/provenance tests pass without practice implementation ownership.
  <!-- verified: 2026-07-12 evidence="OpenAPI audit, storage v18, completion snapshot, prompt v0.2 activation and profile markers consumed; frozen-load/projection/trust-boundary/exact-provenance tests PASS" -->
- [x] Phase 6 deterministic boundary fixtures reconstruct to exact bytes/hashes and emit `REPORT_BOUNDARY_FIXTURES_READY`; after A3 marker, exact 48,000/+1 runtime tests pass and +1 makes zero provider/repair calls.
  <!-- verified: 2026-07-12 evidence="deterministic manifest/hash tests and exact 48000/48001 runtime boundary tests PASS; A3 real token smoke stop/usage evidence consumed" -->

## Phase 7: Direct semantics and reliability

- [x] Phase 7 HISTORICAL-SUPERSEDED (durable repair only) boundary/cross-field/generic-empty-focus/issue-backed-focus/anchor/pending-question/safety/persistence/legacy-identifier-negative tests passed; former durable-attempt evidence is not current.
  <!-- verified: 2026-07-12 evidence="review/store/API focused and full packages PASS; cross-layer executable legacy scan covers runtime/generated/OpenAPI/fixtures/scenarios" -->

## Phase 8: Replay, eval and UAT

- [x] Phase 8 consumes backend-practice/004 replay markers and F3/004 eval markers；fact→judgment→action product acceptance, privacy evidence and real-provider fixed-five sample pass；strict P0.100 result remains separately attributable.
  <!-- verified: 2026-07-13 evidence="P0.070/P0.072 server-owned focus/isolation PASS; F3 prompt/rubric and offline28/28 PASS; P0.100 run59381 mechanical9/9, semantic8/9, categories4/5, strict FAIL; P0.099 run12381 exact-six PASS" -->
  - [x] evalkit 同源 schema 基础 RED/GREEN：本历史证据只证明 schema validation、usage/latency 聚合与 judge 基础 wiring；被替换的 retry 限制不作为当前证据。
    <!-- verified: 2026-07-13 evidence="focused evalkit Go tests PASS; report prompt semantic lint PASS; P0.100 scenario contract 8/8 PASS" -->
  - [x] HISTORICAL-SUPERSEDED Product max4 generation evidence used durable pre-call reserve and runner requeue; retain only as former-contract audit evidence.
  - [x] HISTORICAL-SUPERSEDED Report job max_attempts4/reservation-failure evidence is not current product behavior.
  - [x] HISTORICAL-SUPERSEDED (reservation only) Lease takeover evidence coupled fencing to `llm_attempt_count`; current contract retains result/failure side-effect fencing only.
  - [x] Evalkit independent max4 generation/judge RED/GREEN：usage/latency aggregation + attempt/retry/reason/scope manifest；judge protocol invalid retries，valid negative typed rejection does not retry。
  - 旧 action-support 静态/离线 RED/GREEN 仅证明逐 code cited behavior、分号片段、umbrella/type-specific support结构；当前 en<=24 whitespace words、zh-CN<=64 Unicode code points边界已由 P0.099 与 UI parity 独立验证，且 backend/frontend 对 ECMAScript `/\s/u` 的 U+FEFF/U+0085 分隔语义有成对回归；18/52仍只作 targeted-repair margin。
  - [x] short-conservative generic-replay rubric conflict RED/GREEN：`e2e-p0-100-20260713T011140Z-36625` attempt1 的 `$.nextActions[0] invalid_partial` 经排除测试/环境/consumer drift后定位为 exact generic-replay 例外与 `report_action_quality` 冲突；judge/rubric、migration、active DB同步后同case复测weighted0.82/min0.70/零违规。完整矩阵在当时仍 pending，后由 run 35103 关闭。
  - [x] full-validator handoff：run80338 incompatible readiness/action escape remains historical；final run59381 proves all nine emitted outputs pass complete validation before judge。
  - [x] focused judge regressions：`e2e-p0-100-20260713T012359Z-59906` direct injection3次PASS；`e2e-p0-100-20260713T013642Z-75753` exact generic empty-focus same-digest+5次PASS；它们未单独替代完整矩阵，完整矩阵后由 run 35103 关闭。
  - [x] run25849 remains aborted/not-PASS；run35103 is historical strict evidence only and final-prompt run59381 is current evidence。
  - 已通过 focus 静态/离线合同：empty 只允许 single `answer_depth` brief 或 single `answer_relevance` control-only；其他 retry 使用完整升序唯一 same-code needs-work issue set，并拒绝 subset/superset 与 `I >= 2` empty；历史run35103与最终run59381的已生成输出均通过机械合同。
  - [x] UX audit：200只证明fuse；P0.099 desktop+390证明合法24/64完整换行，超限typed invalid/no raw；18/52不作UI边界。
    <!-- verified: 2026-07-13 run="e2e-p0-099-20260713T095144Z-12381" evidence="exact six full-page screenshots across two ready states plus generating; DB/API canonical digests and manual content audit bound; trigger+verify PASS" -->
- [x] HISTORICAL-SUPERSEDED P0.058 `report-backend-evidence.v2` satisfied the former durable/backoff contract; P0.056 v1 remains regression evidence.

## Phase 9: Action-local retry contract refresh

- [x] Invocation-local counter/waiter tests prove initial+3, exact10s/20s/40s, provider/protocol/invalid-output retries, dynamic scope/full validation and context cancellation.
  <!-- verified: 2026-07-13 evidence="backend full/race PASS; P0.058 v3 first action calls4 with waits10/20/40" -->
- [x] Two-invocation tests prove return-time retry-state destruction and second-action attempt1 reset after first-action exhaustion; no crash/replay global cap assertion remains.
  <!-- verified: 2026-07-13 evidence="P0.058 v3 records destroyed state and second invocation initial attempt1" -->
- [x] Migration/store/runtime negative tests prove no `llm_attempt_count`, pre-call product reservation, report explicit max_attempts4 product coupling or runner-owned report wait schedule.
  <!-- verified: 2026-07-13 evidence="migration lint/integration reject llm_attempt_count; dev PostgreSQL producer keeps generic max_attempts=5" -->
- [x] PostgreSQL lease takeover tests preserve stale result/failure zero report/outbox/audit/job side effects through jobID+claimed attempts without pre-call retry-state writes.
  <!-- verified: 2026-07-13 evidence="review/store/runner/practice race and PostgreSQL fencing regressions PASS" -->
- [x] P0.058 `report-backend-evidence.v3` exact schema, six markers, `database` fail-closed facts, `runtime` action/reset/separation facts and no-false-PASS checks pass.
  <!-- verified: 2026-07-13 evidence="four-stage P0.058 PASS; six markers; seven frontend files/51 tests PASS" -->
- [x] P0.100 current wording preserves evalkit generation/judge independent max4 evidence while removing report-lifetime durable/crash-cap/job-max4 claims.
  <!-- verified: 2026-07-13 evidence="final run59381 keeps evalkit state separate from P0.058 product state and terminally rejects a valid negative without retry" -->
- [x] Focused/full Go/race/PostgreSQL, contexts, docs/index and diff gates pass before completion.
  <!-- verified: 2026-07-13 evidence="focused/full Go, race, PostgreSQL and P0.058 PASS; contexts valid; docs/index zero drift; git diff --check clean" -->

## Phase 10: Canonical-round report overview

- [x] Closed-wire, canonical coverage/order, independent current/latest selection, tie-break, nullable/error-enum, hidden-404, invalid-context whole-response fail-closed and privacy tests pass.
  <!-- verified: 2026-07-14 evidence="Focused RED/GREEN plus full practice/review/store/reports package runs pass; handler error responses are asserted directly against listTargetJobReports fixtures including X-Request-ID/requestId/details." -->
- [ ] Generated/fixture owner handoff, P0.059 composition, ReportsScreen-only consumer negative, focused/full review/store/API and scoped stale-pointer/pagination searches pass.
