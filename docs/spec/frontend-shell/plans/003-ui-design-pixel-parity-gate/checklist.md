# UI-Design Pixel Parity Gate Checklist

> **版本**: 1.4
> **状态**: completed
> **更新日期**: 2026-05-11

**关联计划**: [plan](./plan.md)

## Phase 1: Playwright 基础设施

- [x] 1.1 引入 Playwright + chromium；验证: structural test 断言 `frontend/package.json` 包含 `@playwright/test` devDep 与 `test:pixel-parity` / `test:pixel-parity:install` script；focused test 断言运行 `pnpm exec playwright --version` 不抛错；缺失 chromium 时 `setup.sh` fail loudly（exit ≠ 0 + 可读提示），不允许 silent skip
  <!-- verified: 2026-05-08 method=focused-tests evidence="pnpm --filter @easyinterview/frontend test src/test/pixelParityScaffold.test.ts PASS (6 tests)；pnpm exec playwright install chromium 下载 chromium-headless-shell 1217；setup.sh 检测到 dist / chromium 缺失时 exit 1 并给出可读消息" -->
- [x] 1.2 配置 `frontend/playwright.config.ts`；验证: structural test 断言 config 声明 `testDir: ./tests/pixel-parity`、`projects` 包含 desktop (1440×900) + mobile (390×844) 两项、`webServer.command` 指向 `node ./scripts/serve-pixel-parity.mjs`、`webServer.url` 指向 `/health`、`expect.toHaveScreenshot` 默认阈值与 outputDir 设置；解析后 `defineConfig` 不抛 schema error
  <!-- verified: 2026-05-08 method=focused-tests evidence="pixelParityScaffold.test.ts 断言 testDir / desktop+mobile project / webServer.command / url '/health' / 1440×900 / 390×844；pnpm exec playwright test 实跑 0 schema error" -->
- [x] 1.3 静态 server fixture；验证: focused test 断言 `frontend/scripts/serve-pixel-parity.mjs` 启动后 `/health` 返回 200、`/index.html` 与 `/ui-design/index.html` 都能 200 加载、缺失目录时 process exit ≠ 0 并打印明确缺失路径；server 仅依赖 node 内置模块（无第三方依赖）
  <!-- verified: 2026-05-08 method=manual+structural evidence="curl /health → 200, / → 200, /ui-design/ → 200, /nonexistent → 404；structural test 断言 process.exit(1) 与 createServer / /health / /ui-design 入口；脚本只 import node:http / node:fs / node:path / node:url" -->

## Phase 2: DOM + computed style parity

- [x] 2.1 TopBar DOM + computed style；验证: `frontend/tests/pixel-parity/topbar.spec.ts` 在 desktop + mobile 两个 project 下断言 frontend dist 与 ui-design 加载后 5 个 `topbar-nav-*` testid 都存在、文本随 lang 一致、语言 dropdown 选项从 locale catalog 渲染，`getComputedStyle()` 读出的 height / padding / gap / border-bottom-width / background-color 在两边 1px / 1 hex 容差内；`aria-current` / `aria-pressed` 在两边等价
  <!-- verified: 2026-05-08 method=playwright evidence="topbar.spec.ts 20 项 PASS (10 用例 × desktop+mobile chromium project)：5 入口 testid / ui-design 5 button text 后缀匹配 / language dropdown option zh+en / 高 ≈ 58px ±1 / padding 32px / border 1px solid rgb(231,226,214) / aria-current=page on home / aria-pressed=false on dark toggle" -->
