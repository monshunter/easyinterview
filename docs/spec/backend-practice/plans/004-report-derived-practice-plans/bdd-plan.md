# Report-derived Practice Plans BDD Plan

> **版本**: 1.10
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Plan**: [plan](./plan.md)

## 行为合同

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.PRACTICE.DERIVED.001` | 用户已有已完成报告；报告、轮次或绑定也可能无效 | 请求复练当前轮、进入下一轮或重试派生 plan | 从持久化报告与 canonical round identity 幂等派生并保持 resume/TargetJob 绑定；非法关系 fail closed 且不写脏数据 | `backend/internal/practice/conversation_service_test.go` + `backend/internal/store/practice/derived_plan_test.go`，由根 `make test` 承接 |

当前没有覆盖 report-derived plan 创建或 session start 的真实 API/UI E2E owner。代码层回归统一由根 `make test` 承接，不能作为 E2E 证据。
