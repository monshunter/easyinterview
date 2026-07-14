# Core Loop Module Pruning BDD Plan

> **版本**: 1.5
> **状态**: completed
> **更新日期**: 2026-07-14

> Product-scope/001 Phase 7 的 Parse command / Workspace read 分路与最小 custom-accent 控制已按当前源码验证完成；既有 D-22 completed 证据继续有效。

## Phase 1: App Shell 与范围外入口负向

| 场景 ID | 类型 | 场景 | Given | When | Then | 验证入口 |
|---------|------|------|-------|------|------|----------|
| E2E.P0.001 | primary | 默认首页只暴露核心三入口 | 用户打开默认 App | 查看 TopBar 与已登录用户菜单 | 一级导航仅有首页、模拟面试、简历；用户菜单仅有设置与隐私、退出登录；无复盘、用户画像入口 | `test/scenarios/e2e/p0-001-default-home-shell/` |
| E2E.P0.088 | boundary/privacy/out-of-scope-negative | canonical route 保持当前模块与最小 safe params | 用户 direct/reload/back-forward 访问 canonical route 集或范围外 `/debrief`、`/profile` | App 解析 Parse/Workspace/Reports 与范围外路径 | Parse/Workspace 只保留 `targetJobId`，Workspace 可直达详情并丢弃 `planId/resumeId`；范围外路径归一到当前核心入口且不渲染范围外页面 | `test/scenarios/e2e/p0-088-url-addressable-routing-canonical/` |
| E2E.P0.090 | regression/non-current-negative | 旧 hash/query 不 materialize 混合页面或范围外模块 | 用户输入旧 Parse report params、Workspace extra params 或 `#route=debrief/profile` | App canonicalize route | ready detail/report entry 不出现在 Parse；Workspace targetJobId 可进入只读详情；范围外 route state/testid/screen 不生成 | `test/scenarios/e2e/p0-090-url-routing-hash-out-of-scope-negative/` |
| E2E.P0.102 | failure/recovery | 未登录保护路由不把范围外模块当业务目标 | 未登录用户直接打开范围外业务 route | App 执行 route/auth gate | 不创建 pendingAction 到 `debrief` 或 `profile`；核心面试 route 仍按登录 gate 工作 | `test/scenarios/e2e/p0-102-auth-gated-interview-routes/` |

## Phase 2: 核心闭环保留

| 场景 ID | 类型 | 场景 | Given | When | Then | 验证入口 |
|---------|------|------|-------|------|------|----------|
| E2E.P0.098 | primary | API-level 核心闭环不依赖复盘 / 画像 | 用户已有或创建简历并导入 JD | 完成 practice session、生成 report、创建 next_round practice plan | 链路通过 TargetJob / Resume / Practice / Report 完成；无 debrief/profile API、表或 shared event 参与 | `test/scenarios/e2e/p0-098-full-funnel-import-to-next-round-journey/` |
| E2E.P0.099 | primary | Fullstack UI 核心闭环不出现复盘 / 画像入口 | 用户在真实前后端 UI 中完成导入到报告 | 点击复练当前轮或进入下一轮 | 用户能进入对应新 session；TopBar 和用户菜单始终无复盘 / 用户画像 | `test/scenarios/e2e/p0-099-full-funnel-fullstack-ui-journey/` |

## Phase 3: Out-of-scope 场景删除矩阵

| Out-of-scope 场景 | 处理 | 替代覆盖 |
|--------------|------|----------|
| E2E.P0.060-E2E.P0.064 | 删除 backend-debrief 正向场景目录和 INDEX 行 | OpenAPI / backend / migration / shared negative gates |
| E2E.P0.065-E2E.P0.069 | 删除 frontend-debrief 正向场景目录和 INDEX 行 | E2E.P0.001 / P0.088 / P0.090 / P0.099 |
| E2E.P0.071, E2E.P0.073 | 删除 practice debrief-derived 正向场景目录和 INDEX 行 | practice focused tests + E2E.P0.098 |
| E2E.P0.091-E2E.P0.093 | 删除 backend-profile 正向场景目录和 INDEX 行 | account settings / auth scenarios remain via E2E.P0.003 / P0.101 / P0.102 |

## Phase 4: Parse / Workspace 分路、恢复导航与主题控制

| 场景 ID | 类型 | 场景 | Given | When | Then | 验证入口 |
|---------|------|------|-------|------|------|----------|
| E2E.P0.015 | primary/state transition | 新 JD import 进入纯 Parse progress | Home 已选择 ready Resume 并提交 JD | `importTargetJob` 返回 target，Parse 读取 queued/processing，随后变为 ready | 只有 queued/processing 展示解析进度；ready 以 replace 进入 `/workspace?targetJobId=...`，不渲染 ready Parse detail | `test/scenarios/e2e/p0-015-jd-import-and-parse/` |
| E2E.P0.016 | primary/read-only | ready handoff 与 Workspace 报告入口 | TargetJob 已 ready | import handoff 或直接打开只读详情 | 详情位于 `/workspace?targetJobId=...`；报告入口只在 Workspace，Parse 无报告入口/嵌入列表/动画 | `test/scenarios/e2e/p0-016-parse-confirm-to-workspace/` |
| E2E.P0.018 | alternate/direct read | ready 规划卡片直达详情 | Home / Workspace 列表存在 ready 卡片 | 点击卡片或直接打开 Workspace target URL | 只读调用 `getTargetJob`，不进入 Parse、不 import；`/workspace` 无 target 时仍是列表 | `test/scenarios/e2e/p0-018-workspace-default-render/` |
| E2E.P0.046 | failure/recovery | Practice terminal recovery 回当前规划 | 会话存在可信 `targetJobId` 且进入 terminal failure | 用户点击唯一恢复 CTA | 直达 `/workspace?targetJobId=...`，不进入 Parse、不显示 retry 或重复 error banner | `test/scenarios/e2e/p0-046-practice-text-loop-failure-and-recovery/` |
| E2E.P0.058 | failure/recovery | Report/Generating 可信与不可信 Back 分层 | report response 有可信 target，或首次加载没有可信 identity | 用户从 terminal/missing 状态返回 | 可信 Report/Generating 返回 ReportsScreen；不可信回 `/workspace`，不构造 target 或进入 Parse | `test/scenarios/e2e/p0-058-report-failure-and-missing-session/` |
| E2E.P0.059 | primary/navigation | Reports Back 回 Workspace detail | ReportsScreen 已由 Workspace 详情进入且 target identity 可信 | 用户点击 Back | 返回 `/workspace?targetJobId=...`；报告详情仍按既有层级返回 ReportsScreen | `test/scenarios/e2e/p0-059-report-pixel-parity-i18n-and-out-of-scope-negative/` |
| E2E.P0.005 | alternate/preference | Custom Accent 最小 DOM | 用户打开主题菜单并选择 custom | 查看 custom 控件或选择 Ocean/Plum | custom 只有 hue/saturation；无 preview/value/reset；预设选择退出 custom | `test/scenarios/e2e/p0-005-app-shell-visual-system-smoke/` |
| E2E.P0.006 | responsive/parity | Ocean/Plum 与 custom desktop/mobile parity | 1440/390 viewport 与 light/dark 状态 | 对比原型和正式 TopBar | DOM、computed style、bounding box 与截图一致，且不存在被删除的 custom 冗余控件 | `test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/` |
