# Frontend Resume Workshop Listing Routing and Detail Readonly Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-11

**关联计划**: [plan](./plan.md)

## Phase 1: 路由替换 + 容器骨架

- [ ] 1.1 修订 `frontend/src/app/App.tsx`：`resume_versions` 路由从 `PlaceholderScreen` 切到 `ResumeWorkshopScreen`（验证：路由 hook 单测 + grep verify no `PlaceholderScreen.resume_versions`）
- [ ] 1.2 实现 `frontend/src/app/screens/resume-workshop/ResumeWorkshopScreen.tsx` 容器：解析 route param + flow 分发（验证：Vitest 单测 ≥ 5 param case PASS）
- [ ] 1.3 flow=create / branch P0 渲染 `<NotImplementedPlaceholder>`，不阻塞 list 主路径（验证：Vitest 单测）
- [ ] 1.4 实现 adapter 层 `frontend/src/app/screens/resume-workshop/adapters/resume.ts`，含 `mapResumeAssetToUiSource` / `mapResumeVersionToUi` / `mapBulletSuggestionToUi`（验证：adapter unit test ≥ 8 case PASS，覆盖 null / archived / parent chain）

## Phase 2: ResumeListView + TreeView + FlatView + StatsStrip

- [ ] 2.1 实现 `components/ResumeListView.tsx`：StatsStrip 4 项 + ViewSwitcher + 子视图调度（验证：Vitest 渲染 default fixture 命中 ≥ 8 testid）
- [ ] 2.2 实现 `components/ResumeTreeView.tsx`：originalId 分组 + 折叠 + 行内按钮 toast（验证：Vitest 单测 + 折叠交互 + click 按钮 toast 出现）
- [ ] 2.3 实现 `components/ResumeFlatView.tsx`：版本平铺 + match/updated_at 排序（验证：Vitest 单测 + 排序稳定性）
- [ ] 2.4 实现 `components/ResumeVersionRow.tsx`：indent / tag / match / date / click → detail（验证：Vitest 渲染 + nav 调用）
- [ ] 2.5 generated client `listResumes` / `listResumeVersions` 通过 mock transport 消费 [B2 fixtures](../../../mock-contract-suite/spec.md) `default` / `empty` / `paginated` scenario（验证：fixture parity test PASS）

## Phase 3: ResumeDetailView Preview Tab + 原件弹层

- [ ] 3.1 实现 `components/ResumeDetailView.tsx` 容器：Breadcrumb + 版本分支图 + 三 tab（验证：Vitest 渲染所有 testid + tab 切换）
- [ ] 3.2 默认 tab 选择：按 `resumeDefaultTab(version)` MASTER→preview / TARGETED→preview (P0 fallback，因 rewrites 未实现)（验证：Vitest 单测 2 case）
- [ ] 3.3 Preview Tab：渲染 `buildResumePlainText(lang, version)` adapter 投影（验证：Vitest 渲染 EN / ZH）
- [ ] 3.4 "查看原件" 按钮 → 原件弹层 modal（focus trap + ESC + 外层遮罩 + X）（验证：Vitest + Playwright a11y 键盘交互）
- [ ] 3.5 rewrites / edit Tab P0 渲染 `<ComingSoonTab>` 占位（容器 / testid / 切换逻辑保留）（验证：Vitest 渲染 + 切换不报错）
- [ ] 3.6 generated client `getResumeVersion` 消费 `default` / `targeted-with-suggestions` / `not-found-404` scenario（验证：fixture parity test）

## Phase 4: i18n + a11y + 隐私红线

- [ ] 4.1 复用 [frontend-shell i18n](../../../frontend-shell/spec.md) en/zh 配置，新增 `resumeWorkshop.*` key（验证：Vitest 切换 EN/ZH 关键文案）
- [ ] 4.2 a11y：focus 管理 / aria-label / 键盘导航完整（验证：Playwright a11y 键盘 + screen reader role 断言）
- [ ] 4.3 Accept-Language header 携带：lang 切换时 generated client 请求 header 携带 BCP47（验证：integration test 验证 header）
- [ ] 4.4 隐私红线 grep：raw resume text / parsed_summary / structured_profile 不出现在 console.log / URL / localStorage / telemetry / mock transport log（验证：Vitest + Playwright grep negative）

## Phase 5: UI parity gate + BDD + 旧入口负向 grep

- [ ] 5.1 复用 [frontend-shell/003-ui-design-pixel-parity-gate](../../../frontend-shell/plans/003-ui-design-pixel-parity-gate/plan.md) 框架，新增 baseline `frontend/tests/visual-baseline/resume-workshop/`（验证：baseline 截图存档）
- [ ] 5.2 Playwright pixel parity：desktop 1440px + mobile 390x844 viewport baseline + 容差 ≤ 0.1%（验证：`pnpm --filter @easyinterview/frontend build && pnpm --filter @easyinterview/frontend test:pixel-parity` PASS；首次或新机器先跑 `pnpm --filter @easyinterview/frontend test:pixel-parity:install`）
- [ ] 5.3 DOM parity：关键 testid 完整命中（StatsStrip 4 项 / TreeView 行 / FlatView 行 / DetailView Breadcrumb / Tab / Modal）（验证：Playwright DOM 断言）
- [ ] 5.4 computed style parity：accent / bg / 字号 / 字号 / 间距与 ui-design 源一致（验证：Playwright computed style 断言）
- [ ] 5.5 BDD-Gate: E2E.P0.036 resume-list-tree-flat-toggle PASS（详见 [bdd-checklist.md](./bdd-checklist.md)）
- [ ] 5.6 BDD-Gate: E2E.P0.037 resume-detail-preview-readonly PASS
- [ ] 5.7 旧入口 grep：`git grep -nE "mistake|growth|drill|onboarding-legacy|experiences-legacy" -- frontend/src/app/screens/resume-workshop/` 0 命中（验证：CI lint）
- [ ] 5.8 data.jsx import grep：`git grep -nE "import.*ui-design/src/data" -- frontend/src/app/screens/resume-workshop/` 0 命中（验证：CI lint）
- [ ] 5.9 在 `test/scenarios/e2e/INDEX.md` 追加 P0.036 + P0.037 行（关联需求 `frontend-resume-workshop C-1..C-9`，状态 Ready，automated）
- [ ] 5.10 同步 `docs/spec/engineering-roadmap/spec.md` §5.2 `frontend-resume-workshop` 状态从 "未创建" 改为 "active"（与 backend-upload / backend-resume 同步行）（验证：`sync-doc-index --check`）