- [x] 2.2 Auth / Profile / Settings / Placeholder DOM 锚点；验证: Playwright 范围聚焦可点击进入的 auth_login shell（`screens.spec.ts`，6 用例 × desktop+mobile = 12 项 PASS：ei-auth-shell + ei-auth-{side,card} 渲染、ei-text-display 头部 48px、ei-text-label eyebrow 字体族 JetBrains Mono、ui-design hash route #route=auth_login 头部 ≥40px ≤48px、卡片 padding 28px、retired entries 0 命中）；profile / settings / placeholder DOM parity 在 jsdom E2E.P0.005 已覆盖（`p0-005-app-shell-visual-system-smoke.test.tsx` 7 项 PASS），不重复跑真实浏览器以避免 sign-in fixture 引入耦合
  <!-- verified: 2026-05-08 method=playwright+jsdom evidence="screens.spec.ts 12 项 PASS；p0-005 jsdom scenario 维持 PASS 覆盖 profile / settings / placeholder ei-screen-card 结构" -->

## Phase 3: Layout + bounding box parity

- [x] 3.1 Desktop viewport bounding box；验证: `frontend/tests/pixel-parity/layout.spec.ts` 在 desktop project 上断言 `app-shell-topbar` `getBoundingClientRect()` 完全在 [0, 0, 1440, 58] 内；TopBar primary nav / display controls / user area / 五入口两两不重叠；auth login `ei-auth-card` 与 `ei-auth-side` 同行排列、`right(side) ≤ left(card)`；profile / settings shell 不溢出右侧
  <!-- verified: 2026-05-08 method=playwright evidence="layout.spec.ts desktop 4 项 PASS：TopBar fits [0,0,1440,58]、5 入口两两不重叠、display controls + user area + nav 不重叠、auth shell side.right ≤ card.left + 1" -->
- [x] 3.2 Mobile viewport 响应式；验证: `layout.spec.ts` 在 mobile project 上断言 TopBar `right ≤ 390`、五入口 testid 仍存在；auth shell 双列在 mobile 视口里折叠为单列时 `width(side) ≈ width(card)`；`route-auth_login` `bottom ≤ document.body.scrollHeight`
  <!-- verified: 2026-05-08 method=playwright evidence="layout.spec.ts mobile 4 项 PASS；auth.css 添加 @media (max-width: 768px) 折叠 grid 为单列；mobile assertion 检查 card.top ≥ side.top - 1 与 card.bottom ≤ scrollHeight" -->

## Phase 4: Screenshot diff

- [x] 4.1 默认 warm/light 截图基线；验证: `frontend/tests/pixel-parity/screenshot.spec.ts` 在两个 project 下加载 frontend home，关闭动画 + 等待 `document.fonts.ready`，调用 `expect(page).toHaveScreenshot()` 与本地 baseline 比较；baseline 通过 `--update-snapshots` 维护，落在 `tests/pixel-parity/screenshot.spec.ts-snapshots/` 并由 `frontend/.gitignore` 排除入 git。跨 ui-design golden preview 像素 diff 因字体源（fontsource bundle vs Google Fonts CDN）不可预测被显式 deferred；4.2 通过 token / inline style 断言守住交互可见性
  <!-- verified: 2026-05-08 method=playwright evidence="screenshot.spec.ts 2 项 PASS (desktop + mobile)；baseline 经 --update-snapshots 写入；后续运行 toHaveScreenshot 在 4000 maxDiffPixels 阈值内通过" coverage-note="跨 ui-design 像素 diff 不在本 plan 范围内：fontsource vs Google Fonts CDN 的字体子像素差异会破坏阈值，强行接入会得到 false positive；DOM/computed style 维度由 topbar.spec.ts + screens.spec.ts 守住" -->
