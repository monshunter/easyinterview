# App Shell Visual System BDD Checklist

> **版本**: 1.5
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.005 App Shell visual smoke

- [x] 场景目录存在：`test/scenarios/e2e/p0-005-app-shell-visual-system-smoke/`
- [x] 测试数据覆盖未登录用户、默认 runtime config、`ocean/light` 初始偏好、dark、custom accent、auth route、settings route 和通用 screen shell route。
- [x] `trigger.sh` 运行 `src/app/scenarios/p0-005-app-shell-visual-system-smoke.test.tsx`，验证 DOM 锚点、className、`:root[data-theme][data-mode]` computed variables、`customAccent` inline overlay 和 route alias negative checks。
- [x] `verify.sh` 校验 trigger log、Vitest 文件和通过结果。
- [x] `cleanup.sh` 清理 `.test-output/e2e/p0-005-app-shell-visual-system-smoke/`。
  <!-- verified: 2026-07-07 method=scenario evidence="E2E.P0.005 setup/trigger/verify/cleanup PASS; trigger ran src/app/scenarios/p0-005-app-shell-visual-system-smoke.test.tsx PASS (7 tests); existing C.UTF-8 locale warning only" -->

## Regression 场景

- [x] `E2E.P0.001` 默认首页与三入口 Shell 保持通过。
- [x] `E2E.P0.002` 登录打断后恢复原业务动作保持通过。
- [x] `E2E.P0.004` App Shell 中英语言切换保持通过。
