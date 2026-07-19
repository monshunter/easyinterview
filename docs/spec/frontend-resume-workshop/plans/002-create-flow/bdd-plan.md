# Resume Workshop Create Flow BDD Plan

> **版本**: 1.15
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Plan**: [plan](./plan.md)

## 行为合同

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.RESUME.CREATE.001` | 用户在 desktop/mobile 选择 upload/paste；输入、类型、大小或请求也可能失败 | 创建、修正或重试 resume | 参考图层级的 Header/tab/dropzone/CTA 在两类 viewport 可达；UI 通过 API 进入 readonly detail/waiting；失败时保留可恢复输入，且不提前 presign/register 或泄露原文 | `frontend/src/app/screens/resume-workshop/create/CreateFlowIntegration.test.tsx` + `ResumeCreateFlow.test.tsx` + responsive source gate，由根 `make test` 承接 |

当前没有覆盖 upload/paste create flow 的真实 API/UI E2E owner。代码层回归统一由根 `make test` 承接，不能作为 E2E 证据。