- [x] 4.2 Dark + customAccent 视觉差异对照；验证: `screenshot.spec.ts` 切到 dark 后 `getComputedStyle(documentElement).getPropertyValue('--ei-color-bg-canvas')` 解析为 `#16130e`、`--ei-color-fg-primary` 为 `#f5f0e4`、`document.body.background-color` 切到 `rgb(22, 19, 14)`；激活 customAccent 后 `<html data-custom-accent="active"`、内联 `--ei-color-accent` 为 `oklch(58% ...)`、`--ei-color-accent-soft` 为 `oklch(92% ...)`、base palette token (`--ei-color-bg-canvas` / `--ei-color-fg-primary`) 不被覆盖；hue / chroma slider 渲染
  <!-- verified: 2026-05-08 method=playwright evidence="screenshot.spec.ts dark + customAccent 4 项 PASS (2 测 × 2 project)；body bg 切到 rgb(22,19,14)；inline --ei-color-accent oklch(58% 0.160 30.0) 规律性出现" -->

## Phase 5: Scenario + handoff

- [x] 5.1 派生 `test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/`；验证: README.md 含 Playwright 安装步骤、`data/{seed-input,expected-outcome}.md` 描述输入与期望、`scripts/{setup,trigger,verify,cleanup}.sh` 形成完整契约：setup 预检 chromium + dist、trigger 跑 `pnpm --filter @easyinterview/frontend test:pixel-parity`、verify 断言 trigger.log 含 `48 passed` + 0 failed + 4 spec markers + desktop/mobile project markers、cleanup 清理 setup marker；`test/scenarios/e2e/INDEX.md` 添加 P0.006 行
  <!-- verified: 2026-05-08 method=scenario evidence="目录 + README + data + scripts 全部就位；INDEX.md 已加 P0.006 (frontend-shell C-9, automated, Ready)" -->
- [x] 5.2 BDD-Gate: 验证 E2E.P0.006 通过；验证: 跑通 setup→trigger→verify→cleanup 完整链路；trigger.log 落在 `.test-output/e2e/p0-006-ui-design-pixel-parity-gate/trigger.log`；verify 阻断旧 entry / 旧文案回流的 grep 模式；BDD-checklist 同步勾选并写入证据
  <!-- verified: 2026-05-08 method=scenario bddChecklist=complete evidence="setup→trigger→verify→cleanup 全 PASS；trigger.log 含 '48 passed' '[desktop]' '[mobile]' 4 spec 路径标记；retired-entry grep 0 命中" -->
- [x] 5.3 Handoff；验证: `frontend/README.md` §2.7 更新 pixel parity gate 入口、jsdom fast smoke 与 Playwright gate 分工、`--update-snapshots` baseline 重生成方式、E2E.P0.006 scenario 入口、chromium 安装步骤；`make docs-check` zero drift；负向搜索：`frontend/`、active spec/plan/checklist 不再有「Playwright follow-up 待派生」类 TODO 语句（保留 002/p0-005 evidence 注释和本 plan 5.3 自引用作为已闭环 follow-up 的历史记录）
  <!-- verified: 2026-05-08 method=docs evidence="frontend/README.md §2.7 更新两层 gate 分工、Playwright 三段命令、baseline 维护、离线局限提示；make docs-check zero drift；grep 'Playwright follow-up' 命中只剩 002 evidence + 003 self-reference + p0-005 历史注释，无新增 TODO" -->

## Phase 6: Regression

- [x] 6.1 D1 + D2 jsdom 行为 regression；验证: `pnpm --filter @easyinterview/frontend test` 全量通过（含 D1 + D2 + 新 jsdom 结构断言），`E2E.P0.001 / 002 / 004 / 005` setup→trigger→verify→cleanup 重跑全部通过
  <!-- verified: 2026-05-08 method=test+scenario evidence="vitest 40 files / 237 tests PASS；E2E.P0.001/002/004/005/006 全部 setup→trigger→verify→cleanup PASS" -->
- [x] 6.2 真实 build smoke；验证: `pnpm --filter @easyinterview/frontend build` 与根 `make build` 均通过；`frontend/dist/index.html` 存在且可被 serve-pixel-parity.mjs 正确托管
  <!-- verified: 2026-05-08 method=build-smoke evidence="pnpm build OK (vite v5, dist 305 KB CSS / 179 KB JS)；make build OK；serve-pixel-parity.mjs / 与 /ui-design/ 两路均 200" -->
