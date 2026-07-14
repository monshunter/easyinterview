# Workspace and Interview Context BDD Plan

> **版本**: 1.27
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Plan**: [plan](./plan.md)

## 当前真实 E2E owner

| 场景 | Given | When | Then | 真实 E2E |
|---|---|---|---|---|
| 完成后的进度刷新 | real frontend/backend 已运行，用户已真实登录，TargetJob 已有轮次与 session | 通过真实 completion API 完成第一轮，并刷新 Home、Workspace 与 TargetJob 详情 | 用户在三处都看到第一轮已完成、第二轮为当前、第三轮待进行，并可打开同一 TargetJob 详情 | `E2E.P0.098` |

`E2E.P0.098` 不承接 JD import/parse、chat、session start、quick-start 或下一轮 plan 创建。

## Domain behavior

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.WORKSPACE.CONTEXT.001` | 用户打开 Workspace list/detail，后端 progress/plan 可能完整、终态或无效 | 选择 TargetJob、查看轮次或发起训练 | 页面只消费后端投影并保持 route、隐私、exact-plan reuse 与 fail-closed 约束 | `frontend/src/app/screens/workspace/WorkspaceScreen.test.tsx` + `hooks/useWorkspaceTargetJobs.test.tsx`，由根 `make test` 承接 |

`E2E.P0.098` 是 completion/progress refresh 的独立 suite handoff；quick-start/session start/next-round 不归入该 E2E。
