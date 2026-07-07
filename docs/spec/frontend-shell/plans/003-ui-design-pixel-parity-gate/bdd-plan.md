# UI-Design Pixel Parity Gate BDD Plan

> **版本**: 1.5
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 Plan**: [plan](./plan.md)
**关联 Checklist**: [checklist](./checklist.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 场景矩阵

| 场景 ID | 名称 | 类别 | 验证入口 |
|---------|------|------|----------|
| `E2E.P0.006` | real-browser ui-design pixel parity gate | visual parity + regression | `test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/` |

## 2 场景明细

### E2E.P0.006 real-browser ui-design pixel parity gate

| Given | When | Then |
|-------|------|------|
| `frontend/dist` has been built; `ui-design/index.html` can load; Chromium is installed; current App Shell and migrated screens are available | Scenario trigger runs `pnpm --filter @easyinterview/frontend test:pixel-parity`, which starts `serve-pixel-parity.mjs` and executes the desktop/mobile Playwright projects | Current 13 parity specs pass in both projects; DOM anchors, computed styles, bounding boxes, responsive geometry, dark/customAccent and screenshot smoke are covered; current TopBar/user menu and current business screens render; non-current route/module entries stay negative-only; the gate does not require ignored local screenshot baselines |

## 3 执行入口

```bash
test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/scripts/setup.sh && test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/scripts/trigger.sh && test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/scripts/verify.sh && test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/scripts/cleanup.sh
```

## 4 AC 映射

| spec AC / decision | 覆盖场景 |
|--------------------|----------|
| C-8 UI source structure parity | `E2E.P0.006` |
| C-9 real-browser pixel parity gate | `E2E.P0.006` |
| C-10 authenticated user-menu browser parity | `E2E.P0.006` |
| C-13 non-current route negative regression | `E2E.P0.006` |
