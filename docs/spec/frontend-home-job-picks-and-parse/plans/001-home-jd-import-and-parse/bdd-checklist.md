# 001 BDD Checklist

> **版本**: 2.24
> **状态**: active
> **更新日期**: 2026-07-14

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.014 Home 默认渲染与最近模拟面试

- [x] 场景目录 `test/scenarios/e2e/p0-014-home-default-render/` 保留 README、seed、expected outcome 与 `scripts/{setup,trigger,verify,cleanup}.sh`。
- [x] Historical trigger 覆盖 generated-client routing、Home shell/control、resume select、submit row、英文 i18n、ready filter、empty/one/twelve-plus fixtures、sort/3-card cap、More、card detail 与 quick-start handoff。
- [x] Verify 要求 real-mode generated-client 配置 marker、目标测试文件 marker、Vitest pass marker 和 out-of-scope source/log negatives。
- [x] 场景资产合同拒绝 TopBar/theme/mobile/build/Playwright/live-backend 等 runner 未执行的覆盖声明。
  <!-- verified: 2026-07-10 method=scenario-asset-contract evidence="HomeScreen asset gate passes inside the P0.014 trigger; generated-client 1 test plus Home 34 tests and all four scenario scripts pass." -->
- [x] Revision 2026-07-13 trigger proves Home intake only renders textarea、ready Resume select 和 CTA，并拒绝旧 source controls/modal/trigger/copy；Resume upload surface remains available under its own owner.
- [x] Revision 2026-07-13 captures 1440×900 / 390×844 Home screenshots and verifies DOM/computed-style/bbox/viewport parity against the updated prototype.
  <!-- verified: 2026-07-14 method=artifact-reconcile evidence="P0.014 result.json is PASS; desktop/mobile screenshots are 1440x900 and 1170x2532 for CSS viewports 1440x900 and 390x844, with changedRatio 0.000323 and 0.000481." -->

## E2E.P0.015 Home import 到 Parse command progress

- [x] 场景目录 `test/scenarios/e2e/p0-015-jd-import-and-parse/` 保留 README、seed、expected outcome 与 `scripts/{setup,trigger,verify,cleanup}.sh`。
- [x] Historical trigger covered explicit ready resume selection、import、4xx inline error、failed parse state、loading browser gate 和 preview mapping；Phase 18 supersedes the current intake assertions.
- [x] Historical verify required real-mode generated-client、idempotency、真实 `resumeId` parse route、loading screenshot 与 privacy markers；Phase 18 replaces its current request-shape evidence.
- [x] Revision 2026-07-13 updates prototype/formal loading assertions and active negatives so internal model/provider、rubric/prompt/version/hash、provenance and typical-latency metadata cannot render.
- [x] Revision 2026-07-13 captures clean 1440/390 Parse loading screenshots and verifies DOM/computed-style/bbox/viewport parity before preview.
- [x] Historical 2026-07-13 trigger covered successful `targetJobId + resumeId` route handoff；Phase 20 retains exact `{ rawText, targetLanguage, resumeId }` POST/idempotency/error coverage but supersedes navigation to targetJobId-only.
- [x] Revision 2026-07-13 auth trigger proves `pendingAction` stores only `opaquePendingImportId`；raw JD / language / resume / original idempotency key live only in a process-memory one-shot vault. Normal login atomically consumes and dispatches once；refresh/lost vault、expired and duplicate consume dispatch zero imports, clear the invalid action and return Home with localized re-entry guidance；route/browser storage/log/telemetry remain raw-JD-free.
- [x] Revision 2026-07-13 verify requires paste-only real-mode marker、privacy marker、old intake negative marker and clean 1440×900 / 390×844 Home/Parse screenshot evidence.
  <!-- verified: 2026-07-14 method=artifact+focused-test-reconcile evidence="P0.015 result.json is PASS; focused Home/auth/Parse tests prove the exact request, one-shot opaque vault recovery and clean loading state. Home and Parse desktop/mobile screenshots carry exact viewport dimensions; Parse changedRatio is 0.000556 desktop and 0.001499 mobile." -->

## E2E.P0.016 面试规划详情只读收据与 Start handoff

