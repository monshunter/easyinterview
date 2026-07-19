# Practice Text Event Loop BDD Checklist

> **版本**: 2.13
> **状态**: completed
> **更新日期**: 2026-07-20

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## `BDD.PRACTICE.TEXT.001` 持续文本对话

- [x] Owner behavior tests 覆盖发送、assistant turn、retry、completion 与终态 fail-closed。
- [x] 根 `make test` 已执行对应 Vitest；该结果是代码层行为证据，不是 E2E PASS。
- [x] 当前无 text-loop 真实 E2E owner；不创建 wrapper 场景。
- [x] Reference revision：owner tests 覆盖 Session Header、message surface、Composer、desktop/mobile containment，并保留发送、pending、retry、completion 与终态语义。
- [x] 根 `make test` 与正式 frontend repository fixture 的 Chrome 1916×821 / 390×844 视图验收完成；未运行的真实 active-session 业务流不声明 PASS。<!-- verified: 2026-07-19 evidence="root frontend 131 files/1054 tests PASS; desktop/mobile visual containment and refined role/prompt styles accepted" -->
- [x] Composer anchoring revision：owner tests 证明 Transcript 是唯一滚动区、Composer 固定在会话卡底部、helper 不在 Transcript 中；短/长聊天和 Transcript scroll 不改变输入框坐标或 helper/input gap。<!-- verified: 2026-07-19 evidence="Chrome scroll invariant input position and gap=8 at desktop/mobile" -->
- [x] 根 `make test` 与 Chrome desktop/mobile anchoring 验收完成；fixture 证据不声明真实 active-session E2E PASS。<!-- verified: 2026-07-19 evidence="root frontend 131 files/1054 tests PASS; desktop/mobile input position and helper gap remained invariant while transcript scrolled" -->
- [x] Composer input-surface revision：owner tests 与 Chrome desktop/mobile 证明 textarea/send 同属一个内层边界、send 位于非叠加底部 action area、窄屏文本保持完整宽度，且 Composer/helper 坐标和发送行为不回退。<!-- verified: 2026-07-20 evidence="focused 55 PASS; real-session Chrome desktop/mobile containment, full-width text, helper gap=8 and zero overflow PASS" -->
