# Frontend Resume Workshop Listing Routing and Detail Readonly

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-11

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [frontend-resume-workshop spec](../../spec.md) §6 C-1..C-9 落到 `frontend/` 实现：

- 替换 [frontend-shell](../../../frontend-shell/spec.md) `PlaceholderScreen` 中 `resume_versions` 的 D2-D6 占位，由本 plan 落地的 `ResumeWorkshopScreen` 接管；
- 源级复刻 [`ui-design/src/screen-resume-workshop.jsx`](../../../../../ui-design/src/screen-resume-workshop.jsx) 中以下组件：
  - `ResumeWorkshopScreen`（路由容器，flow 参数解析）
  - `ResumeListView`（统计条 StatsStrip + ViewSwitcher + 子视图调度）
  - `ResumeTreeView`（原始简历树 + 折叠 + "选为底稿" / "基于这棵树新建版本" 按钮，**P0 范围只渲染**，按钮点击 P0 toast "即将开放"，留 plan 002/003 启用）
  - `ResumeFlatView`（版本平铺 + 排序）
  - `ResumeDetailView` 容器 + **Preview Tab 只读**（Breadcrumb + 版本分支图 + 三 tab 切换 + 默认 tab 按 `resumeDefaultTab(version)` + "查看原件" 弹层）
  - `ResumeVersionRow`（版本行复用）
- 实现 adapter 层 `frontend/src/app/screens/resume-workshop/adapters/`：把 generated client `ResumeAsset` / `ResumeVersion` 映射为 UI 真理源 `ResumeSource` / `ResumeVersion`；
- 通过 [B2 fixtures](../../../mock-contract-suite/spec.md) `listResumes` / `listResumeVersions` / `getResumeVersion` `default` / `empty` / `paginated` / `not-found-404` 等 scenario 完成 happy path + 边界 + 错误态；
- 落地 UI parity gate（Vitest + Playwright + pixel parity）；
- 完成 spec §6 C-1..C-9 验收 + E2E.P0.036 / E2E.P0.037 两个 BDD 场景；
- 不实现 Create Flow / Branch Flow / Rewrites Tab / Edit Tab（归 plan 002 / 003）；不依赖 [backend-resume](../../../backend-resume/spec.md) 真实落地（mock-first）。

## 2 背景

本 plan 是 frontend-resume-workshop 第一批 plan，承担 P0 用户路径 "进入 Resume 入口 → 看到简历列表 → 打开版本详情查看预览" 的前端端到端。它是 Resume Workshop 阶段 1 三个新 subspec 中第一个**纯 mock-first 可独立推进**的 plan：

- 不依赖 [backend-resume/001](../../../backend-resume/plans/001-asset-register-parse-and-listing/plan.md) 落地，因为 listResumes / listResumeVersions / getResumeVersion fixtures 由 [openapi-v1-contract/004](../../../openapi-v1-contract/plans/004-resume-additive-coverage/plan.md) 落地即可消费。
- 不依赖 [backend-upload/001](../../../backend-upload/plans/001-file-objects-and-presign-baseline/plan.md) 落地，因为 P0 不含 Create Flow（upload tab）。
- 与 [frontend-workspace-and-practice/001 ResumePickerModal disabled-list 模式](../../../frontend-workspace-and-practice/plans/001-workspace-and-interview-context/plan.md) 解锁路径并行：本 plan 直接消费 fixture-backed listResumes，不需要等 backend-resume 切真。

每个 phase 是可独立验证的纵向切片：Phase 1 起来就有路由接管 + 容器；Phase 2 起来就有 ResumeListView TreeView + FlatView + StatsStrip；Phase 3 起来就有 ResumeDetailView Preview Tab；Phase 4 起来就有 i18n + a11y + 隐私红线；Phase 5 起来就有 UI parity gate + BDD + 旧入口负向 grep。

执行本 plan 前必须确认：

