# App Shell Visual System Checklist

> **版本**: 3.2
> **状态**: completed
> **更新日期**: 2026-07-14

**关联计划**: [plan](./plan.md)

## Phase 1: Token / theme

- [x] 1.1 建立语义 token、4 个 theme/mode 色板组合和 source-to-target 追溯测试；验证: `tokens.test.ts` 断言 token 只暴露语义键、CSS variable 非空、色板值来自 `frontend/src`，且 active theme key 只包含 `ocean` / `plum`。
- [x] 1.2 接入 `customAccent` helper 与根级 display wiring；验证: `DisplayPreferencesRootWiring.test.tsx` 断言 `data-theme`、`data-mode`、`data-custom-accent` 和 accent variables 在切换后即时更新。
- [x] 1.3 保持 Vite CSS 管线；验证: `globalCss.test.ts` 断言 `main.tsx` 单次导入 `global.css`，且 `frontend/package.json` 不引入 Tailwind / CSS-in-JS 依赖。

## Phase 2: Fonts / typography

- [x] 2.1 使用可随仓库交付的开源字体；验证: `fonts.test.ts` 断言 fontsource 依赖、fallback 链和 private-font negative gate。
- [x] 2.2 暴露 `ei-text-*` typography scale；验证: `typography.test.tsx` 断言语义 className、字体族、字号、行高和 TopBar inline-px contract。

## Phase 3: TopBar

- [x] 3.1 TopBar shell 源级迁移；验证: `TopBarVisual.test.tsx` 断言 `ei-shell-topbar`、header 高度、padding、gap、sticky 层级、brand mark 和 nav structure。
- [x] 3.2 三入口与显示控制迁移；验证: `TopBar.test.tsx` 与 `TopBarVisual.test.tsx` 断言 `topbar-nav-home` / `workspace` / `resume_versions`、theme menu、custom accent row、dark toggle、language dropdown、login/user menu、i18n 和 aria behavior。

## Phase 4: Auth / Settings / screen shell

- [x] 4.1 Auth shell 迁移；验证: `AuthVisual.test.tsx` 断言 `auth_login` / `auth_verify` / `auth_profile_setup` / `auth_logout` 的 card shell、form、CTA、status 和 D1 testid regression。
- [x] 4.2 Settings 与通用 screen shell 迁移；验证: visual smoke 断言 `route-settings`、`ei-screen-shell`、`ei-screen-card`、skeleton 和 route shell anchor。

## Phase 5: BDD / handoff

- [x] 5.2 Handoff docs；验证: `frontend/README.md` 记录 token、display wiring、fonts、className 接入点、visual smoke 与 browser responsive browser verification gate 分工。

## Phase 6: theme option pruning

- [x] 6.1 UI design document and formal frontend expose only `ocean` / `plum` plus custom accent in the TopBar theme menu（验证：`frontend/src`, `frontend/src/app/topbar/TopBar.tsx`, `TopBar.test.tsx`）
- [x] 6.2 Active palette types and CSS remove `warm` / `forest` from the supported theme matrix（验证：`tokens.test.ts`, `DisplayPreferencesProvider.test.tsx`, `DisplayPreferencesRootWiring.test.tsx`）



## Phase 8: Noto Serif SC bundle deduplication

- [x] 8.1 删除 `fonts.css` 中重复的 Noto Serif SC full-font imports，保留含中文 unicode-range 的 400/500 分片（验证：focused font source red/green、visual-system/full frontend tests、build 产物无 `noto-serif-sc-chinese-simplified-*`、dist/asset byte delta、owner context/docs gates）

## Phase 9: Western font subset pruning

- [x] 9.1 将五个 Western/mono family 的 default imports 收敛为当前权重的 Latin imports，保留 EI_FONT_PRESETS 与中英 fallback 合同（验证：focused font red/green、visual-system/full frontend tests、build family asset count/byte delta、owner context/docs gates）



## Phase 11: Theme data ownership wording

- [x] 11.1 删除不存在的 themes.css generator 说明，将 handoff 文档对齐 3 个 font preset、固定 mono family、traceability test 与 TopBar metadata 真实消费者；tokens tests、owner context 和 docs gates 通过。<!-- verified: 2026-07-10 method=theme-data-ownership-reconciliation evidence="Runtime source and handoff README now name traceability tests and TopBar metadata, with 3 preset pairs plus fixed mono. Tokens pass 13/13 and typecheck passes." -->

## Phase 12: zero-consumer primitive export removal

