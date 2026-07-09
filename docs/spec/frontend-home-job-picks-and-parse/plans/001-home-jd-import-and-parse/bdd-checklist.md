# 001 BDD Checklist

> **版本**: 2.10
> **状态**: completed
> **更新日期**: 2026-07-09

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.014 Home 默认渲染与最近模拟面试

- [x] 场景目录 `test/scenarios/e2e/p0-014-home-default-render/` 保留 README、seed、expected outcome 与 `scripts/{setup,trigger,verify,cleanup}.sh`。
- [x] Trigger 覆盖 Home source controls、resume select、submit row、empty/one/twelve-plus fixtures、3-card cap、More handoff、TopBar、i18n、theme 与 responsive layout。
- [x] Verify 要求 real-mode generated-client marker、Home focused tests marker、layout marker、3-card marker 与 privacy marker。

## E2E.P0.015 Home import 到 Parse preview

- [x] 场景目录 `test/scenarios/e2e/p0-015-jd-import-and-parse/` 保留 README、seed、expected outcome 与 `scripts/{setup,trigger,verify,cleanup}.sh`。
- [x] Trigger 覆盖 explicit ready resume selection、paste/manual_text import、upload/presign/file import、URL import、4xx inline error、failed parse state、loading browser gate 和 preview mapping。
- [x] Verify 要求 real-mode generated-client marker、`createUploadPresign` / `importTargetJob` idempotency marker、真实 `resumeId` parse route marker、loading screenshot marker 与 privacy marker。

## E2E.P0.016 面试规划详情只读收据与 Start handoff

- [x] 场景目录 `test/scenarios/e2e/p0-016-parse-confirm-to-workspace/` 保留 README、seed、expected outcome 与 `scripts/{setup,trigger,verify,cleanup}.sh`。
- [x] Historical trigger 覆盖 route `resumeId` inheritance、explicit resume selection fallback、`updateTargetJob` supplied-fields body、Save plan workspace route、Start interview auto-start route、auth continuation 和 failed/empty guard。
- [x] Historical verify 要求 real-mode generated-client marker、body schema marker、workspace context marker、practice route marker 与 privacy marker。
- [x] Revision 2026-07-09 trigger covers user-facing copy `面试规划详情 / 面试上下文确认`, shared detail DOM, and absence of old "JD 解析结果" page naming in positive UI.
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
- [x] Verify 要求 workspace list re-entry marker、unified plan-detail marker、retired independent workspace detail negative marker、resume binding marker 与 privacy marker。

## 整体收口

- [x] P0.014、P0.015、P0.016、P0.018 均执行 `setup -> trigger -> verify -> cleanup` 或记录 focused equivalent + scenario wrapper 证据。
- [x] `make validate-fixtures` 确认相关 fixture 仍在 35-operation 合同内。
- [x] Owner 文档、context、INDEX 与 product-scope / workspace 证据同步。
- [x] Round assumptions shared data-binding regression and focused equivalent evidence are linked back to owner checklist Phase 7.
- [x] Structured interview rounds contract and scenario evidence are linked back to owner checklist Phase 8.
