# App Shell Visual System Checklist

> **版本**: 3.1
> **状态**: completed
> **更新日期**: 2026-07-14

**关联计划**: [plan](./plan.md)

## Phase 1: Token / theme

- [x] 1.1 建立语义 token、4 个 theme/mode 色板组合和 source-to-target 追溯测试；验证: `tokens.test.ts` 断言 token 只暴露语义键、CSS variable 非空、色板值来自 `ui-design/src/primitives.jsx`，且 active theme key 只包含 `ocean` / `plum`。
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

- [x] 5.1 BDD-Gate: `E2E.P0.005` App shell visual smoke；验证: 场景脚本覆盖 default shell、TopBar controls、auth/settings/screen shell、computed variables、`customAccent` overlay 和 route alias negative checks。
  <!-- verified: 2026-07-07 method=focused-frontend-and-scenario evidence="validate_context.py frontend-shell/002 frontend PASS; targeted owner wording grep returned no matches; pnpm --filter @easyinterview/frontend test src/app/theme/tokens.test.ts src/app/display/DisplayPreferencesRootWiring.test.tsx src/app/theme/globalCss.test.ts src/app/theme/fonts.test.ts src/app/theme/typography.test.tsx src/app/topbar/TopBarVisual.test.tsx src/app/topbar/TopBar.test.tsx src/app/auth/AuthVisual.test.tsx src/app/scenarios/p0-005-app-shell-visual-system-smoke.test.tsx PASS (9 files, 91 tests); pnpm --filter @easyinterview/frontend build PASS with existing chunk-size warning; E2E.P0.005 setup/trigger/verify/cleanup PASS (7 tests, existing C.UTF-8 locale warning only)" -->
- [x] 5.2 Handoff docs；验证: `frontend/README.md` 记录 token、display wiring、fonts、className 接入点、visual smoke 与 browser pixel parity gate 分工。

## Phase 6: theme option pruning

- [x] 6.1 UI truth source and formal frontend expose only `ocean` / `plum` plus custom accent in the TopBar theme menu（验证：`ui-design/src/primitives.jsx`, `ui-design/src/app.jsx`, `frontend/src/app/topbar/TopBar.tsx`, `TopBar.test.tsx`）
- [x] 6.2 Active palette types and CSS remove `warm` / `forest` from the supported theme matrix（验证：`tokens.test.ts`, `DisplayPreferencesProvider.test.tsx`, `DisplayPreferencesRootWiring.test.tsx`）
- [x] 6.3 BDD-Gate: `E2E.P0.005` rejects `topbar-theme-option-warm` and `topbar-theme-option-forest` while preserving custom accent（验证：`p0-005-app-shell-visual-system-smoke.test.tsx`）

## Phase 7: visual-smoke test lifecycle isolation

- [x] 7.1 P0.005 三个同步结构/负向用例在断言后显式 unmount，清除无关 runtime/Home state updates（验证：P0.005 7 tests 无 act warning、visual-system focused/full frontend tests、build、owner context/docs gates）
  <!-- verified: 2026-07-10 method=visual-smoke-test-lifecycle-isolation evidence="Focused red exposed 18 AppRuntimeProvider/HomeScreen act warnings. Added explicit unmount to the three synchronous structure/negative tests without changing assertions or production code. P0.005 7 tests and visual-system 9 files/92 tests pass warning-free; frontend build and owner/product contexts pass. Full frontend 137 files/829 tests pass and P0.005 is absent from the remaining warning list; completed-state docs/diff/pruning gates rerun during closeout." -->

## Phase 8: Noto Serif SC bundle deduplication

- [x] 8.1 删除 `fonts.css` 中重复的 Noto Serif SC full-font imports，保留含中文 unicode-range 的 400/500 分片（验证：focused font source red/green、visual-system/full frontend tests、build 产物无 `noto-serif-sc-chinese-simplified-*`、dist/asset byte delta、owner context/docs gates）
  <!-- verified: 2026-07-10 method=noto-serif-sc-bundle-deduplication evidence="Focused red rejected the duplicate chinese-simplified-400 import after proving the retained 400 bundle covers U+4e00. Removed only the two full-font imports and corrected the stale subset comment. Focused font 6, visual-system 93 and full frontend 830 tests pass warning-free; build and owner/product contexts pass. Built full-font assets fell 4->0, Noto Serif SC bytes 21,727,604->14,442,688 and dist 27,752->20,632 KB; completed-state docs/diff/pruning gates rerun during closeout." -->

