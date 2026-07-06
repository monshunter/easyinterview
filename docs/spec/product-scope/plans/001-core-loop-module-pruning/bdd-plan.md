# Core Loop Module Pruning BDD Plan

> **版本**: 1.2
> **状态**: active
> **更新日期**: 2026-07-06

> Reopened with product-scope/001 Phase 6 document deletion cleanup; existing D-22 BDD matrix remains the behavior gate while active docs are reconciled against current truth sources.

## Phase 1: App Shell 与旧入口负向

| 场景 ID | 类型 | 场景 | Given | When | Then | 验证入口 |
|---------|------|------|-------|------|------|----------|
| E2E.P0.001 | primary | 默认首页只暴露核心三入口 | 用户打开默认 App | 查看 TopBar 与已登录用户菜单 | 一级导航仅有首页、模拟面试、简历；用户菜单仅有设置与隐私、退出登录；无复盘、用户画像入口 | `test/scenarios/e2e/p0-001-default-home-shell/` |
| E2E.P0.088 | regression/legacy-negative | canonical route 不再包含 debrief/profile | 用户访问 canonical route 集 | 深链到 workspace/practice/generating/report/resume-versions 与旧 `/debrief`、`/profile` | 核心 route 保持 safe params；旧路径归一到当前核心入口，不渲染旧页面 | `test/scenarios/e2e/p0-088-url-addressable-routing-canonical/` |
| E2E.P0.090 | regression/legacy-negative | hash legacy route 不 materialize retired modules | 用户输入 `#route=debrief`、`#route=debrief_full`、`#route=profile` | App 解析 route | 不生成旧 route state，不出现 retired testid 或 screen title | `test/scenarios/e2e/p0-090-url-routing-hash-legacy-negative/` |
| E2E.P0.102 | failure/recovery | 未登录保护路由不把旧模块当业务目标 | 未登录用户直接打开旧业务 route | App 执行 route/auth gate | 不创建 pendingAction 到 `debrief` 或 `profile`；核心面试 route 仍按登录 gate 工作 | `test/scenarios/e2e/p0-102-auth-gated-interview-routes/` |

## Phase 2: 核心闭环保留

| 场景 ID | 类型 | 场景 | Given | When | Then | 验证入口 |
|---------|------|------|-------|------|------|----------|
| E2E.P0.098 | primary | API-level 核心闭环不依赖复盘 / 画像 | 用户已有或创建简历并导入 JD | 完成 practice session、生成 report、创建 next_round practice plan | 链路通过 TargetJob / Resume / Practice / Report 完成；无 debrief/profile API、表或 shared event 参与 | `test/scenarios/e2e/p0-098-full-funnel-import-to-next-round-journey/` |
| E2E.P0.099 | primary | Fullstack UI 核心闭环不出现复盘 / 画像入口 | 用户在真实前后端 UI 中完成导入到报告 | 点击复练当前轮或进入下一轮 | 用户能进入对应新 session；TopBar 和用户菜单始终无复盘 / 用户画像 | `test/scenarios/e2e/p0-099-full-funnel-fullstack-ui-journey/` |

## Phase 3: Retired 场景删除矩阵

| Retired 场景 | 处理 | 替代覆盖 |
|--------------|------|----------|
| E2E.P0.060-E2E.P0.064 | 删除 backend-debrief 正向场景目录和 INDEX 行 | OpenAPI / backend / migration / shared negative gates |
| E2E.P0.065-E2E.P0.069 | 删除 frontend-debrief 正向场景目录和 INDEX 行 | E2E.P0.001 / P0.088 / P0.090 / P0.099 |
| E2E.P0.071, E2E.P0.073 | 删除 practice debrief-derived 正向场景目录和 INDEX 行 | practice focused tests + E2E.P0.098 |
| E2E.P0.091-E2E.P0.093 | 删除 backend-profile 正向场景目录和 INDEX 行 | account settings / auth scenarios remain via E2E.P0.003 / P0.101 / P0.102 |
