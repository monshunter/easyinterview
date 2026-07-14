# Practice Text Event Loop BDD Plan

> **版本**: 2.10
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Plan**: [plan](./plan.md)

## 行为合同

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.PRACTICE.TEXT.001` | 已存在可训练或已完成的 plan/session；输入也可能不可接受 | 用户发送、重试、完成或在终态继续发送文本 | UI 展示持久化 user/assistant turns 并一致恢复；终态/非法输入 fail closed，不伪造消息或完成状态 | `frontend/src/app/screens/practice/PracticeScreen.test.tsx` + `hooks/usePracticeMessages.test.tsx`，由根 `make test` 承接 |

当前没有覆盖 chat、session start 或文本 event loop 的真实 API/UI E2E owner。代码层回归统一由根 `make test` 承接，不能作为 E2E 证据。
