# Practice Text Event Loop BDD Plan

> **版本**: 2.13
> **状态**: completed
> **更新日期**: 2026-07-20

**关联 Plan**: [plan](./plan.md)

## 行为合同

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.PRACTICE.TEXT.001` | 已存在可训练或已完成的 plan/session；输入也可能不可接受 | 用户在参考图层级的 Session Header、Transcript、Composer 中输入、发送、重试、完成或在终态继续发送文本 | desktop/mobile 布局无溢出；helper 固定属于 Composer 且不随 Transcript 移动；textarea/send 同属内层 input surface，按钮在非叠加底部 action area 右对齐且正文保持完整宽度；UI 展示持久化 user/assistant turns 并一致恢复；终态/非法输入 fail closed，不伪造消息或完成状态 | `frontend/src/app/screens/practice/PracticeScreen.test.tsx` + `Transcript.test.tsx` + `InputBar.test.tsx` + `PracticeVisual.test.ts` + `hooks/usePracticeMessages.test.tsx`，由根 `make test` 承接 |

当前没有覆盖 chat、session start 或文本 event loop 的真实 API/UI E2E owner。代码层回归统一由根 `make test` 承接，不能作为 E2E 证据。
