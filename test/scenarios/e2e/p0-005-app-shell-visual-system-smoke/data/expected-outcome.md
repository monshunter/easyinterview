# Expected Outcome

- TopBar 渲染 `app-shell-topbar` 元素，根级 className 命中 `ei-shell-topbar`，
  primary nav / display controls / user area 子节点 className 命中
  `ei-topbar-nav` / `ei-topbar-controls` / `ei-topbar-user`，三入口与显示控制
  按 ui-design 源同步。
- 切换主题 / 暗色后 `<html>` 的 `data-theme` / `data-mode` 即时翻转；
  `getComputedStyle(documentElement)` 在 warm/light 下读到
  `--ei-color-bg-canvas: #fdfcf8`，warm/dark 下读到
  `--ei-color-fg-primary: #f5f0e4`，ocean/light 下读到
  `--ei-color-bg-canvas: #0c0f17`（按 EI_THEMES 转写）。
- 激活 `customAccent` 后 `<html>` 的 inline style 仅包含
  `--ei-color-accent` 与 `--ei-color-accent-soft` 两个 oklch 值，base palette
  token 没有被覆盖；`topbar-custom-accent-hue` / `-chroma` slider 渲染。
- `route-auth_login` 渲染 `ei-auth-shell` 双列布局，`auth-login-email-form` /
  `auth-login-submit-email` 存在，范围外 password / OAuth stub 不渲染；
  `route-settings` 渲染 `ei-screen-shell` + `ei-screen-card`；
  范围外 standalone insight 输入不渲染 `route-standalone_insight`，而是 fallback 到
  home。
- out-of-scope 入口 / 范围外路由 / 范围外文案在 DOM 中均不可见。
- `ui-design/src/{app,screen-auth}.jsx` 字面量尺寸（height 58
  / padding 0 32 / gap 28 / max-width 1160 / padding 54 48 96 / grid
  0.88fr 1.12fr / card padding 28）能在 D2 CSS 内对应位置找到。

trigger 输出 `trigger.log` 必须出现：
- `src/app/scenarios/p0-005-app-shell-visual-system-smoke.test.tsx`
- `Tests  7 passed (7)`
- `Test Files  1 passed (1)`
