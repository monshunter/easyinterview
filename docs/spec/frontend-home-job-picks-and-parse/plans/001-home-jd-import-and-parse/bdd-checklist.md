# 001 BDD Checklist

> **版本**: 2.21
> **状态**: completed
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

## E2E.P0.015 Home import 到 Parse preview

- [x] 场景目录 `test/scenarios/e2e/p0-015-jd-import-and-parse/` 保留 README、seed、expected outcome 与 `scripts/{setup,trigger,verify,cleanup}.sh`。
- [x] Historical trigger covered explicit ready resume selection、import、4xx inline error、failed parse state、loading browser gate 和 preview mapping；Phase 18 supersedes the current intake assertions.
- [x] Historical verify required real-mode generated-client、idempotency、真实 `resumeId` parse route、loading screenshot 与 privacy markers；Phase 18 replaces its current request-shape evidence.
- [x] Revision 2026-07-13 updates prototype/formal loading assertions and active negatives so internal model/provider、rubric/prompt/version/hash、provenance and typical-latency metadata cannot render.
- [x] Revision 2026-07-13 captures clean 1440/390 Parse loading screenshots and verifies DOM/computed-style/bbox/viewport parity before preview.
- [x] Revision 2026-07-13 trigger covers ready Resume selection、paste submit、exact `{ rawText, targetLanguage, resumeId }` body、idempotency、empty/4xx/failed paths and successful `targetJobId + resumeId` handoff, with no source discriminator or source route param.
- [x] Revision 2026-07-13 auth trigger proves `pendingAction` stores only `opaquePendingImportId`；raw JD / language / resume / original idempotency key live only in a process-memory one-shot vault. Normal login atomically consumes and dispatches once；refresh/lost vault、expired and duplicate consume dispatch zero imports, clear the invalid action and return Home with localized re-entry guidance；route/browser storage/log/telemetry remain raw-JD-free.
- [x] Revision 2026-07-13 verify requires paste-only real-mode marker、privacy marker、old intake negative marker and clean 1440×900 / 390×844 Home/Parse screenshot evidence.
  <!-- verified: 2026-07-14 method=artifact+focused-test-reconcile evidence="P0.015 result.json is PASS; focused Home/auth/Parse tests prove the exact request, one-shot opaque vault recovery and clean loading state. Home and Parse desktop/mobile screenshots carry exact viewport dimensions; Parse changedRatio is 0.000556 desktop and 0.001499 mobile." -->

## E2E.P0.016 面试规划详情只读收据与 Start handoff

- [x] 场景目录 `test/scenarios/e2e/p0-016-parse-confirm-to-workspace/` 保留 README、seed、expected outcome 与 `scripts/{setup,trigger,verify,cleanup}.sh`。
- [x] Current trigger covers route-only `resumeId` ignored for binding, `updateTargetJob` absence, direct Start handoff, auth continuation and failed/empty guard.
- [x] Current verify requires real-mode generated-client marker, body schema marker, direct practice route marker and privacy marker.
- [x] Revision 2026-07-09 trigger covers user-facing copy `面试规划详情 / 面试上下文确认`, shared detail DOM, and absence of out-of-scope "JD 解析结果" page naming in positive UI.
- [x] Historical revision 2026-07-09 verify confirmed Save/Start used generated `updateTargetJob` and handed off to workspace auto-start without raw JD / source leakage before the readonly simplification.
- [x] Revision 2026-07-09 readonly trigger covers inherited bound resume display, disabled Start when saved plan lacks bound resume, absence of editable inputs / requirement toggles / hidden remove / resume picker / create fallback / success Re-parse / Save plan / Cancel, and direct Start click.
- [x] Revision 2026-07-09 readonly verify confirms Parse success detail does not call `updateTargetJob`, enters practice through practice handoff, and does not leak raw JD / source URL.
- [x] Revision 2026-07-09 round-data trigger now covers saved `TargetJob.summary.interviewRounds[]` rendering inside `home-recent-mock-rail-*`, `parse-round-*` cards and shared navigation context, including variable round count, round type/name, duration and focus.
- [x] Revision 2026-07-09 round-data verify confirms related surfaces and shared navigation context do not use static `parse.round*Focus` / local default round strings when structured backend rounds exist.
- [x] Revision 2026-07-09 structured-rounds trigger covers saved 2~5 item `TargetJob.summary.interviewRounds[]` rendering inside `home-recent-mock-rail-*`, `parse-round-*` cards and shared navigation context, including variable round count, round type/name, duration and focus.
- [x] Revision 2026-07-09 structured-rounds verify rejects fixed 4-round template strings and fixed duration defaults when structured backend rounds exist.
- [x] Revision 2026-07-09 screenshot acceptance verifies readonly detail through Playwright screenshot attachment plus `E2E.P0.016 ... screenshotBytes=` marker.
- [x] Revision 2026-07-14 setup prepares one trusted ready TargetJob and hostile route/query inputs；不再组合报告 overview 或 Parse 内报告状态 fixture。
- [x] Revision 2026-07-14 trigger proves内容区标题行右上角“面试报告”精确进入 `{ name: "reports", params: { targetJobId } }`，且全局 TopBar 无报告入口。
- [x] Revision 2026-07-14 trigger proves Parse ready/loading/failed/target-switch 均不渲染嵌入式报告列表、不调用 `listTargetJobReports`；原只读详情与 Start 保持可用。
- [x] Revision 2026-07-14 verify proves `section=reports`、report/status/round hostile query 被丢弃，旧锚点/滚动/聚焦/兼容逻辑为零，列表 consumer 只属于独立 report owner。
- [x] Revision 2026-07-14 captures formal/prototype Parse 入口 at 1440×900 / 390×844 and validates DOM/computed-style/bbox/viewport/pixel parity plus no report/model/rubric/provenance debug leakage.
  <!-- verified: 2026-07-14 scenario=E2E.P0.016 evidence="All setup/trigger/verify assertions and six Playwright cases pass; acceptance manifest records exact 1440x900/390x844 entry/list images, TopBar count=3, current-target isolation and no model/rubric leakage." -->

## E2E.P0.018 Workspace 列表回访统一详情

- [x] 场景目录 `test/scenarios/e2e/p0-018-workspace-default-render/` 保留 README、seed、expected outcome 与 `scripts/{setup,trigger,verify,cleanup}.sh`。
- [x] Trigger 覆盖无上下文 `WorkspacePlanList`、plan-card selection、parse readonly detail route with `targetJobId/resumeId`、shared detail DOM 和 no-workspace-auto-start boundary。
- [x] Verify 要求 workspace list re-entry marker、unified plan-detail marker、out-of-scope independent workspace detail negative marker、resume binding marker 与 privacy marker。

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
