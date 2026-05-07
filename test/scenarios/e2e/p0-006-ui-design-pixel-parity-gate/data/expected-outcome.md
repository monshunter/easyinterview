# Expected Outcome

- Playwright `webServer` 在端口 4173 起来后健康检查 `/health` 返回 200。
- desktop（1440×900）+ mobile（390×844）两个 chromium project 各跑：
  - `topbar.spec.ts` 7 个用例：5 入口 testid、ui-design 5 个 nav 文本、
    TopBar height ~58 / padding 32 / border-bottom 1px、aria-current 与
    aria-pressed 默认值。
  - `screens.spec.ts` 6 个用例：navigates to auth_login + ei-auth-shell
    渲染、ei-text-display 头部、ei-auth-eyebrow 字体族、ui-design hash
    route h1 hero、auth_login 卡片 padding 28、retired entries 0 命中。
  - `layout.spec.ts` 4 个用例：TopBar fits viewport、五入口不重叠、display
    controls + user area 不重叠、auth_login 两栏在 desktop 双列 / mobile
    单列堆叠。
  - `screenshot.spec.ts` 4 个用例：home warm/light baseline 比对、dark
    toggle 翻转 token + body bg、customAccent 内联仅覆盖 accent / accent-
    soft、retired entries 0 命中。
- trigger.log 必须出现：
  - `42 passed`
  - `0 failed`
  - 不含 `topbar-nav-welcome` / `topbar-nav-mistakes` / `topbar-nav-growth` /
    `topbar-nav-drill` / `topbar-nav-voice` / `route-welcome` 任何一个 token。
- Playwright report 写到 `.playwright-output/`（gitignored）。
