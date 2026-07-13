# 001 BDD Checklist

> **版本**: 2.19
> **状态**: active
> **更新日期**: 2026-07-13

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.014 Home 默认渲染与最近模拟面试

- [x] 场景目录 `test/scenarios/e2e/p0-014-home-default-render/` 保留 README、seed、expected outcome 与 `scripts/{setup,trigger,verify,cleanup}.sh`。
- [x] Historical trigger 覆盖 generated-client routing、Home shell/control、resume select、submit row、英文 i18n、ready filter、empty/one/twelve-plus fixtures、sort/3-card cap、More、card detail 与 quick-start handoff。
- [x] Verify 要求 real-mode generated-client 配置 marker、目标测试文件 marker、Vitest pass marker 和 out-of-scope source/log negatives。
- [x] 场景资产合同拒绝 TopBar/theme/mobile/build/Playwright/live-backend 等 runner 未执行的覆盖声明。
  <!-- verified: 2026-07-10 method=scenario-asset-contract evidence="HomeScreen asset gate passes inside the P0.014 trigger; generated-client 1 test plus Home 34 tests and all four scenario scripts pass." -->
- [ ] Revision 2026-07-13 trigger proves Home intake only renders textarea、ready Resume select 和 CTA，并拒绝旧 source controls/modal/trigger/copy；Resume upload surface remains available under its own owner.
- [ ] Revision 2026-07-13 captures 1440×900 / 390×844 Home screenshots and verifies DOM/computed-style/bbox/viewport parity against the updated prototype.

## E2E.P0.015 Home import 到 Parse preview

- [x] 场景目录 `test/scenarios/e2e/p0-015-jd-import-and-parse/` 保留 README、seed、expected outcome 与 `scripts/{setup,trigger,verify,cleanup}.sh`。
- [x] Historical trigger covered explicit ready resume selection、import、4xx inline error、failed parse state、loading browser gate 和 preview mapping；Phase 18 supersedes the current intake assertions.
- [x] Historical verify required real-mode generated-client、idempotency、真实 `resumeId` parse route、loading screenshot 与 privacy markers；Phase 18 replaces its current request-shape evidence.
- [ ] Revision 2026-07-13 updates prototype/formal loading assertions and active negatives so internal model/provider、rubric/prompt/version/hash、provenance and typical-latency metadata cannot render.
- [ ] Revision 2026-07-13 captures clean 1440/390 Parse loading screenshots and verifies DOM/computed-style/bbox/viewport parity before preview.
- [ ] Revision 2026-07-13 trigger covers ready Resume selection、paste submit、exact `{ rawText, targetLanguage, resumeId }` body、idempotency、empty/4xx/failed paths and successful `targetJobId + resumeId` handoff, with no source discriminator or source route param.
- [ ] Revision 2026-07-13 auth trigger proves `pendingAction` stores only `opaquePendingImportId`；raw JD / language / resume / original idempotency key live only in a process-memory one-shot vault. Normal login atomically consumes and dispatches once；refresh/lost vault、expired and duplicate consume dispatch zero imports, clear the invalid action and return Home with localized re-entry guidance；route/browser storage/log/telemetry remain raw-JD-free.
- [ ] Revision 2026-07-13 verify requires paste-only real-mode marker、privacy marker、old intake negative marker and clean 1440×900 / 390×844 Home/Parse screenshot evidence.

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

## E2E.P0.018 Workspace 列表回访统一详情

- [x] 场景目录 `test/scenarios/e2e/p0-018-workspace-default-render/` 保留 README、seed、expected outcome 与 `scripts/{setup,trigger,verify,cleanup}.sh`。
- [x] Trigger 覆盖无上下文 `WorkspacePlanList`、plan-card selection、parse readonly detail route with `targetJobId/resumeId`、shared detail DOM 和 no-workspace-auto-start boundary。
- [x] Verify 要求 workspace list re-entry marker、unified plan-detail marker、out-of-scope independent workspace detail negative marker、resume binding marker 与 privacy marker。

## 整体收口

- [ ] Revision 2026-07-13 reruns P0.014/P0.015 after paste-only intake and loading-metadata cleanup, then records current 1440×900 / 390×844 screenshot evidence before this plan returns to completed.
- [x] Historical P0.014、P0.015、P0.016、P0.018 baseline executed `setup -> trigger -> verify -> cleanup` or recorded focused equivalent + scenario wrapper evidence；Phase 18 reruns P0.014/P0.015.
- [x] `make validate-fixtures` 确认相关 fixture 仍在当前 37-operation 合同内。
- [x] Owner 文档、context、INDEX 与 product-scope / workspace 证据同步。
- [x] Round assumptions shared data-binding regression and focused equivalent evidence are linked back to owner checklist Phase 7.
- [x] Structured interview rounds contract and scenario evidence are linked back to owner checklist Phase 8.
- [ ] Phase 18 physically deletes URL-only `E2E.P0.011` and its active INDEX row without reusing the ID.
- [ ] Phase 18 active zero-reference scan rejects old JD source controls/modal/route source/public schema/generated/backend branch/fixture/scenario, excludes legal historical evidence, and explicitly allows Resume upload assets.

## Internal Cleanup Substitute Gate

- [x] Phase 12 source negative, focused Home auth and frontend typecheck gates pass without changing E2E.P0.014-P0.018 behavior.<!-- verified: 2026-07-10 method=substitute+scenario-gate evidence="Source and Home focused tests passed; P0.015 setup/trigger/verify/cleanup passed including desktop/mobile browser checks." -->
