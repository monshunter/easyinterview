# UI-Design Pixel Parity Gate

> **版本**: 1.3
> **状态**: completed
> **更新日期**: 2026-05-11

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 目标

承接 D2 [002-app-shell-visual-system](../002-app-shell-visual-system/plan.md) 已落地的 ui-design 源级复刻视觉系统，把 D2 retrospective（[2026-05-07](../../../../reports/2026-05-07-frontend-shell-app-shell-visual-system-assessment.md)）列出的最高优先级 follow-up 落地：引入 Playwright + chromium，在 desktop (1440×900) / mobile (390×844) 两个 viewport 下并行加载 `frontend/dist/index.html` 与 `ui-design/index.html` golden preview，断言 DOM 锚点、computed style、bounding box、必要截图差异满足 100% 源级复刻阈值；闭环 spec.md C-9 真实浏览器 parity gate。E2E.P0.005（jsdom 范围）保留作为 fast smoke gate，不被取代。

## 2 背景

D2 plan 002 已在 vitest+jsdom 范围内验证 DOM/className/CSS variable resolution/customAccent overlay/legacy 负向/ui-design 源字面量追溯六类 parity，见
[`p0-005-app-shell-visual-system-smoke.test.tsx`](../../../../../frontend/src/app/scenarios/p0-005-app-shell-visual-system-smoke.test.tsx) 与场景 README §6。jsdom 不做 CSS 布局，不计算 bounding box，不渲染像素，所以以下维度需要真实浏览器：

- **viewport 布局**：flex / grid / sticky / responsive 表现是否与 ui-design 一致；TopBar 在 1440×900 与 390×844 下都不溢出。
- **bounding box**：核心控件（TopBar 五入口、显示控制、用户区、auth CTA、profile / settings 卡片）在两边的 `getBoundingClientRect()` 不重叠且 stays in viewport。
- **screenshot smoke / optional diff**：默认 warm/light、dark、customAccent 三种状态下，常规 gate 以非空 screenshot buffer + DOM / computed / bounding 断言证明真实浏览器可渲染；只有 baseline 可由 checkout / CI artifact 稳定取得或显式 `--update-snapshots` 维护时，才把 `toHaveScreenshot` diff 升级为 hard gate。

`ui-design/index.html` 是单文件静态原型（含 `<script src="...react...">` CDN 引用 + 内嵌 `src/*.jsx` Babel 转译），不需要构建。`frontend/dist/index.html` 由 vite build 产物。两者通过同一个 Playwright server fixture 提供，互不耦合。

2026-05-10 remediation：完整 `test:pixel-parity` 已扩展到 home / parse / jd_match / workspace 等业务屏，不能继续依赖过期的本地 hydrated workspace 前提或 `.gitignore` 排除的 screenshot baseline。常规 clean checkout gate 必须只依赖可重建的 DOM anchor、computed style、bounding box、responsive geometry 与 screenshot smoke；`toHaveScreenshot` 只能在 baseline 可由 checkout / CI artifact 稳定取得或显式 `--update-snapshots` 维护时使用。Workspace full-state pixel tests 必须从 server-bound route params 进入完整规划态，不能通过 Home recent card 的 `resume-unbound` 路径绕到 missing-resume 状态。

