# E2E.P0.006 UI-Design Pixel Parity Gate

> **场景 ID**: E2E.P0.006
> **执行方式**: automated
> **隔离级别**: real browser (chromium via Playwright)
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

- 用户未登录，无任何保存的 session 与 route。
- `frontend/dist/index.html` 已构建（`pnpm --filter @easyinterview/frontend build`）。
- `ui-design/index.html` 静态原型可加载（依赖外网 CDN：React 18 + Babel
  standalone + Google Fonts）。
- chromium 浏览器二进制已通过 `pnpm --filter @easyinterview/frontend
  test:pixel-parity:install`（即 `playwright install --with-deps chromium`）
  下载到 `~/Library/Caches/ms-playwright/`（macOS）或对应的本地缓存目录。
- D2 (`002-app-shell-visual-system`) 视觉系统全部交付，frontend dist 包含
  `ei-shell-topbar` / `ei-screen-shell` / `ei-auth-shell` 等语义 className 与
  fontsource 字体 bundle。

## 2 When

`trigger.sh` 跑 `pnpm --filter @easyinterview/frontend test:pixel-parity`，
等价于：

```bash
cd frontend && pnpm exec playwright test
```

Playwright 配置同时拉起两个 project（`desktop` 1440×900 + `mobile` 390×844，
均使用 chromium 引擎），依次执行：

- `tests/pixel-parity/topbar.spec.ts` — TopBar DOM 锚点 + computed style
  parity（四入口、显示控制、语言 dropdown、authenticated 头像菜单 dropdown /
  logout flow、ui-design 对照与 mode/aria contract）。
- `tests/pixel-parity/screens.spec.ts` — auth_login 卡片 shell DOM 锚点 +
  ui-design hash route `#route=auth_login` 对照 + retired-module 负向断言。
- `tests/pixel-parity/layout.spec.ts` — TopBar 与 auth shell 在两个 viewport
  下的 bounding box 不重叠 / 不溢出。
- `tests/pixel-parity/screenshot.spec.ts` — 默认 warm/light screenshot smoke
  （非空截图 buffer，不依赖 ignored local baseline）+ dark / customAccent 状态变更
  的可见 token diff（不依赖跨 ui-design 像素 diff，避免字体源差异引入 false
  positive）。
- `tests/pixel-parity/home.spec.ts` — Home hero / textarea / aux cards DOM 锚点、
  bounding box 与 dark/customAccent token 变化。
- `tests/pixel-parity/parse.spec.ts` — Home 到 parse 入口、textarea submit enable、
  upload modal DOM 锚点。
- `tests/pixel-parity/workspace.spec.ts` — workspace empty + full-state DOM anchor、
  bounding box、modal、theme 与 screenshot smoke；full-state 通过 server-bound
  initial route bootstrap 进入，不依赖 Home recent card 的 `resume-unbound`。
- `tests/pixel-parity/resume-workshop.spec.ts`、`resume-workshop-create.spec.ts`、
  `resume-workshop-branch-rewrites-edit.spec.ts` — Resume Workshop flat list、
  upload/paste create flow、rewrites/edit/detail parity 与 retired tree/branch/guided
  负向断言。
- `tests/pixel-parity/practice.spec.ts`、`generating.spec.ts`、`report.spec.ts`、
  `debrief.spec.ts` — 面试、生成、报告与复盘核心页面 DOM、布局、主题与
  screenshot smoke。

`webServer` 由 `frontend/scripts/serve-pixel-parity.mjs` 提供（Node 内置
模块；同时挂载 `frontend/dist` 与 `ui-design/`，并暴露 `/health` 探活）。

## 3 Then

- 全部 Playwright 用例（14 个 spec × 2 project）PASS、0 failed。
- TopBar 四入口 testid 在两个 project 下都存在；TopBar shell 高 58 / padding
  0 32 / border-bottom 1px solid `rgb(231, 226, 214)`。
