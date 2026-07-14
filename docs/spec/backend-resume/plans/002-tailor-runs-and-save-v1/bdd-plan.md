# Resume Tailor Runs and Save BDD Plan

> **版本**: 1.10
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Plan**: [plan](./plan.md)

## 行为合同

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.RESUME.TAILOR.001` | 用户选择 resume 与 TargetJob；也可能遇到重复请求、provider 失败、过期 run 或跨用户访问 | 创建 tailor run、等待、重试、读取或保存新 resume | 持久化 provenance、异步状态与独立新资产；异常保持幂等、隔离、隐私与可恢复错误 | `backend/internal/resume/service_test.go` + `backend/internal/resume/jobs/tailor_test.go`，由根 `make test` 承接 |

当前没有覆盖 tailor/create/save 流程的真实 API/UI E2E owner。代码层回归统一由根 `make test` 承接，不能作为 E2E 证据。