- [openapi-v1-contract/004](../../../openapi-v1-contract/plans/004-resume-additive-coverage/plan.md) Phase 1-5 已完成（`listResumes` / `listResumeVersions` / `getResumeVersion` schema、fixtures、inventory lint 与 generated client artifacts 均已就位）。
- [frontend-shell](../../../frontend-shell/spec.md) `PlaceholderScreen` D2-D6 ownership 处于 active；本 plan 修订 App.tsx 路由表替换 resume_versions 映射。
- [mock-contract-suite/001](../../../mock-contract-suite/plans/001-fixture-backed-mock-runtime/plan.md) Vite dev preview 默认 fixture-backed 已就位。
- UI 真理源 [`ui-design/src/screen-resume-workshop.jsx`](../../../../../ui-design/src/screen-resume-workshop.jsx) + [`docs/ui-design/resume-module.md`](../../../../ui-design/resume-module.md) v1.7 / [`docs/ui-design/jd-resume-management.md`](../../../../ui-design/jd-resume-management.md) v1.5 已 active。

## 3 质量门禁分类

- **Plan 类型**: `code-internal + feature-behavior`。本 plan 实现前端组件 / adapter / 路由替换；用户可见 UI 行为。
- **TDD 策略**: 适用。Red-Green-Refactor 入口：
  1. Vitest 组件单测：ResumeWorkshopScreen / ResumeListView / ResumeTreeView / ResumeFlatView / ResumeDetailView Preview Tab / ResumeVersionRow render / route param 解析 / lang 切换 / a11y attribute；
  2. adapter unit test：`ResumeAsset ↔ ResumeSource` / `ResumeVersion ↔ ResumeVersion` 映射 + 边界（null / archived / parent_version_id 链）；
  3. fixture parity test：组件渲染从 `default` / `empty` / `paginated` / `not-found-404` fixture 时 DOM testid 覆盖 spec §6 锚点；
  4. Playwright pixel parity：desktop + mobile viewport baseline；
  5. negative grep test：`frontend/src/app/screens/resume-workshop/` 不出现 retired 模块名 / 不 import data.jsx。
  执行入口：`/implement frontend-resume-workshop/001-listing-routing-and-detail-readonly` → `/tdd`。
- **BDD 策略**: 适用（Feature plan requires BDD）。E2E.P0.036 list-tree-flat-toggle + E2E.P0.037 detail-preview-readonly。详见 [bdd-plan.md](./bdd-plan.md) / [bdd-checklist.md](./bdd-checklist.md)。
- **替代验证 gate**:
  - `pnpm --filter @easyinterview/frontend test` (Vitest)
  - `pnpm --filter @easyinterview/frontend build` + `pnpm --filter @easyinterview/frontend test:pixel-parity` (Playwright；首次或新机器先跑 `pnpm --filter @easyinterview/frontend test:pixel-parity:install`)
  - `pnpm --filter @easyinterview/frontend lint` (ESLint + UI parity rules)
  - `pnpm --filter @easyinterview/frontend build`
  - `git grep -E "mistake|growth|drill|onboarding-legacy|experiences-legacy" -- frontend/src/app/screens/resume-workshop/`（旧入口 negative）
  - `sync-doc-index --check`

### 3.1 Frontend / Backend Operation Matrix

本 plan 走 mock-first frontend path：fixtures 和 generated client 可用即能实现 UI，但 real backend handler 状态必须显式标注，避免把 fixture-backed UI 误判为真实后端闭环。

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `listResumes` | `openapi/fixtures/Resumes/listResumes.json` `default` / `empty` / `paginated` | `ResumeListView` + adapter `mapResumeAssetToUiSource` | `backend-resume/001` not-yet-implemented at this plan start; fixture-backed until landed | `resume_assets` | none | E2E.P0.036 |
| `listResumeVersions` | `openapi/fixtures/Resumes/listResumeVersions.json` `default` / `master-only` / `with-targeted-branches` | `ResumeTreeView` / `ResumeFlatView` + adapter `mapResumeVersionToUi` | future `backend-resume/002-versions-and-tailor-runs`; fixture-backed in this plan | `resume_versions` | none | E2E.P0.036 |
| `getResumeVersion` | `openapi/fixtures/Resumes/getResumeVersion.json` `master-default` / `targeted-with-suggestions` / `not-found-404` | `ResumeDetailView` Preview Tab / original modal | future `backend-resume/002-versions-and-tailor-runs`; fixture-backed in this plan | `resume_versions` / `resume_version_suggestions` | none for readonly preview; provenance is fixture-backed | E2E.P0.037 |

## 4 实施步骤