- 默认 home 渲染 `topbar-nav-home[aria-current=page]`、`topbar-dark-toggle`
  `aria-pressed=false`、语言 dropdown 暴露 `topbar-lang-option-zh` /
  `topbar-lang-option-en`、retired-module（welcome / mistakes / growth / drill /
  独立 voice）testid 不可达。
- authenticated user menu 通过 mocked auth API 完成 login → avatar chip → dropdown
  → logout：dropdown 对齐 `ui-design/src/app.jsx` 的 220px 最小宽度、6px top gap、
  header / profile / settings / logout 项；mobile 下菜单保持在 viewport 内。
- auth_login 渲染 `ei-auth-shell` 双列（desktop）或单列（mobile，760px 媒体
  查询）；卡片 padding 28px；ei-text-display 头部字号 48px。
- dark toggle 把 `<html data-mode>` 翻到 `dark`、`--ei-color-bg-canvas`
  resolves 为 `#16130e`；`document.body.background-color` 切到 `rgb(22,
  19, 14)`。
- 激活 customAccent 后 `<html data-custom-accent="active"`、内联
  `--ei-color-accent` 为 `oklch(58% ...)`、base palette token 不被覆盖。
- screenshot smoke 生成非空 browser screenshot buffer；常规 gate 不要求
  `tests/pixel-parity/*-snapshots/` 下存在 ignored local baseline。

## 4 执行

```bash
# 0. 一次性预装 chromium 二进制（首次）
pnpm --filter @easyinterview/frontend test:pixel-parity:install

# 1. 跑场景
./test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/scripts/setup.sh
./test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/scripts/trigger.sh
./test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/scripts/verify.sh
./test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/scripts/cleanup.sh
```

`setup.sh` 检查 chromium 缓存 + `frontend/dist/index.html` 存在；缺失任一
都 exit ≠ 0 并给出可读提示。`trigger.sh` 跑 Playwright 后把日志写到
`.test-output/e2e/p0-006-ui-design-pixel-parity-gate/trigger.log`。
`verify.sh` 断言日志包含 passing summary 且没有 failed summary，并 grep
retired-module testid 不在 trigger 输出里的 failing trace 出现，同时确认当前
14 个 parity spec 已实际执行。

## 5 污染控制

- 场景使用真实 chromium 浏览器，但所有 page 实例由 Playwright fixture 隔离，
  不写共享数据库 / 集群 / 全局 localStorage。
- `webServer` 由 Playwright 自动启停，端口 4173 在 `reuseExistingServer:
  !CI` 模式下复用本地实例；CI 上每次重启。
- `setup.sh` 仅写一个 marker；`cleanup.sh` 移除 marker，保留 trigger.log
  与 Playwright report 作为证据。
- 截图 baseline 默认通过 `frontend/.gitignore` 排除入 git；常规 gate 使用
  screenshot smoke，不依赖 ignored baseline。

## 6 截图基线维护

baseline 文件位于 `frontend/tests/pixel-parity/*-snapshots/`
（`<test name>-<project>-<platform>.png`），不入 git。显式重生成：

```bash
cd frontend
pnpm exec playwright test tests/pixel-parity/screenshot.spec.ts --update-snapshots
```

`--update-snapshots` 会覆盖现有 baseline。默认 P0.006 gate 不把这些
`.gitignore` 文件作为 PASS 前提；只有在 CI / checkout 能稳定提供 baseline
artifact 时，才把 screenshot diff 升级为 hard gate。

## 7 不依赖外网时的局限

`ui-design/index.html` 通过 unpkg.com 引入 React + Babel + Google Fonts。
完全离线运行场景会让 ui-design 对照断言失败（topbar.spec.ts / screens
.spec.ts 内 ui-design hash route 部分）。如果需要离线运行：

1. vendor `https://unpkg.com/react@18.3.1/...`、`react-dom@18.3.1/...`、
   `@babel/standalone@7.29.0/...` 以及 Google Fonts CSS / woff 到
   `ui-design/vendor/`；
2. 修改 `ui-design/index.html` 的 `<script src=...>` / `<link href=...>`
   指向本地路径；
3. 重跑场景。

本场景目前预设 CI / 本地有外网，离线 vendor 是后续优化项。
