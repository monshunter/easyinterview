# App Shell Visual System BDD Checklist

> **版本**: 1.7
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.005 App Shell visual smoke

- [x] 场景目录存在：`test/scenarios/e2e/p0-005-app-shell-visual-system-smoke/`
- [x] 测试数据覆盖未登录用户、默认 runtime config、`ocean/light` 初始偏好、dark、custom accent、auth route、settings route 和通用 screen shell route。
- [x] `trigger.sh` 运行 `src/app/scenarios/p0-005-app-shell-visual-system-smoke.test.tsx`，验证 DOM 锚点、className、`:root[data-theme][data-mode]` computed variables、`customAccent` inline overlay 和 route alias negative checks。
- [x] `verify.sh` 校验 trigger log、Vitest 文件和通过结果。
- [x] `cleanup.sh` 清理 `.test-output/e2e/p0-005-app-shell-visual-system-smoke/`。
  <!-- verified: 2026-07-07 method=scenario evidence="E2E.P0.005 setup/trigger/verify/cleanup PASS; trigger ran src/app/scenarios/p0-005-app-shell-visual-system-smoke.test.tsx PASS (7 tests); existing C.UTF-8 locale warning only" -->
- [x] 场景说明与 expected outcome 使用当前 ocean/plum token，Warm / Forest 仅为负向菜单断言，并把 browser parity 直接交给当前 E2E.P0.006。
  <!-- verified: 2026-07-10 method=scenario-assets-and-runner evidence="P0.005 asset contract and executable smoke pass 8 tests; setup/trigger/verify/cleanup pass with current log markers; frontend-shell/002 owner suite passes 95 tests." -->

## Regression 场景

- [x] `E2E.P0.001` 默认首页与三入口 Shell 保持通过。
- [x] `E2E.P0.002` 登录打断后恢复原业务动作保持通过。
- [x] `E2E.P0.004` App Shell 中英语言切换保持通过。

## Phase 19 minimal CustomAccentPicker

- [x] `E2E.P0.005` 只通过 hue/saturation 激活 custom accent，并证明 preview/value/reset DOM、双语 reset key、`onClear` / `active` props 零引用。
  <!-- verified: 2026-07-14 method=p0-005-minimal-accent evidence="Scenario changes hue to 120 and saturation to 0.205, observes exact root overlay values, and asserts no clear testid or reset copy. Production/prototype source searches return no preview/value/reset, onClear caller or active prop residue; setup/trigger/verify/cleanup pass with 8/8 Vitest." -->
- [x] `E2E.P0.005` 证明选择 Ocean 或 Plum 会退出 custom accent，且 dark/language/auth/settings/screen-shell 回归不变。
  <!-- verified: 2026-07-14 method=p0-005-theme-exit-regression evidence="Scenario proves Ocean and Plum each remove data-custom-accent and both inline tokens; the same 8-test run covers dark token resolution, zh/en round trip, auth shell, settings shell, route shell and old-route negatives. All four scenario phases pass and cleanup removes the setup marker." -->
- [x] `E2E.P0.006` 在 1440 desktop 与 390 mobile 捕获主题菜单，断言 picker DOM、computed style、bounding box、viewport containment 和 screenshot parity。
- [x] `E2E.P0.006` 证明删除旧区域后无空白占位、menu 溢出或点击目标退化，并记录 fresh artifact/source hashes。
  <!-- verified: 2026-07-14 evidence="P0.006 setup/trigger/verify/cleanup PASS with 170/170 assertions and fresh desktop/mobile DOM, style, bbox, viewport and screenshot artifacts; old preview/reset/value region is absent without overflow." -->