- [x] 场景目录 `test/scenarios/e2e/p0-016-parse-confirm-to-workspace/` 保留 README、seed、expected outcome 与 `scripts/{setup,trigger,verify,cleanup}.sh`。
- [x] Historical trigger covered route-only `resumeId` ignored for binding；Phase 20 removes that param and keeps `updateTargetJob` absence, direct Workspace-detail Start, auth continuation and failed/empty guard.
- [x] Current verify requires real-mode generated-client marker, body schema marker, direct practice route marker and privacy marker.
- [x] Revision 2026-07-09 trigger covers user-facing copy `面试规划详情 / 面试上下文确认`, shared detail DOM, and absence of out-of-scope "JD 解析结果" page naming in positive UI.
- [x] Historical revision 2026-07-09 verify confirmed Save/Start used generated `updateTargetJob` and handed off to workspace auto-start without raw JD / source leakage before the readonly simplification.
- [x] Revision 2026-07-09 readonly trigger covers inherited bound resume display, disabled Start when saved plan lacks bound resume, absence of editable inputs / requirement toggles / hidden remove / resume picker / create fallback / success Re-parse / Save plan / Cancel, and direct Start click.
- [x] Revision 2026-07-09 readonly verify established no `updateTargetJob` and direct practice handoff；Phase 20 applies that behavior only to Workspace success detail.
- [x] Revision 2026-07-09 round-data trigger now covers saved `TargetJob.summary.interviewRounds[]` rendering inside `home-recent-mock-rail-*`, `parse-round-*` cards and shared navigation context, including variable round count, round type/name, duration and focus.
- [x] Revision 2026-07-09 round-data verify confirms related surfaces and shared navigation context do not use static `parse.round*Focus` / local default round strings when structured backend rounds exist.
- [x] Revision 2026-07-09 structured-rounds trigger covers saved 2~5 item `TargetJob.summary.interviewRounds[]` rendering inside `home-recent-mock-rail-*`, `parse-round-*` cards and shared navigation context, including variable round count, round type/name, duration and focus.
- [x] Revision 2026-07-09 structured-rounds verify rejects fixed 4-round template strings and fixed duration defaults when structured backend rounds exist.
- [x] Revision 2026-07-09 screenshot acceptance verifies readonly detail through Playwright screenshot attachment plus `E2E.P0.016 ... screenshotBytes=` marker.
- [x] Revision 2026-07-14 setup prepares one trusted ready TargetJob and hostile route/query inputs；不再组合报告 overview 或 Parse 内报告状态 fixture。
- [x] Revision 2026-07-14 trigger proves Workspace detail 内容区标题行右上角“面试报告”精确进入 `{ name: "reports", params: { targetJobId } }`，且全局 TopBar / Parse progress 无报告入口。
- [x] Revision 2026-07-14 trigger proves Parse command states 与 Workspace ready/target-switch 均不渲染嵌入式报告列表、不调用 `listTargetJobReports`；只读详情与 Start 保持可用。
- [x] Revision 2026-07-14 verify proves `section=reports`、report/status/round hostile query 被丢弃，旧锚点/滚动/聚焦/兼容逻辑为零，列表 consumer 只属于独立 report owner。
- [x] Historical revision 2026-07-14 captured the shared-detail entry source；Phase 20 requires its current formal/prototype Workspace route at 1440×900 / 390×844 and no Parse ready entry/report/model/rubric/provenance leakage.
  <!-- verified: 2026-07-14 scenario=E2E.P0.016 evidence="All setup/trigger/verify assertions and six Playwright cases pass; acceptance manifest records exact 1440x900/390x844 entry/list images, TopBar count=3, current-target isolation and no model/rubric leakage." -->

## E2E.P0.018 Workspace 列表回访统一详情

- [x] 场景目录 `test/scenarios/e2e/p0-018-workspace-default-render/` 保留 README、seed、expected outcome 与 `scripts/{setup,trigger,verify,cleanup}.sh`。
- [x] Historical trigger covered card selection into a Parse readonly route；Phase 20 supersedes it with targetJobId-only Workspace detail, shared DOM and no route-side auto-start.
- [x] Verify 要求 workspace list re-entry marker、unified plan-detail marker、out-of-scope independent workspace detail negative marker、resume binding marker 与 privacy marker。

## Phase 20 command-only Parse and request-count revision