### Phase 1: 路由替换 + 容器骨架

#### 1.1 修订 `frontend/src/app/App.tsx`
- 路由表中 `resume_versions` 从 `PlaceholderScreen` 切到 `ResumeWorkshopScreen`
- 移除 D2-D6 ownership 注释中的 `resume_versions` 占位标记

#### 1.2 实现 `frontend/src/app/screens/resume-workshop/ResumeWorkshopScreen.tsx`
- 解析 route param：`flow=create|branch|list（默认）` + `versionId` + `tab` + `branchOriginalId`
- flow=create / branch 时 P0 渲染 `<NotImplementedPlaceholder>`（不阻塞，由 002/003 接管）
- flow=list 时渲染 `ResumeListView`
- versionId 存在时渲染 `ResumeDetailView` 容器

#### 1.3 adapter 层骨架
- `frontend/src/app/screens/resume-workshop/adapters/resume.ts`：`mapResumeAssetToUiSource(asset)` / `mapResumeVersionToUi(version)` / `mapBulletSuggestionToUi(suggestion)`
- adapter 单元测试覆盖 null / archived / parent_version_id 链

### Phase 2: ResumeListView + TreeView + FlatView + StatsStrip

#### 2.1 `frontend/src/app/screens/resume-workshop/components/ResumeListView.tsx`
- 顶部 StatsStrip（4 项统计：originals / versions / top-match / recent）
- ViewSwitcher（tree / flat 两态）
- 子视图调度
- 数据消费：generated client `listResumes` + `listResumeVersions`；通过 adapter 层投影到 UI 类型

#### 2.2 `ResumeTreeView.tsx`
- 按 originalId 分组的折叠树
- 行内 icon=file（原始）/ chevron_right/down（折叠态）
- "选为底稿" / "基于这棵树新建版本" 按钮 P0 toast "即将开放"（按钮 DOM + click handler 不缺失，但 P0 不触发实际逻辑）

#### 2.3 `ResumeFlatView.tsx`
- 版本平铺，按 `match DESC nullsLast / updated_at DESC` 排序
- 行内列：版本名（MASTER / TARGETED tag）/ 来源原始 / 目标岗位 / 匹配分 / 最近编辑

#### 2.4 `ResumeVersionRow.tsx`
- 复用组件，indent / tag / match / date 字段
- 点击进入详情 `nav("resume_versions", { versionId })`

### Phase 3: ResumeDetailView Preview Tab + 原件弹层

#### 3.1 `ResumeDetailView.tsx`
- 顶部 Breadcrumb（resume_versions / 当前 master / 当前 version）
- 中部版本分支图（仅渲染当前 version 的 parent 链 + 同级 targeted versions）
- 三 tab 切换（preview / rewrites / edit；P0 只 preview 可点；rewrites / edit P0 渲染 `<ComingSoonTab>`）
- 默认 tab：按 `resumeDefaultTab(version)` (MASTER → preview / TARGETED → rewrites P0 fallback preview)
- Preview Tab：渲染 `buildResumePlainText(lang, version)` 投影
- "查看原件" 按钮 → 打开原件弹层 modal（focus trap + ESC 关闭 + 外层遮罩关闭 + X 按钮）

#### 3.2 数据消费
- 通过 generated client `getResumeVersion(versionId)` + adapter 投影
- 错误态：404 → 渲染 `<NotFoundEmptyState>` + 返回 list CTA

### Phase 4: i18n + a11y + 隐私红线

#### 4.1 i18n
- 复用 [frontend-shell i18n](../../../frontend-shell/spec.md) `en.ts` / `zh.ts` 配置；新增 key 前缀 `resumeWorkshop.*`
- `buildResumeData(lang)` / `resumeDefaultTab(version)` / `buildResumePlainText(lang, version)` / `buildBullets(lang, version)` 已在 UI 真理源 `ui-design/src/screen-resume-workshop.jsx` 中定义；本 plan 只转写必要投影，不运行时 import 原型文件
- 测试覆盖 EN / ZH 切换的关键文案 + Accept-Language header 携带

#### 4.2 a11y
- 焦点管理：ResumeVersionRow → ResumeDetailView 进入时 focus 移到 Breadcrumb；原件弹层 focus trap
- aria-label 完整：StatsStrip 统计 / TreeView 折叠按钮 / ViewSwitcher / Tab 切换
- 键盘导航：Tab / Enter / ESC

