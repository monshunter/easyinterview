# Interview Card Fusion, Action Row, and Theme Pruning 交付复盘报告

> **日期**: 2026-07-09
> **审查人**: Codex

## 1 复盘范围与成功证据

- 交付范围：修复 `frontend-workspace-and-practice/001-workspace-and-interview-context` 中面试规划列表卡片在 1 个、2 个、3 个数据量下会横向拉伸的问题；按用户补充目标把首页“最近模拟面试”和面试列表卡片融合为同一个 `MockInterviewCard` 主体；后续追加 action-row 优化：workspace 卡片点击自身进入规划，footer 改为 `立即面试` 主按钮 + 使用简历列表 trash 图标的删除按钮，Home recent 复用同一卡片动作模型但不展示删除；同步 `frontend-shell/002-app-shell-visual-system`，将主题色选项收敛为深海、梅子和自定义入口，移除暖陶、苔林的正式前端与 `ui-design/` active 原型口径。
- 已同步资产：
  - [workspace owner plan](../spec/frontend-workspace-and-practice/plans/001-workspace-and-interview-context/plan.md)、[workspace checklist](../spec/frontend-workspace-and-practice/plans/001-workspace-and-interview-context/checklist.md)、[workspace spec](../spec/frontend-workspace-and-practice/spec.md)
  - [home owner plan](../spec/frontend-home-job-picks-and-parse/plans/001-home-jd-import-and-parse/plan.md)、[home checklist](../spec/frontend-home-job-picks-and-parse/plans/001-home-jd-import-and-parse/checklist.md)、[home spec](../spec/frontend-home-job-picks-and-parse/spec.md)
  - [shell owner plan](../spec/frontend-shell/plans/002-app-shell-visual-system/plan.md)、[shell checklist](../spec/frontend-shell/plans/002-app-shell-visual-system/checklist.md)、[shell spec](../spec/frontend-shell/spec.md)
  - [workspace UI design doc](../ui-design/module-job-workspace.md)、`ui-design/` static prototype、formal frontend implementation、focused tests
