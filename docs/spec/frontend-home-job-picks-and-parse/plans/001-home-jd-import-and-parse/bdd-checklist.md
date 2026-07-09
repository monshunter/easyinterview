# 001 BDD Checklist

> **版本**: 2.6
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

## E2E.P0.016 面试规划详情 Save/Start handoff

- [x] 场景目录 `test/scenarios/e2e/p0-016-parse-confirm-to-workspace/` 保留 README、seed、expected outcome 与 `scripts/{setup,trigger,verify,cleanup}.sh`。
- [x] Trigger 覆盖 route `resumeId` inheritance、explicit resume selection fallback、`updateTargetJob` supplied-fields body、Save plan workspace route、Start interview auto-start route、auth continuation 和 failed/empty guard。
- [x] Verify 要求 real-mode generated-client marker、body schema marker、workspace context marker、practice route marker 与 privacy marker。
- [x] Revision 2026-07-09 trigger covers user-facing copy `面试规划详情 / 面试上下文确认`, shared detail DOM, and absence of old "JD 解析结果" page naming in positive UI.
- [x] Revision 2026-07-09 verify confirms Save/Start still use generated `updateTargetJob` and hand off to workspace auto-start without raw JD / source leakage.

## E2E.P0.018 Workspace 列表回访统一详情

- [x] 场景目录 `test/scenarios/e2e/p0-018-workspace-default-render/` 保留 README、seed、expected outcome 与 `scripts/{setup,trigger,verify,cleanup}.sh`。
- [x] Trigger 覆盖无上下文 `WorkspacePlanList`、plan-card selection、workspace detail route with `targetJobId/resumeId`、shared detail DOM 和 `autoStartPractice=1` boundary。
- [x] Verify 要求 workspace list re-entry marker、unified plan-detail marker、retired independent workspace detail negative marker、resume binding marker 与 privacy marker。

## 整体收口

- [x] P0.014、P0.015、P0.016、P0.018 均执行 `setup -> trigger -> verify -> cleanup` 或记录 focused equivalent + scenario wrapper 证据。
- [x] `make validate-fixtures` 确认相关 fixture 仍在 35-operation 合同内。
- [x] Owner 文档、context、INDEX 与 product-scope / workspace 证据同步。