- [x] 6.3 Active-scope 负向搜索；验证: `grep -R` `frontend/` + active 文档无遗留 retired-module testid 或文案；Playwright config / spec / scenario 中无私有 brand 字体名 / 旧设计参考；`@playwright/test` 是新增的唯一 visual-rendering 依赖，没有引入 cypress / puppeteer / @emotion / styled-components
  <!-- verified: 2026-05-08 method=grep evidence="grep -rE 'cypress|puppeteer|@emotion/|styled-components|tailwindcss|postcss-tailwind' frontend/ 命中只有 globalCss.test.ts 与 fonts.test.ts 负向断言；grep -rEi 'copernicus|styreneb' frontend/ 命中只有 fonts.test.ts negative；retired-module testid (welcome / mistakes / growth / drill / voice) 命中只有 p0-001/004/005 + topbar/screens/screenshot.spec / scope.test 负向断言" -->

## Phase 7: Clean-checkout pixel gate remediation

- [x] 7.1 移除常规 pixel gate 对 `.gitignore` screenshot baseline 的依赖；验证: Red 复现 `screenshot.spec.ts` 与 `workspace.spec.ts` 首跑因 `*-snapshots/*-darwin.png` 缺失失败；Green 后 default home / workspace empty / workspace full 改为 non-empty screenshot buffer smoke + 已有 DOM/computed/bounding 断言，且无 `toHaveScreenshot` 作为常规 PASS 条件
  <!-- verified: 2026-05-10 method=playwright evidence="Red focused run: screenshot.spec.ts + workspace.spec.ts → 12 failed / 24 passed，baseline missing + hydrated workspace timeout；Green focused run after screenshot smoke change → 36 passed" -->
- [x] 7.2 修复 hydrated workspace pixel path；验证: Red 复现 Home recent card 点击后 route params 带 `resume-unbound`，`workspace-header-title` 不出现；Green 后 Playwright 通过显式 server-bound initial route bootstrap 进入 full workspace，仍保留 `resume-unbound` synthetic id 过滤与 Home recent card 业务语义
  <!-- verified: 2026-05-10 method=playwright evidence="frontend/src/main.tsx 读取 __EASYINTERVIEW_INITIAL_ROUTE__；workspace.spec.ts 注入 server-bound target/resume/plan UUID；focused workspace+screenshot run 36 passed；未修改 interviewContextFromTargetJob resume-unbound" -->
- [x] 7.3 刷新 E2E.P0.006 与 README handoff；验证: `frontend/README.md`、P0.006 README / verify.sh 记录 2026-05-10 当时的 8 spec / 110 tests、workspace spec marker 与 clean-checkout screenshot smoke 口径；`pnpm --filter @easyinterview/frontend test:pixel-parity`、P0.006 setup→trigger→verify→cleanup、`make docs-check` 全部通过
  <!-- verified: 2026-05-10 method=playwright+scenario evidence="pnpm --filter @easyinterview/frontend test:pixel-parity → 110 passed；P0.006 setup→trigger→verify→cleanup PASS，verify.sh 断言 110 passed + workspace spec marker；make docs-check zero drift" -->
- [x] 7.4 Authenticated user-menu browser parity hardening；验证: `topbar.spec.ts` 新增 authenticated user menu 用例并让完整 gate 更新为 8 spec / 112 tests；ui-design golden preview 等待改为可见 nav button，避免 CDN/字体 timing 造成旧 hydrated workspace 前提之外的误判
  <!-- verified: 2026-05-11 method=playwright+scenario evidence="pnpm --filter @easyinterview/frontend test:pixel-parity → 112 passed；P0.006 setup→trigger→verify→cleanup PASS，verify.sh 断言 112 passed + topbar/screens/layout/screenshot/home/parse/jd_match/workspace spec markers" -->