## Phase 9: Western font subset pruning

- [x] 9.1 将五个 Western/mono family 的 default imports 收敛为当前权重的 Latin imports，保留 EI_FONT_PRESETS 与中英 fallback 合同（验证：focused font red/green、visual-system/full frontend tests、build family asset count/byte delta、owner context/docs gates）
  <!-- verified: 2026-07-10 method=western-font-subset-pruning evidence="Focused red rejected the generic Inter 400 import. Replaced 11 imports across five Western/mono families with current-weight Latin subsets; Noto Serif SC and already-Latin Geist remained unchanged. Font 7, visual-system 94 and full frontend 831 tests pass warning-free; build and owner/product contexts pass. Western assets fell 134->28 and 2,072,824->783,036 bytes, non-Latin Western assets fell to zero and dist 20,632->19,140 KB; completed-state docs/diff/pruning gates rerun during closeout." -->

## Phase 10: P0.005 scenario contract reconciliation

- [x] 10.1 BDD-Gate: 场景 README/seed/expected 与 executable smoke 统一为 ocean/light -> ocean/dark -> plum/dark，Warm / Forest 只作不存在负向断言；P0.005 直接引用当前 E2E.P0.006 browser gate，不维护浏览器安装或截图文件流程。
  <!-- verified: 2026-07-10 method=p0005-scenario-contract-reconciliation evidence="Red asset contract read stale warm tokens, old ocean value, future browser instructions and 7-test marker. Updated README/seed/expected/verify to current tokens and P0.006 boundary; focused P0.005 passes 8 tests, scenario setup/trigger/verify/cleanup passes, and the 9-file visual-system owner suite passes 95 tests." -->

## Phase 11: Theme data ownership wording

- [x] 11.1 删除不存在的 themes.css generator 说明，将 handoff 文档对齐 3 个 font preset、固定 mono family、traceability test 与 TopBar metadata 真实消费者；tokens tests、owner context 和 docs gates 通过。<!-- verified: 2026-07-10 method=theme-data-ownership-reconciliation evidence="Runtime source and handoff README now name traceability tests and TopBar metadata, with 3 preset pairs plus fixed mono. Tokens pass 13/13 and typecheck passes." -->

## Phase 12: zero-consumer primitive export removal

- [x] 12.1 Add a focused UI contract RED assertion requiring `Sparkline` and `KV` declarations/exports to stay absent.
  <!-- verified: 2026-07-10 method=prototype-primitive-export-red evidence="Focused UI contract ran the new primitive-global test only and failed on const Sparkline before reaching the exact six-export assertion; cross-file AST and repository text inventory found no consumer for Sparkline or KV." -->
- [x] 12.2 Delete both zero-consumer primitive implementations and their `window` export entries without adding aliases or replacements.
  <!-- verified: 2026-07-10 method=zero-consumer-prototype-primitive-removal evidence="Deleted both component blocks and removed both window export entries. Focused primitive gate passes, full UI contract passes 36/36, exact source inventory is zero, the export surface contains only Icon/Tag/Btn/Card/SectionHeader/ReadinessDial, and strict TypeScript unused checks remain clean." -->
- [x] 12.3 Run UI contract, AST/export inventory, P0.005, visual-system/full frontend tests, owner contexts and docs/diff/pruning gates.
  <!-- verified: 2026-07-10 method=prototype-primitive-export-regression evidence="UI contract passes 36/36; symbol inventory is zero and the exact six-global export is present. Visual-system owner passes 9 files/95 tests, P0.005 passes 8/8 with setup/trigger/verify/cleanup, build and full frontend 137 files/841 tests pass. Static-browser Home loads with all six kept globals, both removed globals absent, no page errors or horizontal overflow. Both owner contexts and docs/diff/pruning gates pass; no environment restart or data cleanup occurred." -->

## Phase 13: design canvas consumer-surface pruning

