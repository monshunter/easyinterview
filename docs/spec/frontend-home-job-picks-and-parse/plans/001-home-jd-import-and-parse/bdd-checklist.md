# Home JD Import and Parse BDD Checklist

> **版本**: 2.27
> **状态**: completed
> **更新日期**: 2026-07-15

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## `BDD.HOME.JD.001` JD 导入与解析交接

- [x] Owner behavior tests 覆盖 import、parse、失败恢复、确认 handoff 与隐私。
- [x] 根 `make test` 已执行对应 Vitest；该结果是代码层行为证据，不是 E2E PASS。
- [x] 当前无 JD import/parse 真实 E2E owner；P0.098 progress 场景不承接该行为。

## `BDD.HOME.JD.002` 共享 Workspace ready-detail 首行动作

- [x] 标题旁“绑定简历”只消费 saved `TargetJob.resumeId` 并打开对应 Resume 详情；缺失绑定不链接、不伪造、不提供 rebind。
- [x] “立即面试 + 面试报告”在标题下左对齐首行动作行按序呈现；desktop 同排、mobile 同序换行，Start/Report 事实与错误边界不变。
- [x] 独立 launch/binding block、标题右侧 Report、页尾 Start 的 DOM/source 负向 gate 为零；根 `make test` 与独立 responsive/a11y gate 通过，不声明真实 E2E PASS。
