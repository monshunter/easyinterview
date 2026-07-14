# TargetJob Import, Parse and Progress BDD Plan

> **版本**: 1.16
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Plan**: [plan](./plan.md)

## 当前真实 E2E owner

| 场景 | Given | When | Then | 真实 E2E |
|---|---|---|---|---|
| TargetJob 进度刷新 | real frontend/backend 已运行，用户已真实登录，TargetJob 已绑定轮次与 session | 调用真实 completion API 并刷新 Home、Workspace 与 TargetJob 详情 | 用户在三处都看到第一轮已完成、第二轮为当前、第三轮待进行，并可打开同一 TargetJob 详情 | `E2E.P0.098` |

`E2E.P0.098` 不承接 JD import/parse、chat、session start 或下一轮 plan 创建。

## Domain behavior

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.TARGETJOB.IMPORT.001` | 用户提交合法/非法 JD，解析也可能 retryable/non-retryable failure | import、parse、重试或读取 TargetJob | 合法输入幂等进入 ready；失败可恢复或终止且不回显原始材料；Get/List progress 投影 fail closed | `backend/internal/targetjob/workflow_contract_test.go` + `service_test.go`，由根 `make test` 承接 |

当前没有 JD import/parse 真实 E2E owner；`E2E.P0.098` 只作为 progress refresh 的独立 suite handoff。