2026-05-11 remediation：Phase 6 登录态菜单 parity 不能只停留在 jsdom/component test。`topbar.spec.ts` 必须覆盖 authenticated user menu 的真实浏览器几何：登录后头像 chip、dropdown header / profile / settings / logout 项、desktop 右对齐、mobile viewport containment、logout 后回到非登录态。完整 `test:pixel-parity` 当前为 8 spec / 112 tests。

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior` + `frontend` + `tooling`（视觉验证基础设施 + scenario）。
- **TDD 策略**: 通过 `/implement frontend-shell/003-ui-design-pixel-parity-gate frontend` → `/tdd` 执行；每个 checklist item 先写 Playwright spec 断言（initial Red 通过 `pnpm exec playwright test --list` 找不到 spec / 找不到 selector / 截图缺失），再实现最小代码（spec / fixture / config / npm script），再让 spec 通过。Playwright server fixture、static 服务、playwright.config.ts、frontend dist build、ui-design 文件 mount 均要有断言；Playwright 安装失败必须 fail loudly 而不是 silent skip。任何 viewport / DOM / computed style / bounding box / screenshot 差异必须修正到与 ui-design 一致或先修订 `ui-design/` 真理源。
- **BDD 策略**: 需要 BDD。本 plan 引入新的真实浏览器 parity gate，必须维护 [bdd-plan](./bdd-plan.md) / [bdd-checklist](./bdd-checklist.md)，并在主 checklist 使用 `BDD-Gate:` 引用 `E2E.P0.006`。
- **替代验证 gate**: 不适用；BDD gate 是本 plan 的 pixel-level 验证入口。补充 gate 包括 D1+D2 jsdom 全量测试、`E2E.P0.005` jsdom smoke、`pnpm --filter @easyinterview/frontend build`、`make build`、active-scope 负向搜索、`make docs-check`。

## 4 实施步骤

### Phase 1: Playwright 基础设施

#### 1.1 引入 Playwright + chromium

把 `@playwright/test` 加入 `frontend/package.json` devDependencies；新增 `pnpm --filter @easyinterview/frontend test:pixel-parity` 与 `test:pixel-parity:install` script。安装步骤明确为 `pnpm exec playwright install --with-deps chromium`，离线 / 失败时给出可读 error，不允许 silent skip。Playwright 默认 reporter 设为 `list + html`。

#### 1.2 配置 playwright.config.ts

在 `frontend/playwright.config.ts` 中声明：
- `testDir = "./tests/pixel-parity"`
- 两个 project：`desktop`（viewport 1440×900）+ `mobile`（viewport 390×844 + Safari mobile UA）
- `webServer.command` 启动 `frontend/scripts/serve-pixel-parity.mjs`（同时托管 `frontend/dist` 与 `ui-design/`），`webServer.url` 设 `http://127.0.0.1:4173/health`，`reuseExistingServer = !CI`
- `expect.toHaveScreenshot` 默认阈值 `maxDiffPixels: 200, threshold: 0.05`；可在单 spec 内放宽
- `outputDir = "./.playwright-output"`，testid 截图基线放 `frontend/tests/pixel-parity/__screenshots__`

不允许把 baseline 截图签入大量二进制；`__screenshots__` 不进 git，截图 baseline 通过 `--update-snapshots` 在 CI / 本地按需生成（参见 5.1 重跑指引）。

#### 1.3 静态 server fixture

新增 `frontend/scripts/serve-pixel-parity.mjs`：node 内置 `http` 模块，无第三方依赖；监听 4173；同时挂载 `frontend/dist`（路径 `/`）与 `ui-design/`（路径 `/ui-design/`），并且暴露 `/health` 200 响应供 Playwright `webServer` 探活。要求脚本启动前先检查两个目录都存在，否则 fail loudly 提示用户先 `pnpm --filter @easyinterview/frontend build`。

### Phase 2: DOM + computed style parity

#### 2.1 TopBar DOM 锚点 + computed style

`frontend/tests/pixel-parity/topbar.spec.ts` 在两个 project 下都验证：

- `frontend/dist/index.html` 与 `ui-design/index.html` 加载后 TopBar 五入口 testid（`topbar-nav-{home,jd_match,workspace,resume_versions,debrief}`）都存在，文本一致（按 lang）。
- `getComputedStyle()` 在两边读出的 TopBar `height` / `padding` / `gap` / `border-bottom-width` / `background-color` 一致到 1px / 1 hex 容差内。
- `aria-current` / `aria-pressed` 行为一致。

#### 2.2 Auth / Profile / Settings / Placeholder DOM 锚点

`frontend/tests/pixel-parity/screens.spec.ts` 验证：

- 加载 frontend `#auth_login`、`#profile`、`#settings`、`#company_intel` 时 D2 testid（`route-*` / `ei-*` className）存在；并且 `ui-design/index.html` 在同样的 hash 路由下渲染等价 DOM 节点（不要求 testid 完全一致，但要求结构同源 / 文案语义同源）。
- 主要卡片（`ei-auth-card` / `ei-screen-card`）的 computed `padding` / `border-radius` / `border-color` 与 ui-design 对应卡片一致。

### Phase 3: Layout + bounding box parity

#### 3.1 Desktop viewport 不重叠

`frontend/tests/pixel-parity/layout.spec.ts`（desktop project）在 frontend dist 上验证：

