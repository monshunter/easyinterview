# Practice Plan and Session Orchestration BDD Plan

> **版本**: 2.6
> **状态**: active
> **更新日期**: 2026-07-15

**关联 Plan**: [plan](./plan.md)

## 行为合同

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.PRACTICE.PLAN.001` | 用户具有 TargetJob 与 resume 上下文；创建或启动也可能遇到 provider / persistence 失败 | 创建、复用或重试 practice plan / session | 只复用身份完全匹配的资源；成功保持幂等与隔离，失败不产生重复事实或泄露原始材料 | `backend/internal/practice/conversation_service_test.go` + `backend/internal/store/practice/create_plan_integration_test.go`，由根 `make test` 承接 |

当前没有覆盖 plan 创建、session start 或 opening message 的真实 API/UI E2E owner。数据库 integration 与单元测试属于代码层 gate，阶段回归统一由根 `make test` 承接。

## Phase 8 BDD 适用性

`listPracticeSessions` 没有当前用户可见入口或行为流，因此 Phase 8 不新增 Behavior ID、BDD 场景或 E2E 场景，也不把 inventory、fixture、codegen、mock 或 source negative 包装成 BDD。删除公共列表的验证由主 checklist Phase 8 与 test checklist 的 OpenAPI exact inventory/diff、fixture/codegen/mock parity、focused handler/source negatives 和根 `make test` 承接；既有 `BDD.PRACTICE.PLAN.001` 继续只回归 plan/session 创建与恢复合同。
