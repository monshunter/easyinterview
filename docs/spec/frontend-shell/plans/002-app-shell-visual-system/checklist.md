# App Shell Visual System Checklist

> **版本**: 2.3
> **状态**: completed
> **更新日期**: 2026-07-10

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