- [x] 12.1 Add a focused UI contract RED assertion requiring `Sparkline` and `KV` declarations/exports to stay absent.
- [x] 12.2 Delete both zero-consumer primitive implementations and their `window` export entries without adding aliases or replacements.

## Phase 13: design canvas consumer-surface pruning

- [x] 13.1 Add a focused UI contract RED assertion that compares the design-canvas component/helper surface with the only tracked `canvas.html` consumer.
- [x] 13.2 Delete unpassed wrapper props, the one-use Practice alias and iframe params duplicated by the App context defaults; do not add aliases, placeholders or compatibility branches.

## Phase 14: unavailable design canvas sidecar removal

- [x] 14.1 Add a focused UI contract RED assertion that rejects the unprovisioned sidecar/host bridge while requiring current in-memory editing state.
- [x] 14.2 Delete the sidecar constant, delayed hydration/write effects and stale persistence comments; render children immediately and retain current session state behavior.

## Phase 15: unavailable prototype edit-mode bridge removal

- [x] 15.1 Add a focused UI contract RED assertion that rejects the host-only edit-mode bridge/panel while retaining current reachable display controls.
- [x] 15.2 Delete edit-mode state/messages, parent writebacks, panel/helpers and its exclusive role tweak channel; update the canvas note to current hash-driven behavior.

## Phase 16: zero-read canvas mode binding removal

- [x] 16.1 Add a focused UI contract RED assertion rejecting `isCanvasIframe` while retaining the no-chrome TopBar condition; classify positional callback placeholders separately in the cross-file reference inventory.
  <!-- verified: 2026-07-10 method=ui-contract-red+typescript-references evidence="Cross-file TypeScript language-service inventory scanned 10 JSX files / 884 bindings and found isCanvasIframe plus two positional callback placeholders. UI contract ran 45 tests with the prior 44 passing and only the new isCanvasIframe absence assertion failing." -->
- [x] 16.2 Delete the zero-read binding without adding an alias, wrapper or replacement state.
  <!-- verified: 2026-07-10 method=ui-contract-green+typescript-references evidence="Deleted only isCanvasIframe. UI contract passes 45/45 and retains hideTopBar data-nochrome matching. TypeScript language-service inventory now reports 10 files / 883 bindings / trueUnread=0, with only the Practice `_` and Resume `line` first callback parameters classified as positional placeholders." -->

## Phase 17: zero-consumer formal CSS pruning

- [x] 17.1 Add owner source RED gates for the zero-consumer screen grid, visually-hidden utility and TopBar custom-active modifier, including the stale README handoff entry.
- [x] 17.2 Delete the three CSS rules and stale README token without aliases, placeholders or removal markers; retain all current DOM selectors and prototype-backed `ei-scroll`.

## Phase 18: TopBar login rule consolidation

- [x] 18.1 Add a source RED gate requiring exactly one complete `ei-topbar-auth-login` rule while preserving the independent hover rule and DOM consumer.
- [x] 18.2 Merge the adjacent base declaration blocks without changing declarations, hover behavior or component DOM.
  <!-- verified: 2026-07-10 method=topbar-login-css-cascade-green evidence="TopBarVisual passes 15/15; exactly one base rule contains all 10 prior declarations, the independent hover rule remains, and TopBar keeps its existing login button consumer." -->
- [x] 18.3 Run focused TopBar, visual-system/full frontend, typecheck/build, owner/product contexts and docs/index/diff/pruning gates; then restore the owner to `completed`.

## Phase 19: minimal CustomAccentPicker

- [x] 19.1 Add source/DOM RED tests that pin the current preview/value/reset UI and `onClear` / `active` prop surface as removable targets while preserving hue/saturation consumers.
- [x] 19.2 Delete preview/value output, “恢复主题默认色 / Reset to theme accent” button, `onClear` / `active` props and matching caller/i18n/style residue; do not add aliases, hidden controls or compatibility props.
  <!-- verified: 2026-07-14 method=custom-accent-minimal-green evidence="Removed the prototype and formal preview/value/reset UI, active/onClear props and caller arguments, conditional track opacity, and three preview CSS selectors. UI source contract passes 63/63 and TopBarVisual passes 16/16 without aliases or hidden controls." -->
- [x] 19.3 Prove hue and saturation still update the root custom-accent overlay, and selecting Ocean or Plum is the only explicit exit from custom accent.

## BDD Gate

- [x] BDD-Gate: `BDD.SHELL.VISUAL.001` 由 [BDD checklist](./bdd-checklist.md) 关联 shell/display-preference owner behavior tests；视觉 parity 不包装为 E2E。