- [x] 13.1 Add a focused UI contract RED assertion that compares the design-canvas component/helper surface with the only tracked `canvas.html` consumer.
  <!-- verified: 2026-07-10 method=design-canvas-consumer-surface-red evidence="Focused node:test ran only the new design canvas contract and failed on the existing DesignCanvas children/minScale/maxScale/style signature. Babel AST inventory confirms canvas.html is the only tracked consumer and never supplies those props or the other targeted extension params." -->
- [x] 13.2 Delete unpassed wrapper props, the one-use Practice alias and iframe params duplicated by the App context defaults; do not add aliases, placeholders or compatibility branches.
  <!-- verified: 2026-07-10 method=design-canvas-consumer-surface-green evidence="Focused contract passes after removing unpassed DesignCanvas/viewport/section/artboard/post-it extension props, inlining the one-use PracticeScreen alias, and delegating duplicate interview defaults to createInterviewContext. Existing scale bounds, gap, card style, post-it positions, fixed zh locale and session-24 behavior remain as local constants." -->
- [x] 13.3 Run UI contract, AST consumer inventory, static-browser canvas smoke, P0.005, visual-system/full frontend tests, owner contexts and docs/diff/pruning gates.
  <!-- verified: 2026-07-10 method=design-canvas-consumer-surface-regression evidence="UI contract passes 37/37 and AST inventory reports only consumer-backed signatures with 10 sections and 26 artboards. Static browser loads all 26 iframes with the expected minimal Home/Workspace/Practice/Report URLs, no horizontal overflow, and working focus, arrow navigation, Esc and 0.1..8 zoom behavior. Visual owner passes 9 files/95 tests, P0.005 passes 8/8 with setup/trigger/verify/cleanup, build and full frontend 137 files/841 tests pass, and both owner contexts validate. No environment restart or data cleanup occurred." -->

## Phase 14: unavailable design canvas sidecar removal

- [x] 14.1 Add a focused UI contract RED assertion that rejects the unprovisioned sidecar/host bridge while requiring current in-memory editing state.
  <!-- verified: 2026-07-10 method=design-canvas-sidecar-red evidence="Focused node:test failed on the existing sidecar constant/bridge. Repository inventory found no sidecar file, generator, host implementation or docs contract, and the tracked run.sh browser path logged GET /.design-canvas.state.json 404 on every load." -->
- [x] 14.2 Delete the sidecar constant, delayed hydration/write effects and stale persistence comments; render children immediately and retain current session state behavior.
  <!-- verified: 2026-07-10 method=design-canvas-sidecar-green evidence="Focused contract passes after deleting the sidecar constant, read/write effects, 150ms ready gate and host-specific comments. DesignCanvas still owns state.sections/focus, exposes patchSection/setFocus, and renders children immediately without a replacement persistence layer." -->
- [x] 14.3 Verify zero sidecar/host references, zero static-server 404s, browser reorder/rename/focus behavior, UI/P0.005/full frontend, owner contexts and docs/diff/pruning gates.
  <!-- verified: 2026-07-10 method=design-canvas-sidecar-regression evidence="Production sidecar/host symbols are zero and UI contract passes 38/38. The tracked run.sh browser path loads 10 sections and all 26 iframes with empty browser errors and only 200/304 server responses; renaming home, dragging it behind workspace and focusing it shows the changed label/order in-session, while refresh restores source order. Visual owner passes 9 files/95 tests, P0.005 passes 8/8 with setup/trigger/verify/cleanup, build and full frontend 137 files/841 tests pass, and both contexts validate. No environment restart or data cleanup occurred." -->

## Phase 15: unavailable prototype edit-mode bridge removal

- [x] 15.1 Add a focused UI contract RED assertion that rejects the host-only edit-mode bridge/panel while retaining current reachable display controls.
  <!-- verified: 2026-07-10 method=prototype-edit-mode-red evidence="Focused node:test failed on the existing host message protocol. Repository inventory found no host/listener/docs entry; TweaksPanel is only rendered from that protocol, tweaksAvailable is write-only, and role is exclusive to the panel plus ignored Home/Practice call props." -->
- [x] 15.2 Delete edit-mode state/messages, parent writebacks, panel/helpers and its exclusive role tweak channel; update the canvas note to current hash-driven behavior.
  <!-- verified: 2026-07-10 method=prototype-edit-mode-green evidence="Focused contract passes after deleting both edit-mode states, message effect, parent writebacks, TweaksPanel/TweakRow/selectStyle, EDITMODE markers and role defaults/hash/call props. TopBar display controls, AccentPicker, Settings font preset state and direct canvas hash variants remain." -->