- 成功证据：
  - `pnpm --filter @easyinterview/frontend test src/app/screens/home/MockInterviewCard.test.tsx src/app/screens/home/HomeRecentMocks.test.tsx src/app/screens/workspace/WorkspaceScreen.test.tsx src/app/screens/workspace/WorkspaceEmptyState.test.tsx src/app/scenarios/p0-005-app-shell-visual-system-smoke.test.tsx` 通过，5 files / 33 tests。
  - `pnpm --filter @easyinterview/frontend test src/app/theme/tokens.test.ts src/app/display/DisplayPreferencesProvider.test.tsx src/app/display/DisplayPreferencesRootWiring.test.tsx src/app/topbar/TopBar.test.tsx` 通过，4 files / 35 tests。
  - `pnpm --filter @easyinterview/frontend test src/app/topbar/TopBarVisual.test.tsx` 通过，13 tests，并断言 Warm / Forest 不再出现在主题菜单。
  - `pnpm --filter @easyinterview/frontend typecheck` 通过。
  - `node --test ui-design/ui-design-contract.test.mjs` 通过，30 tests。
  - `pnpm --filter @easyinterview/frontend build` 通过；仅保留既有 Vite chunk size warning。
  - `pnpm --filter @easyinterview/frontend test:pixel-parity tests/pixel-parity/workspace.spec.ts` 通过，desktop/mobile 32 tests。
  - `test/scenarios/e2e/p0-018-workspace-default-render/scripts/trigger.sh` 通过，7 files / 54 tests；`verify.sh` 返回 `E2E.P0.018 PASS`。
  - action-row 追加验收：`pnpm --filter @easyinterview/frontend test src/app/screens/home/MockInterviewCard.test.tsx src/app/screens/home/HomeRecentMocks.test.tsx src/app/screens/workspace/WorkspaceScreen.test.tsx src/app/screens/workspace/WorkspaceEmptyState.test.tsx` 通过，4 files / 30 tests。
  - action-row 追加验收：`node --test ui-design/ui-design-contract.test.mjs` 通过，31 tests；`pnpm --filter @easyinterview/frontend typecheck`、`pnpm --filter @easyinterview/frontend build`、`pnpm --filter @easyinterview/frontend test:pixel-parity tests/pixel-parity/workspace.spec.ts` 通过，pixel parity 32 tests。
  - `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/frontend-workspace-and-practice/plans/001-workspace-and-interview-context/context.yaml --docs-root docs --target frontend` 通过。
  - `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/frontend-home-job-picks-and-parse/plans/001-home-jd-import-and-parse/context.yaml --docs-root docs --target frontend` 通过。
  - `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/frontend-shell/plans/002-app-shell-visual-system/context.yaml --docs-root docs --target frontend` 通过。
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`、`make docs-check`、`git diff --check` 通过。
  - 浏览器验收截图已生成：`.test-output/screenshots/home-recent-fused-fixed-card.png`、`.test-output/screenshots/workspace-plan-list-fused-recent-card.png`、`.test-output/screenshots/theme-menu-ocean-plum-custom-fused-card.png`、`.test-output/screenshots/home-recent-action-card.png`、`.test-output/screenshots/workspace-plan-list-action-card.png`；Playwright parity 反查确认 workspace desktop/mobile 卡片 affordance、mini rail、action row 和主题菜单数量。

## 2 会话中的主要阻点/痛点

- 面试规划列表原先只约束“卡片存在”和信息结构，没有约束列宽上限。
  - **证据**：用户截图显示单卡横向撑满整行；实现中 grid 使用 `repeat(auto-fit, minmax(300px, 1fr))`，导致数据量少时卡片宽度随容器伸缩。
  - **影响**：需要回到 workspace owner spec / plan / checklist 补固定规格合同，并在正式前端和 `ui-design/` 原型一起改为固定最大列宽。

- Home 最近模拟面试与 workspace 面试列表曾各有一套卡片主体。
  - **证据**：用户补充目标要求“使用首页的最近模拟面试的卡片为主体，加上面试列表的功能按钮”；原 workspace 卡片已经有 footer CTA，但主体没有复用 Home recent 的 mini round rail。
  - **影响**：需要同时 reopening home owner 和 workspace owner，把 shared `MockInterviewCard` 扩展为可注入 workspace testids/footer 的主体组件，并在 `ui-design/` 双入口同步。

- 卡片 action model 在截图验收后继续变化，要求避免重新分叉组件。
  - **证据**：用户追加目标要求 `进入规划` 不再显示、点击卡片替代其功能，同时新增 `立即面试` 与删除，Home recent 只保留 `立即面试`。
  - **影响**：需要继续在同一个 `MockInterviewCard` 上扩展 primary/delete action props，并补 action propagation 与 Home no-delete 断言，避免 workspace 再次形成独立卡片。

- 主题色范围不只存在于 TopBar 菜单。
  - **证据**：暖陶、苔林同时存在于 `tokens.ts` 类型、主题 metadata、CSS 变量块、i18n 文案、display preference 测试、P0.005 smoke、`ui-design/src/primitives.jsx` 与 `ui-design/canvas.html` artboards。
  - **影响**：如果只从菜单隐藏，会留下可被存储、测试、原型或文档继续引用的 active palette，不能算完成删除。

- 浏览器截图验收需要先通过本地 mock 登录进入受保护页面。
  - **证据**：验收时必须走 `test@example.com` / `654321` / profile setup 后才能进入 workspace 页面截图。
  - **影响**：增加环境操作成本，但没有发现需要修改鉴权流程或场景环境脚本的产品缺陷。

## 3 根因归类

- 卡片几何规格缺少负向 gate。
  - **类别**：spec-plan。
  - 既有 workspace 计划已有卡片视觉合同，但没有明确“desktop 列宽最大 360px、不得因 `auto-fit` + `1fr` 被单卡撑满”的断言。

- 卡片 action affordance 属于 shared component contract，不应只在单页面测试中隐式覆盖。
  - **类别**：spec-plan。
  - Home 与 workspace 现在共享 `MockInterviewCard` 主体，但按钮数量、删除可见性和 action propagation 需要同时由 owner checklist、unit tests、prototype contract 和截图验收承接。

- 主题范围属于跨层 active contract，历史四主题口径没有单点 owner。
  - **类别**：spec-plan / README。
  - shell owner plan、frontend README、正式前端类型和 `ui-design` palette 都需要同步收敛，否则会出现菜单已删但代码和文档仍承认旧主题的 drift。

- mock 登录是受保护 workspace 的正常验收前置条件。
  - **类别**：no repo change needed。

## 4 对流程资产的改进建议

- 后续 UI 列表/卡片计划应把稳定几何规格写成可执行 gate，包括 `grid-template-columns`、`justify-content`、关键 bounding box 和负向 `1fr` 拉伸断言。
  - **落点**：spec-plan。
  - **优先级**：high。

- 后续共享卡片组件的 action 变化应直接在 shared component owner checklist 中列出三类断言：卡片主体点击、primary action stopPropagation、delete action optional visibility。
  - **落点**：spec-plan。
  - **优先级**：high。

- 后续主题或 palette 范围变更应按 active contract 删除清单执行：类型、metadata、CSS、i18n、tests、`ui-design` palette、canvas artboard 和 README 同步清理。
  - **落点**：spec-plan / README。
  - **优先级**：high。

- 对需要截图验收的受保护页面，在最终验收说明中继续显式记录登录前置和 mock 账号，避免把 auth gate 误判成页面缺陷。
  - **落点**：no repo change needed。
  - **优先级**：low。

## 5 建议优先级与后续动作

- high：若继续推进面试规划列表体验，优先补一个跨 viewport 的 pixel/browser parity gate，覆盖 1 张、2 张、3 张卡片时 Home recent 与 workspace plan list 的 desktop 列宽均保持 `360px`，mobile 则保持单列自适应。
- high：若后续继续删减 visual system scope，优先用负向搜索 gate 固定旧 theme key 不得出现在 active code、active docs、`ui-design` 原型和 smoke 场景中。
- medium：本次未创建独立 BUG 记录；原因是该问题已由两个现有 owner plan 原地承接并有测试/截图证据闭环，且当前没有真实关联 commit 可填写到 resolved Bug 记录中。
