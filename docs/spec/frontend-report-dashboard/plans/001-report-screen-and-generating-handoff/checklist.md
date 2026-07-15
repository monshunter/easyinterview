# 001 — Honest Grounded Report Screen Checklist

> **版本**: 3.8
> **状态**: completed
> **更新日期**: 2026-07-15

**关联计划**: [plan](./plan.md)

## Phase 1-5: Conversation-level baseline（历史已完成）

- [x] Conversation report、generating、replay、routing and parity baseline completed.

## Phase 6: Honest GeneratingScreen

- [x] Remove fake progress/observations/notify and render only backend queued/generating/ready/failed truth.
- [x] Polling preserves attempt/delay across hidden/blur, resumes at n+1, rejects duplicate concurrency and keeps one run `<=49` calls.
- [x] Typed timeout/network/context-too-large/failure actions expose no provider/async attempt details.

## Phase 7: Direct ReportDashboard contract

- [x] Consume generated direct-report shape and fail closed on unknown/malformed context/focus.
- [x] Replay/next requests send no client focus/settings and use server-owned projection.
- [x] English 24/25 and zh-CN 64/65 code tests, delimiter parity and no raw/truncation/rewrite pass.

## Phase 8: Visual and real-environment separation

- [x] Deterministic formal DOM/style/viewport component assertions run as a frontend code gate, not E2E.
- [x] BDD-Gate: `BDD.REPORT.UI.001` 由 [BDD checklist](./bdd-checklist.md) 关联 report/generating owner behavior tests。
- [x] E2E-HANDOFF: P0.099 是唯一 real report/generating owner，要求 exactly six `fullPage: true` images 绑定 current API/DB/report/session/context/screenshot digests；本轮未运行，状态仍为 `Ready`。
- [x] P0.099 contract 要求 real mobile ready images 完整显示 action region 且无 clipping/ellipsis/hiding/overflow；exact 24/64 保持 code test。

## Phase 9: Context Strip privacy

- [x] Target/round/resume stay visible while report/session UUIDs are absent from text、tooltip、ARIA and accessible names.
- [x] Formal real-backend acceptance screenshots and manifest use bounded redacted state/hash/viewport evidence only.

## Phase 10: ReportsScreen

- [x] Current target joins canonical round display and renders current/latest only; cross-target/stale/mismatch data fail closed.
- [x] ReportsScreen is the sole list consumer; Parse/Report/Generating have zero list calls and no global/history center exists.
- [x] Report/Generating Back uses trusted target or Workspace fallback while routes stay reportId-only.

## Phase 11: Command/read navigation

- [x] Reports Back reaches targetJobId-only Workspace detail directly with no Parse detour、animation、import or polling.
- [x] Focused component/route/source tests and deterministic parity pass.

## Phase 12: Bottom full-width interview summary hierarchy

- [x] 12.1 RED: `ConversationReport` tests prove the current top readiness+summary card violates the new contract; assert ready DOM order `3/2/2/2/1`、two top metrics、one bottom Overall Summary and exactly one rendered `summary`.
  <!-- verified: 2026-07-15 method=vitest-red expected-failure="summary metrics had 3 children" -->
- [x] 12.2 GREEN: move localized readiness and the existing server `summary` into a localized bottom “面试总评” card; keep dimension/evidence counts as the only Summary Metrics and make no API/backend/persistence/prompt change.
  <!-- verified: 2026-07-15 method=vitest evidence="report owners 10 files 100 tests PASS; focused layout 21 PASS" -->
- [x] 12.3 RESPONSIVE/A11Y: formal frontend 1440/390 gates prove desktop `3/2/2/2/1`、bottom full-width span、mobile same-order single column、complete wrapping、no horizontal overflow and accessible title/readiness/summary.（验证：Chrome desktop counts=3/2/2/2/1、Overall Summary `1 / -1`；390 三组单列、summary 在 actions 后、无横溢）
- [x] 12.4 BDD-Gate: `BDD.REPORT.UI.001` ready branch and [BDD checklist](./bdd-checklist.md) cover the revised hierarchy; historical Phase 1-11 PASS is not current Phase 12 evidence.
- [x] 12.5 RED: `ConversationReport` and responsive contract tests reject the detached conversation button, three-column Context Strip, missing canonical resume-copy URL and detail panel cards whose visible borders do not fill the same grid row.
  <!-- verified: 2026-07-15 method=vitest-red expected-failures="context children 3 not 4; desktop repeat(3) not repeat(4); equal-height panel classes absent" -->
