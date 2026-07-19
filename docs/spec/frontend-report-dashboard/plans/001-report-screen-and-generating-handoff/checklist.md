# 001 — Honest Grounded Report Screen Checklist

> **版本**: 4.3
> **状态**: completed
> **更新日期**: 2026-07-16

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

## Phase 14: Failed report recovery actions

- [x] 14.1 RED: failed-only、old-ready/new-failed、oversize、double-click、unknown outcome、stale target and malformed response tests fail before implementation.<!-- verified: 2026-07-16 method=frontend-regenerate-red evidence="generated-client 16/16 PASS; focused UI 19 expected RED across failed actions, same-ID envelope, single-flight, idempotency outcome, stale refetch, owner fence and trusted failed-conversation back route; typecheck/diff-check PASS" -->
- [x] 14.2 GREEN: generated client sends bodyless POST + IK, matching queued response navigates to same reportId Generating, and explicit terminal errors stay localized in-row.<!-- verified: 2026-07-16 method=frontend-regenerate-green evidence="same-ID five-field 202 validation, single-flight, outcome-aware IK reuse/rotation, typed conflict target+overview refetch and owner fence pass in 65 focused tests; devMock fixture parity raises focused total to 74/74; typecheck PASS" -->
- [x] 14.3 A11Y/PRIVACY: failed/current actions use distinct accessible names, conversation remains available during regenerate, and no raw error/provider/report ID enters visible/a11y text.<!-- verified: 2026-07-16 method=frontend-a11y-privacy-green evidence="old-ready/new-failed exposes four distinct actions, oversize keeps transcript only, localized row error hides raw/provider/id, and failed conversation hides Back while owner is resolving before routing trusted target owner to Reports or a resolved untrusted outcome to workspace; focused report-conversation 19/19 plus recovery suite and frontend diff-check PASS" -->
- [x] 14.4 BDD-Gate/REGRESSION: `BDD.REPORT.REGENERATE.UI.001`, focused frontend tests, typecheck/build and root `make test` pass.<!-- verified: 2026-07-16 method=focused+root-regression evidence="failed recovery/report conversation focused suites PASS including owner-resolving Back fence; typecheck/build PASS; root frontend 126 files/1026 tests with full Python/Go regression PASS" -->
- [x] 14.5 CHROME: current-run Chrome skill verifies the real local failed-report conversation + same-ID regenerate path and its transition through generating to ready；this scoped evidence does not mark P0.099 PASS.<!-- verified: 2026-07-16 method=chrome-real-local evidence="failed row exposed regenerate and latest-attempt conversation; transcript opened; same report entered generating and then ready against the real backend without interception" -->

## Phase 15: Completed-session conversation action is status-independent

- [x] 15.1 RED: queued/generating ReportsScreen tests fail until progress and the distinct latest-attempt conversation action coexist and navigate by the latest report locator.<!-- verified: 2026-07-16 method=frontend-conversation-availability-red evidence="ReportsScreen focused run produced exactly 2 expected missing-entry failures while the remaining 29 tests passed" -->
- [x] 15.2 GREEN/A11Y: every distinct latestAttempt, including queued/generating/latest-ready, exposes a localized conversation action；same-ID ready deduplicates and empty round exposes none.<!-- verified: 2026-07-16 method=frontend-conversation-availability-green evidence="status-neutral latest-attempt conversation action covers queued/generating/failed/latest-ready; same-ID ready deduplicates; ReportsScreen and language-switch focused tests 34/34 PASS; typecheck PASS" -->
- [x] 15.3 BDD/REGRESSION: focused ReportsScreen/i18n tests, typecheck/build and root `make test` pass.<!-- verified: 2026-07-16 method=focused+root-regression evidence="ReportsScreen/i18n status-independent actions PASS; typecheck/build PASS; make test Python 584/4583 subtests, Go all packages, frontend 126/1026 PASS" -->
- [x] 15.4 CHROME: real local queued/generating row shows “查看生成进度” plus “查看面试记录”, and the latter opens the report-owned transcript；do not claim P0.099 PASS.<!-- verified: 2026-07-16 method=chrome-real-local evidence="queued/generating row displayed both localized actions; conversation opened the same report-owned readonly transcript; the row later became ready" -->

## Historical Closeout through Phase 11

- [x] Root `make test` is the independent complete backend/frontend unit regression gate; focused test PASS is development feedback, not full regression.
- [x] P0.099、typecheck/build/lint/docs/index/diff are reported as separate gates; code gates are never wrapped as E2E.

## Phase 16: Reference-aligned report dashboard

- [x] 16.1 RED: report source/responsive tests 固化约 1336px desktop 内容面、共享 Header/Context/Detail/Overall 网格、CTA 主次、语义 icon、圆角卡片和 390px 单列。<!-- verified: 2026-07-19 method=vitest-red evidence="responsive contract failed on the old narrow report hierarchy before implementation" -->
- [x] 16.2 GREEN: 重构 ReportDashboard / ReportHeader / ReportContextStrip 与 owner CSS，清除 ready 主路径内联视觉拼装，保持 report truth、`4/2/2/2/1`、CTA 与状态语义。<!-- verified: 2026-07-19 method=focused-vitest evidence="report focused suite 30 tests PASS, including mobile header flex-basis regression" -->
- [x] 16.3 BDD-Gate: `BDD.REPORT.UI.001` 在新视觉层级下继续覆盖 ready/failure/route/CTA；正式 frontend 的 repository ready fixture 完成 Chrome 1916×821 / 390×844 验收，不声明真实 ready-report 或完整 `E2E.P0.099` PASS。<!-- verified: 2026-07-19 method=chrome-formal-frontend-fixture evidence="desktop x=290 width=1336 with 4/2/2/2/1 hierarchy; mobile header height=263.7 and context starts y=454.7 after flex-basis fix; documentOverflow=0" -->
- [x] 16.4 REGRESSION-GATE: focused frontend、根 `make test`、typecheck/build、context/docs/index/diff 通过后恢复 completed。<!-- verified: 2026-07-19 method=focused+root-regression evidence="report focused 30 PASS; root Python 615/4615 subtests, Go all packages, frontend 131 files/1054 tests PASS; typecheck/build PASS" -->
