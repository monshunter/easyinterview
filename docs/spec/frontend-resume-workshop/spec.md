# Frontend Resume Workshop Spec

> **版本**: 1.2
> **状态**: active
> **更新日期**: 2026-05-23

## 1 背景与目标

[engineering-roadmap §5.2](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 标记 Resume Workshop workstream 候选 subject 包含 `frontend-resume-workshop`（D-X，对齐 [frontend-shell](../frontend-shell/spec.md) 的 D2-D6 ownership）。本 subject 是当前 UI 五入口之一 `resume_versions` 路由的前端 owner：源级复刻 [`ui-design/src/screen-resume-workshop.jsx`](../../../ui-design/src/screen-resume-workshop.jsx) 的完整组件（`ResumeWorkshopScreen` / `ResumeListView` / `ResumeTreeView` / `ResumeFlatView` / `ResumeDetailView` / `ResumeCreateFlow` / `ResumeBranchFlow` / `ResumeVersionRow`），接管 [frontend-shell](../frontend-shell/spec.md) `PlaceholderScreen` D2-D6 占位，并通过 mock-first 路径与 [backend-resume](../backend-resume/spec.md) generated client 集成。

目标：

1. **100% UI 源级复刻**：DOM 构图 / 布局 / 间距 / 字号 / 字体层级 / 控件密度 / 颜色 / 阴影 / 边框 / 圆角 / 状态 / 响应式行为 / 交互节奏必须从 [`ui-design/src/screen-resume-workshop.jsx`](../../../ui-design/src/screen-resume-workshop.jsx) + [`ui-design/src/primitives.jsx`](../../../ui-design/src/primitives.jsx) + [`ui-design/src/app.jsx`](../../../ui-design/src/app.jsx) 直接复刻；不允许重新设计 / 重新解释 / 重新组合视觉。
2. **路由接管**：`resume_versions` 路由从 [frontend-shell](../frontend-shell/spec.md) `PlaceholderScreen` 切换为本 subject 的 `ResumeWorkshopScreen`；route param 兼容 `flow=create / branch` + `versionId` + `tab=preview|rewrites|edit` + `branchOriginalId`。
3. **mock-first 路径**：第一批 plan 不依赖 [backend-resume](../backend-resume/spec.md) 真实落地，通过 [B2 fixtures](../mock-contract-suite/spec.md) `listResumes` / `listResumeVersions` / `getResumeVersion` `default` scenario 完成 happy path。
4. **逐步切真**：backend-resume/002 已于 2026-05-17 完成 resume versions / tailor runs / suggestion decision 真实 handler、fixture parity 与 E2E.P0.074-P0.080 gates；2026-05-23 起，completed frontend plan 001 / 002 / 003 的 scenario trigger 均前置 `frontendOwners.realApiMode.test.ts`，证明 resume generated client 在 `VITE_EI_API_MODE=real` 下使用真实 backend base URL、cookie credentials、无 fixture `Prefer` header 与 side-effect `Idempotency-Key`；fixture-backed mock-first 继续作为 dev/test fallback。
5. **UI parity gate 可执行**：每个 plan 必须含 DOM 锚点 + computed style + bounding box + viewport screenshot smoke；只有 screenshot baseline 可由 clean checkout 稳定取得或本次 gate 明确维护时，才能把 screenshot diff regression 作为完成 gate。参照 [frontend-shell/003-ui-design-pixel-parity-gate](../frontend-shell/plans/003-ui-design-pixel-parity-gate/plan.md) 模式。

本 subject 不实现 backend handler（[backend-resume](../backend-resume/spec.md) / [backend-upload](../backend-upload/spec.md)）；不实现 OpenAPI 契约（[openapi-v1-contract D-18](../openapi-v1-contract/spec.md#31-已锁定决策v100-freeze-范围)）；不恢复旧 Mistakes / Growth / Drill / 独立 Voice / 旧 onboarding / 旧 STAR 等已丢弃模块（[engineering-roadmap §4.1](../engineering-roadmap/spec.md#41-产品与-ui-约束)）。

## 2 范围

### 2.1 In Scope

- **路由壳替换**：替换 [frontend-shell](../frontend-shell/spec.md) `PlaceholderScreen` 中 `resume_versions` 的 D2-D6 占位实现，本 subject 接管渲染。
- **8 个核心组件源级复刻**（按 [`ui-design/src/screen-resume-workshop.jsx`](../../../ui-design/src/screen-resume-workshop.jsx) 顶层 export）：
  1. `ResumeWorkshopScreen` 容器（路由参数解析 + flow 切换）
  2. `ResumeListView`（统计条 + 视图切换 + 调度子视图）
  3. `ResumeTreeView`（原始简历树 + 折叠 + 选为底稿 / 基于树新建）
  4. `ResumeFlatView`（版本平铺 + 排序）
  5. `ResumeDetailView`（Breadcrumb + 版本分支图 + 三标签 preview/rewrites/edit）
  6. `ResumeCreateFlow`（upload / paste / guided 三 tab 创建流）
  7. `ResumeBranchFlow`（分叉配置 + seedStrategy）
  8. `ResumeVersionRow`（版本行复用组件）
- **数据形态映射**：UI 真理源中的 `ResumeSource` 通过 adapter 层映射 OpenAPI `ResumeAsset`；UI `ResumeVersion` 直接对应 OpenAPI 新 schema `ResumeVersion`（[B2 D-18](../openapi-v1-contract/spec.md#31-已锁定决策v100-freeze-范围) + [B1 D-10](../shared-conventions-codified/spec.md#31-已锁定决策)）。
- **i18n**：中英双语，按 [frontend-shell §3](../frontend-shell/spec.md) i18n contract；UI 真理源 `buildResumeData(lang)` / `buildResumePlainText(lang)` / `buildBullets(lang)` 已含双语数据。
- **a11y**：focus trap / aria-label / ESC 关闭 modal / 键盘导航；参照 [frontend-shell C-9 pixel parity](../frontend-shell/spec.md) baseline。
- **UI parity gate**：每个 plan 必须含 DOM 锚点（testid / aria / class）+ computed style + bounding box + viewport screenshot smoke；clean checkout 不依赖 `.gitignore` 排除的本地 screenshot baseline，只有 baseline 可复现/已维护时才触发 screenshot diff regression。
- **Auth boundary**：Resume Workshop 读取 `listResumes` / `listResumeVersions` / `getResumeVersion` 前必须确认当前 runtime auth 为 authenticated；未登录只能显示登录引导 / auth pending action，不得拉取或缓存简历 fixture / real data。
- **隐私红线**：raw resume text / parsed_summary / 简历内容 / questionText 不出现在 console / URL / localStorage / telemetry。

### 2.2 Out of Scope

- backend handler / store / AI 编排：归 [backend-resume](../backend-resume/spec.md) / [backend-upload](../backend-upload/spec.md)。
- OpenAPI 契约升级（D-18 9 个新 op + fixtures）：归 [openapi-v1-contract/004](../openapi-v1-contract/plans/004-resume-additive-coverage/plan.md)。
- 真实 PDF 导出按钮的真实生效（exportResumeVersion P0 是 501 stub）：本 subject `exportPDF` 按钮 P0 仅 toast 兜底；P1 切真时不需修订前端 plan。
- 历史 Mistakes / Growth / Drill / 独立 Voice / 旧 onboarding / 旧 STAR / 旧 experiences 等 retired 模块（[roadmap §4.1](../engineering-roadmap/spec.md#41-产品与-ui-约束)）；本 subject 拒绝任何 backwards-compat 还原。
- 多设备 native shell（iOS / Android / desktop app）：当前 P0 web only。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | UI 真理源唯一性 | [`ui-design/src/screen-resume-workshop.jsx`](../../../ui-design/src/screen-resume-workshop.jsx) + [`ui-design/src/primitives.jsx`](../../../ui-design/src/primitives.jsx) + [`ui-design/src/app.jsx`](../../../ui-design/src/app.jsx) + [`docs/ui-design/resume-module.md`](../../../docs/ui-design/resume-module.md) v1.7 + [`docs/ui-design/resume-onboarding.md`](../../../docs/ui-design/resume-onboarding.md) v1.5 + [`docs/ui-design/jd-resume-management.md`](../../../docs/ui-design/jd-resume-management.md) v1.5 | 不允许从外部品牌设计系统 / AI 审美生成视觉；UI 更新由用户先改 ui-design 再 Agent 迁移 |
| D-2 | 术语映射 adapter 层 | UI 层保留 `ResumeSource` / `ResumeVersion` / `Bullet` 命名；adapter 在 `frontend/src/app/screens/resume-workshop/adapters/` 把 generated client `ResumeAsset` / `ResumeVersion` / `ResumeTailorSuggestion` 映射到 UI 类型；不重命名 generated client 类型 | [B1 D-10](../shared-conventions-codified/spec.md#31-已锁定决策) + [B2 D-18](../openapi-v1-contract/spec.md#31-已锁定决策v100-freeze-范围) 已锁，前端 adapter 层是唯一映射点 |
| D-3 | 路由参数语义 | `resume_versions` route 支持参数：`flow=create | branch | list（默认）` + `versionId`（详情打开）+ `tab=preview | rewrites | edit`（详情子标签，默认按 `resumeDefaultTab(version)`：MASTER→preview / TARGETED→rewrites）+ `branchOriginalId`（branch 流程进入时携带）；通过 [frontend-shell normalizeRoute](../frontend-shell/spec.md) 验证 | 与 UI 真理源 `ResumeWorkshopScreen` flow 参数对齐；route param 与 UI state 一一对应 |
| D-4 | mock-first + real-backend handoff | 第一批 plan（001-listing-routing-and-detail-readonly）依赖 listResumes / listResumeVersions / getResumeVersion fixture；不依赖真实 backend；backend-resume/002 已完成 9 个 versions/tailor/suggestion op 的真实 handler、fixture parity 与 E2E.P0.074-P0.080 gates，后续 002 / 003 默认以 generated client real backend 为目标，fixture-backed dev preview 继续作为本地 fallback | 与 [mock-contract-suite D-5 Vite dev preview 默认 fixture-backed](../mock-contract-suite/spec.md#3-用户决策--待确认事项) 对齐；解除 frontend-resume-workshop 002/003 的 backend 等待条件 |
| D-5 | UI parity gate 强制 | 每个 plan 必须含至少 4 类断言：（a）DOM 结构 parity（testid / aria / 嵌套）；（b）computed style parity（颜色 / 字号 / 间距 / 阴影）；（c）bounding box parity（关键元素位置）；（d）desktop + mobile viewport screenshot smoke；参照 [frontend-shell/003-ui-design-pixel-parity-gate](../frontend-shell/plans/003-ui-design-pixel-parity-gate/plan.md) 模式。Clean checkout 常规 PASS 不依赖未跟踪 screenshot baseline；只有 baseline 可由 CI / checkout 稳定取得或本次显式维护时，screenshot diff regression 才是完成 gate | 不允许"视觉相似""风格接近"作为完成依据 |
| D-6 | PDF 导出按钮 P0 stub | `exportPDF` 按钮在 P0 将 `exportResumeVersion` 的 `501 + RESUME_EXPORT_NOT_AVAILABLE` 或等价本地不可用分支映射为 toast "PDF 导出能力即将开放"；调用必须通过 `frontend/src/lib/conventions/idempotency.ts::generateIdempotencyKey()` 传入 generated client `opts.idempotencyKey`，并由测试断言 `Idempotency-Key` header；`copyText` 按钮 P0 真实可用（通过 `buildResumePlainText` 投影）；P1 backend-resume 003 真实落地后前端按钮自动消费 | [B2 D-18](../openapi-v1-contract/spec.md#31-已锁定决策v100-freeze-范围) `exportResumeVersion` P0 501 + `RESUME_EXPORT_NOT_AVAILABLE` 兜底 |
| D-7 | 旧入口负向 grep | `frontend/src/app/screens/resume-workshop/` 不出现 `welcome` / `mistake` / `growth` / `plan` / `drill` / `followup` / 旧 `onboarding` / 旧 `STAR` / 旧 `experiences` / `voice` 路径或 testid（除 normalizeRoute alias map）；不 import `ui-design/src/data.jsx` 或 `ui-design/src/screen-resume-workshop.jsx` 作为运行时数据 / 组件源 | 防止 retired 模块复活；避免正式前端运行时耦合静态原型 |

### 3.2 待确认事项

- `ResumeCreateFlow` 的"轻量问答 guided"模式是否在 P0 实现：已锁定 P0 实现，由 [plan 002](./plans/002-create-flow-and-onboarding/plan.md) 落地；UI 真理源 [`resume-onboarding.md`](../../../docs/ui-design/resume-onboarding.md) v1.5 已设计完成；backend-resume/002 已解除 versions/tailor/suggestion 后端等待条件，frontend 002 启动前仍需核对 backend-upload `createUploadPresign` / `registerResume` handoff 的当前真实状态，不私造客户端协议。
- accept/reject suggestion 是否需要独立确认弹窗：已锁定 **不新增独立 ConfirmDialog**；按 UI 真理源 [`ResumeRewritesTab`](../../../ui-design/src/screen-resume-workshop.jsx) 的 inline `拒绝 / 编辑 / 采纳` 操作落地，并通过 terminal-state toast / aria-live 反馈结果。当前 generated client 的 `acceptResumeTailorSuggestion` / `rejectResumeTailorSuggestion` 为 bodyless operation，不支持 `manualEditText` request body；manual edit 由 plan 003 Phase 4.3 先显式调用 `updateResumeVersion` patch `structuredProfile.manualEdits[]`，再用 bodyless accept 将 suggestion 标为终态。如需扩展 accept body，由用户先更新 OpenAPI / backend / `ui-design/` 后再修订本 spec。
- 首页 "1 分钟创建简历" 链接的 deep link 形式：已锁定 `nav("resume_versions", { flow: "create" })`，由 plan 002 Phase 6.1 集成验证；不携带初始 sourceType（用户在 CreateFlow 内选择 tab）；如未来需要额外携带 createMode hint 由 spec 修订。

## 4 设计约束

### 4.1 UI 真理源约束

- 任何视觉 / DOM 结构 / token 必须能追溯到 [`ui-design/src/screen-resume-workshop.jsx`](../../../ui-design/src/screen-resume-workshop.jsx) / [`ui-design/src/primitives.jsx`](../../../ui-design/src/primitives.jsx) / [`ui-design/src/app.jsx`](../../../ui-design/src/app.jsx) 的具体函数或常量；不允许 AI 补齐未在原型中出现的视觉值。
- UI 升级流程：用户先更新 [`ui-design/`](../../../ui-design/) + [`docs/ui-design/`](../../../docs/ui-design/)，再创建 frontend-resume-workshop follow-up plan；Agent 不擅自 enhance 视觉。

### 4.2 数据真理源约束

- 运行时数据来源：generated client（消费 [B2 fixtures](../mock-contract-suite/spec.md) 或 [backend-resume](../backend-resume/spec.md) real handler）；不 import `ui-design/src/data.jsx` 或 `ui-design/src/screen-resume-workshop.jsx` 作为运行时数据源。Resume Workshop 原型投影源当前在 `screen-resume-workshop.jsx`；只有当 [B2 prototype mapping](../../../openapi/fixtures/PROTOTYPE_MAPPING.md) 与同步工具明确支持该源文件时，才允许生成对应 `prototype-baseline`。
- 术语 adapter：`frontend/src/app/screens/resume-workshop/adapters/` 是唯一的 `ResumeAsset ↔ ResumeSource` / `ResumeVersion ↔ ResumeVersion` 映射点；业务组件不直接消费 generated client 类型。

### 4.3 隐私约束

- raw resume text / parsed_summary / structured_profile 内容 / suggestion 改写文本不出现在 console.log / URL query / localStorage / fixture transport 日志 / telemetry payload（仅在用户主动 copy / export 时通过 clipboard / blob 流出，且不持久到非 user-owned 存储）。
- pendingAction.params 不携带 raw text（[frontend-shell §4](../frontend-shell/spec.md) auth pending action contract）。
- 未登录状态不得触发 protected Resume endpoints；登录恢复只允许携带 route params（`flow` / `versionId` / `tab` / `branchOriginalId`），不得把原始文本、解析快照或结构化简历放入 pendingAction / URL。

### 4.4 BDD / TDD 约束

- Vitest 单元测试：组件 render / 路由 hook / adapter 映射 / i18n / a11y。
- Playwright E2E：DOM 结构 + computed style + bounding box + viewport screenshot smoke 多端断言；仅在稳定 baseline 可用时启用 screenshot diff regression。
- 每个 plan 必须维护 BDD gate；BDD 不适用的内部纯 UI 组件除外（少见）。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| `resume_versions` 路由组件 | frontend-resume-workshop | `ResumeWorkshopScreen` 接管 PlaceholderScreen |
| UI 真理源 jsx / 数据形态 | [`ui-design/`](../../../ui-design/) + [`docs/ui-design/`](../../../docs/ui-design/) | 视觉 / 数据 schema 来源 |
| route 别名 / pendingAction | [frontend-shell](../frontend-shell/spec.md) | 不变 |
| generated TS client | [B2 openapi-v1-contract](../openapi-v1-contract/spec.md) + frontend-resume-workshop adapter | adapter 层做术语映射 |
| 上传 UI 触发 createUploadPresign | [frontend-shell](../frontend-shell/spec.md) + frontend-resume-workshop/002 | 消费 [backend-upload](../backend-upload/spec.md) fixture 或 real |
| Resume Picker Modal（workspace 中） | [frontend-workspace-and-practice](../frontend-workspace-and-practice/spec.md) | 独立 owner，本 subject 不重复实现 |
| backend handler / store / AI | [backend-resume](../backend-resume/spec.md) / [backend-upload](../backend-upload/spec.md) | backend-resume/002 versions/tailor/suggestion real paths 已就位；frontend 002/003 切真时继续通过 generated client 和 fixture parity 同步 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 路由替换 | `App.tsx` 路由表已修订 | 加载 `resume_versions` route | 渲染 `ResumeWorkshopScreen` 而非 `PlaceholderScreen`；TopBar `topbar-nav-resume_versions` 高亮 | 001-listing-routing-and-detail-readonly |
| C-2 | ResumeListView 主路径 | 用户已登录；fixture `listResumes` / `listResumeVersions` `default` scenario | 加载 list 默认视图 | StatsStrip + ViewSwitcher + ResumeTreeView 渲染；DOM testid 含 `resume-workshop-stats-{originals,versions,top-match,recent}` + `resume-workshop-view-switcher-{tree,flat}` + `resume-tree-row-{id}` | 001 |
| C-3 | Tree / Flat 切换 | 处于 tree 默认视图 | 点击 `resume-workshop-view-switcher-flat` | 切换到 ResumeFlatView，DOM 渲染 `resume-flat-row-{id}` 按 match / updated_at 排序 | 001 |
| C-4 | ResumeDetailView 只读详情 | 点击某 version row | 进入详情 `/resume_versions?versionId={id}` | 渲染 Breadcrumb + 版本分支图 + 三 tab 切换；MASTER 默认 `preview`，TARGETED 默认 `rewrites`（001 阶段内容可为 P0 ComingSoon，但 route/tab 状态不得改成 preview）；Preview Tab 可手动打开并显示 "查看原件" 弹层，可关闭 + ESC 关闭 + focus trap；点击导出 PDF 时请求带 `Idempotency-Key` 并渲染 P0 501 toast | 001 |
| C-5 | UI parity gate | playwright pixel parity 套件已配置 | 跑 desktop + mobile viewport parity | DOM / computed style / bounding box / screenshot smoke 与 ui-design 源一致；只有 baseline 可由 clean checkout 稳定取得时才使用 screenshot diff；测试 fail 时输出可定位 artifact | 001 + 后续 plan |
| C-6 | mock-first + real-mode 字节比对 | mock-contract-suite 已配置 listResumes / listResumeVersions / getResumeVersion fixture；backend-resume real handlers 已落地 | 切换 mock transport 或 `VITE_EI_API_MODE=real` | response 字段集 / status / shape 与 generated client 期望字节一致；组件不 import prototype data；P0.081-P0.087 trigger 必须先跑 `frontendOwners.realApiMode.test.ts`，证明 resume / resume-version / tailor operation 指向真实 backend client surface | 001 / 002 / 003 |
| C-7 | i18n 切换 | EN / ZH lang toggle | 切换 lang | 关键文案 / `buildResumeData(lang)` 输出 / TopBar lang menu 同步；`Accept-Language` header 携带 | 001 |
| C-8 | 隐私红线 | raw resume text / parsed_summary | 用户浏览 list / detail | console / URL / localStorage / telemetry 不出现敏感内容；仅 copyText 通过 clipboard 流出 | 001 + 后续 plan |
| C-9 | 旧入口负向 | grep `frontend/src/app/screens/resume-workshop/` | – | 不出现 `welcome` / `mistake` / `growth` / `plan` / `drill` / `followup` / 旧 `onboarding` / 旧 `STAR` / 旧 `experiences` / `voice` 路径或 testid；不 import `ui-design/src/data.jsx` / `ui-design/src/screen-resume-workshop.jsx` 作为运行时依赖 | 001 + 后续 plan |
| C-10 | CreateFlow 三 tab + Onboarding | 未登录或首次访问 + flow=create | 三 tab 分别完成 register | upload / paste / guided 三路径 happy path + Agent Parsing loading + Preview Confirm `confirmResumeStructuredMaster` 保存 v1 → list；与 `WorkspaceMissingResumeState` / 首页 "1 分钟创建" CTA 串通；隐私红线 raw text / file binary / guidedAnswers / parsedSummary 不出现在 console / URL / pendingAction / localStorage / mock transport log | [002-create-flow-and-onboarding](./plans/002-create-flow-and-onboarding/plan.md) |
| C-11 | BranchFlow + Rewrites Tab + Edit Tab | 当前在某 version 详情 | 触发 branch / 切到 rewrites / 切到 edit | branch 配置 + 3 seedStrategy + accept/reject/manual edit 终态 + tailor run polling + Edit Tab `updateResumeVersion`；exportPDF P0 toast / copyText 真实可用一致性；retired tailor mode (`inline\|rewrite\|mirror`) 0 命中（与 [B3 D-14](../event-and-outbox-contract/spec.md#31-已锁定决策含-jobtype-映射表) 同步） | [003-branch-rewrites-and-edit](./plans/003-branch-rewrites-and-edit/plan.md) |

## 7 关联计划

- [001-listing-routing-and-detail-readonly](./plans/001-listing-routing-and-detail-readonly/plan.md)：第一批 plan，路由接管 + ResumeListView（TreeView + FlatView + StatsStrip + ViewSwitcher）+ ResumeDetailView Preview Tab 只读 + 原件弹层 + Breadcrumb + 版本分支图 + i18n + a11y + UI parity gate；BDD 覆盖列表 / 树/平铺切换 / 详情预览主路径（E2E.P0.036 + E2E.P0.037）。
- [002-create-flow-and-onboarding](./plans/002-create-flow-and-onboarding/plan.md)：替换 `flow=create` placeholder，源级复刻 `ResumeCreateFlow` 三 tab（upload / paste / guided）+ `ResumeParseFlow` 7 step 动画 + `getResume` 轮询 + `ResumePreviewConfirm` 保存 v1（调 [backend-resume/002 D-10 `confirmResumeStructuredMaster`](../backend-resume/plans/002-versions-tailor-runs-and-save-v1/plan.md#phase-1-b2-d-18-additive-confirmresumestructuredmaster--b1-错误码增补)）+ 双步上传（[backend-upload `createUploadPresign`](../backend-upload/plans/001-file-objects-and-presign-baseline/plan.md) + 浏览器 PUT + `registerResume`）+ 首页 "1 分钟创建" / `WorkspaceMissingResumeState` CTA 串通 + auth pending action 隐私（不携带 form draft / raw text）；BDD 覆盖 `E2E.P0.081 / P0.082 / P0.083` 三场景。
- [003-branch-rewrites-and-edit](./plans/003-branch-rewrites-and-edit/plan.md)：替换 `flow=branch` placeholder + 替换 plan 001 阶段 Rewrites / Edit Tab `<ComingSoonTab>` 占位，源级复刻 `ResumeBranchFlow`（3 seedStrategy 同步 + ai_select 触发 tailor）+ `ResumeRewritesTab`（suggestions 列表 + accept/reject/manual edit 终态状态机 + tailor run polling + 重新运行改写 mode 切换）+ `ResumeEditTab`（headline + summary 提交 `updateResumeVersion`）+ exportPDF / copyText 一致性；BDD 覆盖 `E2E.P0.084 / P0.085 / P0.086 / P0.087` 四场景。
