# Practice Event Loop and Completion BDD Plan

> **版本**: 2.12
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Plan**: [plan](./plan.md)

## 当前真实 E2E owner

| 场景 | Given | When | Then | 真实 E2E |
|---|---|---|---|---|
| completion 驱动进度刷新 | real backend/frontend 已运行，用户已真实登录且 session 可完成 | 调用真实 completion API，再刷新 Home、Workspace 与 TargetJob 详情 | 用户在三处都看到当前轮已完成、第二轮为当前、后续轮待进行的同一进度，并可打开同一 TargetJob 详情 | `E2E.P0.098` |

`E2E.P0.098` 不承接 chat、session start、opening message、quick-start 或下一轮 plan 创建。

## Domain behavior

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.PRACTICE.EVENT_LOOP.001` | session 处于 active、reply pending/retryable，或准备完成 | 用户发送、重试或完成会话 | user/assistant 顺序与 same-ID replay 一致；completion 原子持久化且失败不产生重复消息、报告 job 或进度事实 | `backend/internal/practice/conversation_service_test.go` + `complete_session_service_test.go`，由根 `make test` 承接 |

当前没有 chat/session-start 的真实 E2E owner；`E2E.P0.098` 只作为 completion/progress 的独立 suite handoff。