- [x] 15.3 Verify zero edit-mode/role residuals, browser theme/dark/language/accent/font behavior, UI/P0.005/full frontend, owner contexts and docs/diff/pruning gates.
  <!-- verified: 2026-07-10 method=prototype-edit-mode-regression evidence="UI contract passes 39/39; edit-mode/message/panel/role production residuals are zero and AST finds no zero-read React state. Static browser verifies Ocean-to-Plum, light-to-dark, zh-to-en, Custom sliders and Settings Source Serif Pro/Geist selection with empty browser errors and only 200/304 server responses. Visual owner passes 9 files/95 tests, P0.005 passes 8/8 with setup/trigger/verify/cleanup, build and serial full frontend 137 files/841 tests pass, and both contexts validate. One unrelated P0.037 timing race appeared only during a parallel build/test run; its focused 6/6 and serial full rerun pass, and it is queued for the next independent test-debt batch. No environment restart or data cleanup occurred." -->

## Phase 16: zero-read canvas mode binding removal

- [x] 16.1 Add a focused UI contract RED assertion rejecting `isCanvasIframe` while retaining the no-chrome TopBar condition; classify positional callback placeholders separately in the cross-file reference inventory.
  <!-- verified: 2026-07-10 method=ui-contract-red+typescript-references evidence="Cross-file TypeScript language-service inventory scanned 10 JSX files / 884 bindings and found isCanvasIframe plus two positional callback placeholders. UI contract ran 45 tests with the prior 44 passing and only the new isCanvasIframe absence assertion failing." -->
- [x] 16.2 Delete the zero-read binding without adding an alias, wrapper or replacement state.
  <!-- verified: 2026-07-10 method=ui-contract-green+typescript-references evidence="Deleted only isCanvasIframe. UI contract passes 45/45 and retains hideTopBar data-nochrome matching. TypeScript language-service inventory now reports 10 files / 883 bindings / trueUnread=0, with only the Practice `_` and Resume `line` first callback parameters classified as positional placeholders." -->
- [x] 16.3 Run UI contract, AST inventory, P0.005, visual-system/full frontend, typecheck/build, owner contexts and docs/diff/pruning gates.
  <!-- verified: 2026-07-10 method=prototype-canvas-mode-binding-regression evidence="UI contract passes 45/45; TypeScript language-service inventory reports 10 files / 883 bindings / trueUnread=0 and two positional placeholders. Visual-system owner passes 9 files/95 tests; P0.005 passes 8/8 with setup/trigger/verify/cleanup; full frontend passes 137 files/841 tests, typecheck and build. Owner/product contexts, git diff check and pruning surface pass with real_residuals=0. No Bug or retrospective report was needed because the removed binding had no read path. No environment restart or data cleanup occurred." -->

## Phase 17: zero-consumer formal CSS pruning

- [x] 17.1 Add owner source RED gates for the zero-consumer screen grid, visually-hidden utility and TopBar custom-active modifier, including the stale README handoff entry.
  <!-- verified: 2026-07-10 method=frontend-shell-css-source-red evidence="Focused globalCss/ScreensVisual/TopBarVisual ran 30 tests: all 27 existing contracts passed and exactly the 3 new zero-consumer gates failed on their respective current CSS rules." -->
- [x] 17.2 Delete the three CSS rules and stale README token without aliases, placeholders or removal markers; retain all current DOM selectors and prototype-backed `ei-scroll`.
  <!-- verified: 2026-07-10 method=frontend-shell-css-source-green evidence="Focused owner gates pass 30/30. Target selectors are absent from runtime CSS/README and remain only as negative literals; base TopBar swatch and screen-card selectors retain TSX consumers, while ei-scroll retains prototype Practice consumers and its parity gate." -->
