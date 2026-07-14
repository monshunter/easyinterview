# Resume Asset, Parse and Listing BDD Plan

> **版本**: 1.16
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Plan**: [plan](./plan.md)

## 行为合同

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.RESUME.ASSET.001` | 用户提交 paste 或已登记 upload；也可能重复提交、解析失败或跨用户读取 | 注册、解析、重试或读取 resume | 持久化 user-scoped asset、异步状态与一致 list/detail 投影；异常保持幂等、隔离和可恢复错误 | `backend/internal/resume/service_test.go` + `backend/internal/resume/jobs/parse_test.go`，由根 `make test` 承接 |

当前没有覆盖 resume register/parse/list/detail 的真实 API/UI E2E owner。代码层回归统一由根 `make test` 承接，不能作为 E2E 证据。
