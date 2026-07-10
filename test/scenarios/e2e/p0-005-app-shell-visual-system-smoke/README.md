# E2E.P0.005 App Shell Visual System Smoke

> **场景 ID**: E2E.P0.005
> **执行方式**: automated
> **隔离级别**: in-process (vitest jsdom)
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

- 用户未登录，无任何保存的 session 与 route。
- 显示偏好默认 warm/light，`docs/ui-design/` 与 `ui-design/` 是唯一视觉验收真理源头。
- OpenAPI fixture transport 提供 `getRuntimeConfig` `default` 与 `getMe`
  `unauthenticated` 场景。
- `frontend/src/app/theme/` 设计 token、`themes.css`、`typography.css`、
  `topbar.css`、`auth.css`、`screens.css` 已在主 entry 中加载。

## 2 When

场景在 jsdom 中渲染 `<App client={fixtureClient} />`，注入 `themes.css` /
`typography.css` / `topbar.css` / `auth.css` / `screens.css` 到 document，并：

- 切换 `topbar-dark-toggle` 与 `topbar-theme-button` / `topbar-theme-option-*` 验证 `<html>` 的
  `data-theme` / `data-mode` 翻转后 `--ei-color-bg-canvas` /
  `--ei-color-fg-primary` 通过 `:root[data-theme=X][data-mode=Y]` selector
  resolves 到 ui-design 转写值。
- 点击 `topbar-theme-button` / `topbar-theme-custom-option` 触发 `customAccent` overlay，断言
  `<html>` 内联 `--ei-color-accent` / `--ei-color-accent-soft` 出现 oklch 值，
  且 base palette token（`--ei-color-bg-canvas` / `--ei-color-fg-primary`）
  没有被覆盖。
- 切换 `initialRoute` 渲染 `auth_login` / `settings`，并输入
  out-of-scope standalone insight route 验证其 fallback 到 home，断言
  `ei-auth-shell` / `ei-screen-shell` 双层结构与 `ei-screen-card` 卡片骨架。

## 3 Then

- TopBar 三入口、显示控制、custom accent picker、用户区按 `ei-shell-topbar`
  / `ei-topbar-nav` / `ei-topbar-controls` / `ei-topbar-user` 节奏渲染，
  className 与 ui-design 源同步。
- `:root[data-theme][data-mode]` selector 切换后，关键 CSS variable
  resolves 到 EI_THEMES 转写值；`customAccent` 仅覆盖 accent / accent-soft
  两条 inline style。
- Auth / settings / route shell 全部渲染 ui-design 原生卡片
  scaffold；out-of-scope 入口（`welcome` / `growth` / `mistakes` / `drill` / 独立
  `voice`）以及范围外 module 文案（错题本 / 成长中心 / 经历库 /
  目标角色 / 技能标签）均不可达。
- `ui-design/src/app.jsx` / `screen-auth.jsx` 的字面量
  尺寸（height 58 / padding 0 32 / gap 28 / max-width 1160 / padding 54 48
  96 / grid 0.88fr 1.12fr / card padding 28 / 等）与 D2 CSS 内对应值对齐。

## 4 执行

```bash
./test/scenarios/e2e/p0-005-app-shell-visual-system-smoke/scripts/setup.sh
./test/scenarios/e2e/p0-005-app-shell-visual-system-smoke/scripts/trigger.sh
./test/scenarios/e2e/p0-005-app-shell-visual-system-smoke/scripts/verify.sh
./test/scenarios/e2e/p0-005-app-shell-visual-system-smoke/scripts/cleanup.sh
```

## 5 污染控制

场景在 vitest + jsdom 中运行，不写共享数据库，不启动 Kind cluster；
`trigger.sh` 仅产生
`.test-output/e2e/p0-005-app-shell-visual-system-smoke/trigger.log` 作为验证
证据，`cleanup.sh` 删除 setup marker，保留日志。

## 6 当前覆盖与后续 Playwright follow-up

本场景在 vitest + jsdom 中验证以下维度，已经覆盖 `bdd-checklist.md` 中的
DOM / className / inline-style / 路由级负向 / 文案级负向 / ui-design 源
追溯：

- DOM 锚点：`app-shell-topbar` / `topbar-{primary-nav,display-controls,
  user-area}` / `topbar-nav-*` / `topbar-{theme-button,theme-menu,
  theme-option-*,theme-custom-option,dark-toggle,lang-toggle,
  custom-accent-{swatch,picker,hue,chroma,clear}}` /
  `route-{home,auth_login,settings}` 与 fallback 等全部存在，
  `route-standalone_insight` 不得 materialize。
- className 契约：`ei-shell-topbar` / `ei-topbar-{nav,nav-button,controls,
  user,custom-accent,theme,dark,lang}` / `ei-auth-{shell,side,card,form,cta,
  side-panel-pending}` / `ei-screen-{shell,card,card--route-shell}` /
  `ei-skeleton-stripe` / `ei-text-{display,title,body,caption,label}` 等。
- 状态翻转：`<html>` 的 `data-theme` / `data-mode` / `data-custom-accent`
  随 TopBar 控件变化即时翻转，`getComputedStyle(documentElement)`
  resolves CSS variables 到 EI_THEMES 转写值。
- 负向断言：out-of-scope 路由 / 范围外 testid / 范围外文案均无回流。
- ui-design 源追溯：`app.jsx` / `screen-auth.jsx`
  的字面量尺寸均能在 D2 CSS / TS 中找到。

以下维度依赖真实浏览器渲染，不属于本 scenario 验证范围，列为后续 follow-up
（建议派生 D2.1 plan 或独立场景包）：

- Desktop / mobile viewport 下的 CSS layout（flex / grid 计算）。
- `getBoundingClientRect()` overlap detection（jsdom 不做布局）。
- 与 `ui-design/index.html` golden preview 的截图差异 / pixel diff。
- Playwright / 等价浏览器渲染工具的 100% 像素级 parity gate。

跑通真实浏览器 parity gate 需要：

1. 引入 `@playwright/test` + chromium 浏览器二进制（首次约 100MB）。
2. 在 `frontend/` 增加 `playwright.config.ts` 与 `playwright.spec.ts`
   描述 desktop / mobile viewport、screenshot baseline、ui-design golden
   preview server。
3. 在场景脚本中先 `pnpm --filter @easyinterview/frontend build`，再启动
   静态 server 同时托管 `frontend/dist/index.html` 与 `ui-design/index.html`，
   最后调用 `playwright test`。
4. 把 baseline 截图与 DOM dump 写入
   `.test-output/e2e/p0-005-app-shell-visual-system-smoke/playwright-evidence/`，
   并在 `verify.sh` 中校验。

本场景目前在 vitest + jsdom 范围内就足以阻止 out-of-scope 模块回流、视觉
className 错位、CSS variable 失效、`customAccent` 范围溢出等 D2 BDD-Gate
关心的 regression。