- [x] 17.3 Run focused owner tests, P0.005, full visual-system/frontend regression, typecheck/build, owner/product contexts and docs/index/diff/pruning gates; then restore the owner to `completed`.
  <!-- verified: 2026-07-10 method=frontend-shell-zero-consumer-css-pruning evidence="Focused gates pass 30/30; visual-system owner passes 10 files/107 tests; P0.005 setup/trigger/verify/cleanup passes 8/8; full frontend passes 136 files/840 tests; typecheck/build and both contexts pass. Runtime CSS/README target inventory is zero, prototype-backed ei-scroll remains, and final docs/index/diff/pruning gates run during closeout. No Bug/retrospective report, environment restart or data cleanup was needed." -->

## Phase 18: TopBar login rule consolidation

- [x] 18.1 Add a source RED gate requiring exactly one complete `ei-topbar-auth-login` rule while preserving the independent hover rule and DOM consumer.
  <!-- verified: 2026-07-10 method=topbar-login-css-cascade-red evidence="Focused TopBarVisual ran 15 tests: all 14 existing contracts passed and only the new unique complete login rule gate failed because two base blocks remained." -->
- [x] 18.2 Merge the adjacent base declaration blocks without changing declarations, hover behavior or component DOM.
  <!-- verified: 2026-07-10 method=topbar-login-css-cascade-green evidence="TopBarVisual passes 15/15; exactly one base rule contains all 10 prior declarations, the independent hover rule remains, and TopBar keeps its existing login button consumer." -->
- [x] 18.3 Run focused TopBar, visual-system/full frontend, typecheck/build, owner/product contexts and docs/index/diff/pruning gates; then restore the owner to `completed`.
  <!-- verified: 2026-07-10 method=topbar-login-css-cascade-consolidation evidence="Focused TopBarVisual passes 15/15; visual-system owner passes 10 files/109 tests including the 8 P0.005 smoke assertions; full frontend passes 136 files/845 tests; typecheck/build and both contexts pass. Inventory reports one base rule, one hover rule and one DOM consumer; final docs/index/diff/pruning gates pass during closeout. No Bug/retrospective report, environment restart or data cleanup was needed." -->

## Phase 19: minimal CustomAccentPicker

- [x] 19.1 Add source/DOM RED tests that pin the current preview/value/reset UI and `onClear` / `active` prop surface as removable targets while preserving hue/saturation consumers.
  <!-- verified: 2026-07-14 method=custom-accent-minimal-red evidence="UI source contract ran 63 tests with only the new minimal AccentPicker contract failing on active/onClear/preview/reset. Focused TopBarVisual ran 16 tests with only the new DOM/source residue assertion failing on the existing preview block." -->
- [x] 19.2 Delete preview/value output, “恢复主题默认色 / Reset to theme accent” button, `onClear` / `active` props and matching caller/i18n/style residue; do not add aliases, hidden controls or compatibility props.
  <!-- verified: 2026-07-14 method=custom-accent-minimal-green evidence="Removed the prototype and formal preview/value/reset UI, active/onClear props and caller arguments, conditional track opacity, and three preview CSS selectors. UI source contract passes 63/63 and TopBarVisual passes 16/16 without aliases or hidden controls." -->
- [x] 19.3 Prove hue and saturation still update the root custom-accent overlay, and selecting Ocean or Plum is the only explicit exit from custom accent.
  <!-- verified: 2026-07-14 method=custom-accent-interaction-proof evidence="P0.005 changes hue to 120 and chroma to 0.205, observes the exact root oklch overlay, then proves both Ocean and Plum remove data-custom-accent and both inline accent tokens. TopBarVisual confirms the single setCustomAccent(null) source branch belongs to predefined theme options; scenario passes 8/8 and TopBarVisual passes 16/16." -->
- [x] 19.4 BDD-Gate: `E2E.P0.005` covers interaction and old-anchor zero reference; `E2E.P0.006` covers desktop/mobile DOM/style/bbox/viewport/screenshot parity without empty picker space or overflow.
- [x] 19.5 Run UI contract, focused TopBar/display/P0.005, P0.006 browser parity, full frontend typecheck/build, owner contexts, docs/diff and old reset/prop negative searches before restoring `completed`.
  <!-- verified: 2026-07-14 evidence="P0.005 interaction/negative gate and P0.006 170/170 desktop/mobile parity PASS; UI contract 65/65, full frontend 125 files / 1004 tests, typecheck/build and reset-preview-value negative searches PASS." -->