- [x] `E2E.P0.014` under StrictMode proves bottom-transport `listTargetJobs` count=1 and `listResumes` count=1 for same-key initial load; ready card body directly enters `/workspace?targetJobId` and never Parse.
- [x] `E2E.P0.015` proves POST import navigates only `/parse?targetJobId`; queued/processing polling advances only on scheduler ticks, each tick count=1; ready uses replace and browser Back cannot replay Parse animation.
- [x] `E2E.P0.016` proves ready detail lives at `/workspace?targetJobId` while readonly/report-entry/Start/parity behavior remains；detail uses one same-key `getTargetJob`, zero `listResumes`, and no Parse animation.
- [x] `E2E.P0.018` supersedes the historical card-to-Parse assertion: query-free list card directly enters workspace detail; exactly one initial `getTargetJob`; zero import/poll/auto-start/Parse-animation calls.
- [x] Run and record fresh setup/trigger/verify/cleanup evidence for existing P0.014/P0.015/P0.016/P0.018; do not create a sibling scenario or disable StrictMode.
  <!-- verified: 2026-07-14 evidence="P0.014/P0.015/P0.016/P0.018 all completed fresh setup/trigger/verify/cleanup PASS; transport markers are exact and StrictMode remains enabled." -->

## Phase 21 Workspace detail round-state revision

- [x] `E2E.P0.016` fixture provides at least one completed, one exact current and one pending structured round from persisted `practiceProgress`.
- [x] Browser asserts visible “已进行 / 即将进行 / 未进行” (or English equivalents), exact `data-round-state=done/current/pending`, and three distinct computed background/border treatments.
- [x] Browser/focused coverage proves all-completed renders every card done and invalid/missing progress renders no false done/current; lifecycle status, URL and browser storage do not affect the state.
- [x] Screenshot/parity evidence covers 1440×900 and 390×844 Workspace detail and compares the same persisted state with Home/Workspace mini rail; no sibling scenario is created.
  <!-- verified: 2026-07-14 evidence="P0.016 and focused invalid/all-completed gates PASS; desktop/mobile parity plus acceptance evidence.json and screenshot 03 prove the same persisted rail/card states and three distinct treatments." -->

## Phase 22 JD text boundary

- [ ] P0.015 consumes RuntimeConfig/default 98,304-byte limit and covers ASCII/multibyte limit/+1.
- [ ] Limit sends one exact import; +1 sends zero import and creates no pending vault; localized recovery and privacy remain intact.
- [ ] P0.010 backend evidence proves the same limit is authoritative with zero business/provider side effects on +1.

## 整体收口

- [x] Revision 2026-07-13 reruns P0.014/P0.015 after paste-only intake and loading-metadata cleanup, then records current 1440×900 / 390×844 screenshot evidence before Phase 18 hands off to Phase 19.
  <!-- verified: 2026-07-14 decision="用户批准方案 A" evidence="Both persisted scenario results are PASS with source fingerprints and the required desktop/mobile screenshot metadata. The plan remains active and final completion is owned by Phase 19." -->
- [x] Historical P0.014、P0.015、P0.016、P0.018 baseline executed `setup -> trigger -> verify -> cleanup` or recorded focused equivalent + scenario wrapper evidence；Phase 18 reruns P0.014/P0.015.
- [x] `make validate-fixtures` 确认相关 fixture 仍在当前 37-operation 合同内。
- [x] Owner 文档、context、INDEX 与 product-scope / workspace 证据同步。
- [x] Round assumptions shared data-binding regression and focused equivalent evidence are linked back to owner checklist Phase 7.
- [x] Structured interview rounds contract and scenario evidence are linked back to owner checklist Phase 8.
- [x] Phase 18 physically deletes URL-only `E2E.P0.011` and its active INDEX row without reusing the ID.
  <!-- verified: 2026-07-14 method=filesystem+index-negative evidence="The retired scenario directory is absent and the active E2E INDEX has no P0.011 row." -->
- [x] Phase 18 active zero-reference scan rejects old JD source controls/modal/route source/public schema/generated/backend branch/fixture/scenario, excludes legal historical evidence, and explicitly allows Resume upload assets.
  <!-- verified: 2026-07-14 method=active-surface-negative evidence="Exact production scans are clean; owner-doc/scenario matches are negative assertions only, and Resume upload plus source_records remain separate allowed owners." -->
- [x] Phase 19 P0.016 entry/no-embedded/no-request/no-section/parity evidence and report-owner-only `listTargetJobReports` consumer negative pass before owner closeout.
  <!-- verified: 2026-07-14 method=P0.016+scoped-negative evidence="P0.016 passes and ReportsScreen is the only formal frontend listTargetJobReports consumer." -->

## Internal Cleanup Substitute Gate

- [x] Phase 12 source negative, focused Home auth and frontend typecheck gates pass without changing E2E.P0.014-P0.018 behavior.<!-- verified: 2026-07-10 method=substitute+scenario-gate evidence="Source and Home focused tests passed; P0.015 setup/trigger/verify/cleanup passed including desktop/mobile browser checks." -->