#### 4.3 隐私红线
- raw resume text / parsed_summary / structured_profile 内容不出现在 console.log / URL query / localStorage / telemetry
- Preview Tab 内容仅在 user 主动复制时通过 clipboard 流出（非持久）
- Vitest 单测验证 `getResumeVersion` fixture 中 PII 字段在 DOM 但不在 URL / localStorage / mock-server transport log

### Phase 5: UI parity gate + BDD + 旧入口负向 grep

#### 5.1 UI parity gate
- 复用 [frontend-shell/003-ui-design-pixel-parity-gate](../../../frontend-shell/plans/003-ui-design-pixel-parity-gate/plan.md) 框架
- 新增 baseline：desktop 1440px + mobile 390x844 viewport
- 关键元素 bounding box parity：StatsStrip 4 项 / ResumeTreeView 行高 / ResumeVersionRow indent / ResumeDetailView Breadcrumb
- screenshot 容差与 frontend-shell 003 一致

#### 5.2 BDD 场景验证
- 执行 `test/scenarios/e2e/p0-036-resume-list-tree-flat-toggle/` 全 PASS
- 执行 `test/scenarios/e2e/p0-037-resume-detail-preview-readonly/` 全 PASS
- 在 `test/scenarios/e2e/INDEX.md` 追加 P0.036 + P0.037 行

#### 5.3 旧入口负向 grep
- `git grep -nE "mistake|growth|drill|onboarding-legacy|experiences-legacy" -- frontend/src/app/screens/resume-workshop/`：0 命中
- `git grep -nE "import.*ui-design/src/data" -- frontend/src/app/screens/resume-workshop/`：0 命中（不允许运行时 import data.jsx）

#### 5.4 spec / history / INDEX 同步

- frontend-resume-workshop spec.md 1.0 保持 active（首版）
- frontend-resume-workshop history.md plan 001 完成后追加新行
- 同步 `docs/spec/engineering-roadmap/spec.md` §5.2 `frontend-resume-workshop` 状态从 "未创建" 改为 "active"（与 backend-upload / backend-resume 同步行）

## 5 验收标准

- 本计划列出的 §4 所有 Phase task 全部完成
- §3 替代验证 gate 全部通过
- spec §6 C-1..C-9 全部 PASS（C-10 / C-11 留给 plan 002 / 003）
- BDD E2E.P0.036 + E2E.P0.037 PASS
- UI parity baseline 已保存到 `frontend/tests/visual-baseline/resume-workshop/`
- engineering-roadmap §5.2 `frontend-resume-workshop` 状态已升级到 active

## 6 风险与应对

| 风险 | 应对 |
|------|------|
| R1: 术语 adapter 层与 generated client 类型耦合导致 generated client 更新破坏 adapter | adapter 层只依赖 generated client 字段集合（type-narrow + 不 import generated class）；generated client 升级时跑 fixture parity test 验证 |
| R2: UI parity gate 假阳性（截图细微差异） | screenshot 容差与 frontend-shell 003 baseline 一致；如出现假阳性，先验证 ui-design 源是否变化，再更新 baseline |
| R3: rewrites / edit Tab P0 仅 ComingSoonTab 但 testid 不完整 | Tab 容器渲染完整（DOM 锚点 / aria / 切换逻辑），仅内容是 `<ComingSoonTab>` 占位；plan 003 替换内容时 Tab 容器不变 |
| R4: 原件弹层 a11y 复杂度（focus trap + ESC + 外层遮罩） | 复用 [frontend-shell](../../../frontend-shell/spec.md) Modal primitive；测试覆盖键盘 + 鼠标多路径 |
| R5: 用户在 `resume_versions?flow=create` 直接访问时本 plan 不实现 CreateFlow | flow=create / branch 时 P0 渲染 `<NotImplementedPlaceholder>` + "即将开放" toast；不阻塞 list 主路径；plan 002 / 003 接管时 placeholder 自动替换 |
| R6: backend-resume 真实落地后 fixture 切真时 generated client 类型微变 | adapter 层是 single point of contact；generated client 升级时跑 fixture parity test 自动捕获 |
