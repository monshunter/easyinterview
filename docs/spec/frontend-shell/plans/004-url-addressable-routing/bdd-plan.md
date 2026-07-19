# URL Addressable Routing BDD Plan

> **版本**: 1.13
> **状态**: completed
> **更新日期**: 2026-07-19

**关联 Plan**: [plan](./plan.md)

## 行为合同

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.SHELL.ROUTING.001` | 用户持有受支持、受保护或不支持的 URL | 首次打开、刷新、Back/Forward 或等待 session resolve | 恢复同一业务页面和允许参数，或进入 auth guard / safe fallback；所有 canonical route 保留共享 TopBar，面试上下文 route 高亮“面试”；业务数据不写入 URL | `frontend/src/app/routeUrl.test.ts` + `frontend/src/app/AppRoutingHistory.test.tsx`，由根 `make test` 承接 |

当前没有真实 API/UI E2E owner。路由单元测试、source contract 和静态浏览器检查属于代码层验证，阶段回归统一由根 `make test` 承接，不能作为 E2E 证据。
