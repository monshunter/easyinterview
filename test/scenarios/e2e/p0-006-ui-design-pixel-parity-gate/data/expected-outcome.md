# Expected Outcome

- Playwright `webServer` 在端口 4173 起来后健康检查 `/health` 返回 200。
- desktop（1440×900）+ mobile（390×844）两个 chromium project 各跑：
  - `topbar.spec.ts`：3 入口 testid、ui-design 当前 nav 文本、
    TopBar height ~58 / padding 32 / border-bottom 1px、aria-current 与
    aria-pressed 默认值、语言 dropdown 选项、authenticated user menu
    dropdown + logout flow parity。
  - `screens.spec.ts`：navigates to auth_login + ei-auth-shell
    渲染、ei-text-display 头部、ei-auth-eyebrow 字体族、ui-design hash
    route h1 hero、auth_login 卡片 padding 28、out-of-scope entries 0 命中。
  - `layout.spec.ts`：TopBar fits viewport、三入口不重叠、display
    controls + user area 不重叠、auth_login 两栏在 desktop 双列 / mobile
    单列堆叠。
  - `screenshot.spec.ts`：home ocean/light screenshot smoke、dark
    toggle 翻转 token + body bg、customAccent 内联仅覆盖 accent / accent-
    soft、out-of-scope entries 0 命中。
  - `home.spec.ts`：Home hero / textarea / aux cards DOM 锚点、
    viewport 内布局与 dark mode token 变化。
  - `parse.spec.ts`：Parse 只读规划详情、页面级报告入口、无嵌入报告列表
    与 desktop/mobile source parity。
  - `workspace.spec.ts`：empty state、server-bound full-state、
    modal、bounding box、theme、screenshot smoke 与 out-of-scope entry negative。
  - `resume-workshop.spec.ts` / `resume-workshop-create.spec.ts`：flat list、
    read-only detail、upload/paste create flow、out-of-scope tree/branch/guided/
    rewrite/edit negative。
  - `practice.spec.ts` / `generating.spec.ts` / `report.spec.ts`：
    面试、生成与报告页面 parity。
  - `reports.spec.ts`：当前规划独立报告列表的六种状态、规划隔离与
    desktop/mobile parity。
- trigger.log 必须出现：
  - Playwright passed summary
  - 不含 failed summary
  - 不含 `topbar-nav-welcome` / `topbar-nav-mistakes` / `topbar-nav-growth` /
    `topbar-nav-drill` / `topbar-nav-voice` / `topbar-nav-jd_match` /
    `route-welcome` / `route-jd_match` 的 failing
    trace。
- Playwright report 写到 `.playwright-output/`（gitignored）。