- `app-shell-topbar` `getBoundingClientRect()` 完全在 `[0, 0, 1440, 58]` 内。
- TopBar primary nav 五个 button + display controls + user area 之间两两不重叠（`!intersects(rectA, rectB)`）。
- auth login `ei-auth-card` 与 `ei-auth-side` 在同一行排列、`right(side) <= left(card)`、`top(card) ≈ top(side)`。
- profile / settings shell 卡片在 viewport 内，不溢出右侧。

#### 3.2 Mobile viewport 响应式

`layout.spec.ts`（mobile project）验证：

- TopBar 不溢出 viewport（`right ≤ 390`）、五入口仍可达（DOM 中存在；不要求全部可见）。
- auth shell 双列在 mobile 视口里允许折叠为单列（`width(side) ≈ width(card)`），但 `route-auth_login` 元素的 `bottom` 不超过 `body.scrollHeight`。

### Phase 4: Screenshot diff

#### 4.1 默认 warm/light 截图基线

`frontend/tests/pixel-parity/screenshot.spec.ts` 在两个 project 下执行：

- 加载 `frontend` 默认 home，关闭动画（`page.addStyleTag({ content: "*, *::before, *::after { animation: none !important; transition: none !important; }" })`），等待 fontsource 字体加载完成（`document.fonts.ready`），然后获取 browser screenshot buffer 并断言非空。
- `toHaveScreenshot` 与跨 `ui-design` 像素 diff 只作为显式 baseline 维护流程使用，不作为 clean checkout 常规 PASS gate。

#### 4.2 Dark + customAccent 视觉差异对照

同 `screenshot.spec.ts` 验证：

- 切到 dark 后 frontend `--ei-color-bg-canvas` 解析为 `#16130e`、`document.body.background-color` 切到 `rgb(22, 19, 14)`。
- 激活 customAccent 后 `<html data-custom-accent="active">`、内联 `--ei-color-accent` / `--ei-color-accent-soft` 出现 oklch 值，且 base palette token 不被覆盖，避免 customAccent 静默失效。

### Phase 5: Scenario + handoff

#### 5.1 派生 E2E.P0.006 scenario

新增 `test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/` 包：`README.md`（含 Playwright 安装步骤、跑 `pnpm --filter @easyinterview/frontend test:pixel-parity` 的预期）、`data/{seed-input,expected-outcome}.md`、`scripts/{setup,trigger,verify,cleanup}.sh`。setup.sh 检查 chromium 安装与 frontend dist 是否就绪（必要时调用 build），trigger.sh 执行 Playwright，verify.sh 断言 trigger.log 包含全部 spec PASS / 0 failed，cleanup.sh 清理 setup marker。同步 `test/scenarios/e2e/INDEX.md` 增加 P0.006 行。

#### 5.2 Handoff

更新 `frontend/README.md` §2.7 D2 视觉骨架接入点的「Visual smoke 工具与 parity gate 重跑」章节，说明：

- jsdom fast smoke：`pnpm --filter @easyinterview/frontend test src/app/scenarios/p0-005-app-shell-visual-system-smoke.test.tsx`
- 真实浏览器 pixel parity gate：`pnpm --filter @easyinterview/frontend test:pixel-parity`（前置 `pnpm exec playwright install --with-deps chromium` + `pnpm --filter @easyinterview/frontend build`）
- E2E.P0.006 scenario 入口与 baseline 截图重生成方式（`--update-snapshots`）
- 截图 baseline 默认不入 git；CI 与本地各自维护

### Phase 7: Clean-checkout pixel gate remediation

#### 7.1 Screenshot smoke replaces ignored baseline hard gate

修复 `frontend/tests/pixel-parity/screenshot.spec.ts` 与 `workspace.spec.ts` 中常规 Playwright gate 对 `*-snapshots/*-darwin.png` 的依赖。默认 home、workspace empty、workspace full 的常规验证改为非空 screenshot buffer + 已有 DOM / computed style / bounding box 断言；本地 `toHaveScreenshot` baseline 维护只保留为显式 `--update-snapshots` 工作流说明，不作为 clean checkout PASS 条件。

#### 7.2 Workspace full-state route bootstrap

为 pixel harness 提供最小初始路由 bootstrap，使 Playwright 能以 server-bound `targetJobId` / `resumeVersionId` / `planId` 进入 hydrated workspace。bootstrap 只能由测试环境显式注入，不得修改 Home recent card 的 `resume-unbound` 业务语义，也不得让 synthetic id 绕过 `normalizeServerBoundId` gate。