- [x] 12.6 GREEN: Context Strip renders target/round/resume/interview-record as four peer children；resume exposes canonical `/resume-versions?resumeId=<frozen id>` with SPA plus copy/new-tab semantics；interview record uses an in-strip SPA action without exposing reportId in DOM attributes.
  <!-- verified: 2026-07-15 method=vitest evidence="126 files / 1003 tests PASS; report/session sentinel privacy preserved" -->
- [x] 12.7 RESPONSIVE/A11Y: desktop `4/2/2/2/1` geometry proves both detail pairs have equal top/bottom bounds with internal whitespace on the shorter side；390 keeps the same order as a single column, links remain keyboard accessible and no horizontal overflow appears.
  <!-- verified: 2026-07-15 method=chrome desktop="context=4; columns=4; dimensions/highlights top=488 bottom=856; issues/actions top=874 bottom=998; scrollWidth=viewportWidth; reportIdInDom=false" mobile="390x844; columns=1; scrollWidth=390; natural panel heights" navigation="resume canonical href PASS; report conversation action/back PASS" screenshots=".test-output/acceptance/report-context-grid/report-{desktop-1440x1200,mobile-390x844}-full.png" -->
- [x] 12.8 BDD-Gate: `BDD.REPORT.UI.001` ready branch and [BDD checklist](./bdd-checklist.md) cover the four-item Context Strip、resume-copy navigation、conversation navigation and equal-height desktop pairs.
  <!-- Behavior-Verify: BDD.REPORT.UI.001 owner assertions and current Chrome observations cover the four peers, canonical frozen-resume navigation, privacy-preserving conversation navigation, desktop equal-height pairs, and mobile natural-height single-column flow. -->
- [x] 12.9 E2E-HANDOFF: align the existing P0.099 README/manual-audit/capture-verification contract with the four-item context、responsive detail-pair alignment and bottom interview summary；use the current real-backend ready report for focused Chrome desktop/mobile screenshot and navigation acceptance, without duplicating unchanged generating/language provider resources.
  <!-- verified: 2026-07-15 method="scenario-contract-red-green+real-chrome-focused-acceptance" evidence="P0.099 evidence tests 8 PASS; exact current desktop/mobile PNGs saved; resume-copy and conversation/back routes PASS; historical exact-six images not reused" -->
- [x] 12.10 CLOSEOUT: root `make test`、typecheck/build/lint/docs/index/diff gates pass independently, current Chrome screenshot evidence is recorded, and plan/spec/test/BDD lifecycle returns to `completed` only after all required Phase 12 checks finish.
  <!-- verified: 2026-07-15 evidence="make test: 559 Python / 4481 subtests, Go all packages, frontend 126/1003 PASS; make lint PASS; build PASS; typecheck PASS; docs/index PASS; diff-check PASS; Chrome acceptance manifest hash PASS" -->

## Phase 13: Report-owned readonly conversation integration

- [x] 13.1 RED/GREEN: current main has no `report_conversation` route or entry; merge generated `getReportConversation`, backend read projection, formal route/screen and Report + ReportsScreen entries while preserving Phase 12 layout ownership.
  <!-- verified: 2026-07-15 method=merge-conflict-reconcile+frontend-backend-focused-tests -->
- [x] 13.2 CONTRACT/FAILURE: focused frontend/backend tests prove reportId-only, strict ordered user/assistant Markdown, correct report/generating Back, hidden-404/stale/invalid projection fail-closed and no internal IDs/live controls/session list.
  <!-- verified: 2026-07-15 method=vitest+go-test+fixture-validator evidence="frontend full suite 989 PASS; report backend owner packages PASS; 37/37 fixtures validate" -->
- [x] 13.3 REGRESSION: OpenAPI/fixture/codegen/negative gates and root `make test` pass; deleted `ui-design/` Demo/prototype sync assets remain absent.
  <!-- verified: 2026-07-15 method=make-test+codegen-check+ui-demo-pruning evidence="551 Python tests and 4493 subtests PASS; Go all packages PASS; frontend 989 PASS; active residuals=0" -->
- [x] 13.4 BDD-Gate: `BDD.REPORT.CONVERSATION.001` 由 [BDD checklist](./bdd-checklist.md) 的 code-owner behavior tests 验证。
  <!-- verified: 2026-07-15 method=domain-behavior bddChecklist=complete -->
- [x] 13.5 E2E-HANDOFF: P0.099 remains the real API/UI owner; this merge does not claim a new scenario PASS unless explicitly run.
  <!-- verified: 2026-07-15 method=static-handoff-only evidence="scenario evidence unit tests 7 PASS; no E2E run claimed" -->

## Historical Closeout through Phase 11

- [x] Root `make test` is the independent complete backend/frontend unit regression gate; focused test PASS is development feedback, not full regression.
- [x] P0.099、typecheck/build/lint/docs/index/diff are reported as separate gates; code gates are never wrapped as E2E.