#### 7.3 E2E.P0.006 contract refresh

更新 E2E.P0.006 README / verify 脚本与 frontend README：当前完整 pixel gate 为 8 个 spec × desktop/mobile = 112 tests；verify 必须确认 workspace spec 被执行、0 failed、retired entry 未回流，并且文档明确 clean checkout gate 不依赖 ignored screenshot baseline。

#### 7.4 Authenticated user-menu browser parity hardening

把 Phase 6 authenticated user menu 纳入 `topbar.spec.ts`，通过 mocked Auth API 在 desktop / mobile 下执行 login → avatar chip → dropdown → logout。断言 dropdown 的源码字面量、bounding box、desktop right alignment、mobile viewport containment 和 logout 回到非登录态；ui-design golden preview 加载等待必须以可见 `nav button` 为准，避免只等 `load` 时受 CDN/字体时序影响而误判。

## 5 验收标准

- spec.md C-9 验证通路全部可执行；C-8 jsdom 范围继续有效，retrospective 列出的 high-priority follow-up 闭环。
- `frontend/playwright.config.ts` 声明 desktop + mobile 两个 project，`webServer` 指向 `serve-pixel-parity.mjs`。
- `pnpm --filter @easyinterview/frontend test:pixel-parity` 在 chromium 已安装环境下全部 PASS（当前 8 个 spec × desktop/mobile = 112 tests），并且不依赖 `.gitignore` 排除的本地 screenshot baseline。
- `frontend/scripts/serve-pixel-parity.mjs` 在 dist / ui-design 缺失时 fail loudly。
- `E2E.P0.006` scenario `setup → trigger → verify → cleanup` 通过。
- D1 + D2 jsdom 全量测试不退化；`E2E.P0.001 / 002 / 004 / 005` 仍通过。
- `pnpm --filter @easyinterview/frontend build` + `make build` 通过。
- `frontend/README.md` §2.7 更新 pixel parity gate 接入点；docs-check zero drift。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| chromium 二进制下载失败 / 离线环境 | setup.sh 先检测 `~/.cache/ms-playwright/chromium-*`，缺失时提示用户运行 `pnpm exec playwright install --with-deps chromium` 并 exit 1；不在自动 trigger 中下载 |
| ui-design HTML 依赖 CDN React/Babel，CI 离线时不能加载 | server fixture 检测 `ui-design/index.html` 内的 CDN 链接列表，README 标注需要外网；若需要离线，先 vendor CDN 资源到 `ui-design/vendor/` 并修改 index.html |
| screenshot 阈值过严导致 flaky | 默认 `maxDiffPixels` 容忍字体子像素 / icon SVG 抗锯齿差异；CI 上若仍 flaky，先把 baseline 当作 informative，让 DOM + bounding box 维度作为 hard gate |
| baseline 截图入 git 膨胀 | `__screenshots__` 加入 `.gitignore`，文档说明 baseline 由 CI / 本地 `--update-snapshots` 重生成 |
| clean checkout 缺少 ignored baseline 导致常规 gate 首跑失败 | 常规 gate 使用 screenshot smoke + DOM/computed/bounding/responsive 断言；只有显式 snapshot 维护流程才运行 baseline diff |
| workspace full-state test 通过 Home recent card 进入 `resume-unbound` missing-resume 状态 | pixel harness 使用显式 server-bound route bootstrap；Home recent card 与 workspace 业务 hook 的 synthetic id 过滤保持不变 |
| Playwright server 启动慢 | `webServer.timeout = 30_000`；`reuseExistingServer = !CI` 让本地反复 run 不重启 |
| Frontend build 失败被误判为 parity 失败 | trigger.sh / setup.sh 区分「build 失败」与「test 失败」并给出明确退出码；verify.sh 只在 trigger 成功时才解析 log |
| 与 D2 jsdom 测试出现重复 / 不一致 | 文档明确分工：jsdom 范围保留 fast smoke；Playwright 范围才做 viewport / 布局 / 截图；两者不互相替代 |
| spec.md C-9 出现新的 retired 模块假设 | parity gate 中含旧 entry 负向断言（DOM 不出现 `route-welcome` / `topbar-nav-mistakes` 等），与 D2 保持一致 |
