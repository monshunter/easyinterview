# Frontend Resume Workshop Branch, Rewrites and Edit

> **版本**: 1.3
> **状态**: active
> **更新日期**: 2026-06-14

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

> 2026-05-23 L2 real-backend gate remediation：P0.084-P0.087 trigger 已前置 `frontendOwners.realApiMode.test.ts`，verify 检查 `VITE_EI_API_MODE=real`、默认 backend base URL 与测试文件 marker；branch/rewrites/edit 的 fixture-backed UI variants 继续保留，但真实 resume-version / tailor / suggestion / export generated-client routing 由集中 gate 证明。

> 2026-06-14 D-20 Phase 8 L2 收口修订：本 plan 早期 Phase 1-7 保留为历史实现记录，当前 active 验收以 D-20 flat resume contract 为准。`ResumeBranchFlow` / `branchResumeVersion` / `seedStrategy` / `acceptResumeTailorSuggestion` / `rejectResumeTailorSuggestion` / `updateResumeVersion` / `resumeVersionId` / `resumeAssetId` 均为退役负向项；当前用户路径是 flat `ResumeDetailView` + `ResumeRewritesTab` accept-only + `RewriteSaveConfirmModal`，保存通过 `updateResume` 覆盖或 `duplicateResume` 另存。

## 1 目标

把 [frontend-resume-workshop spec](../../spec.md) §6 C-11（D-20 Rewrites Tab + Edit Tab + 改写采纳保存）落到当前 `frontend/` 实现，并完成 Phase 8 收口：

- `resume_versions` 当前只支持 flat list / create / detail 三类入口：`flow=branch`、`versionId`、`branchOriginalId` 不 materialize；未知 flow 回落到 flat list。
- `ResumeDetailView` 通过 `getResume(resumeId)` 读取 flat resume；detail 子容器必须保留 route `targetJobId`，使 `requestResumeTailor` rerun 在有 JD 上下文时发送 `{ resumeId, targetJobId, mode }`，且绝不恢复 `resumeAssetId` / `resumeVersionId`。
- `ResumeRewritesTab` 按 UI 真理源展示 ephemeral suggestions，只提供「采纳」与 `RewriteSaveConfirmModal`；保存分支为 `updateResume` 覆盖当前简历或 `duplicateResume` 另存新简历，均带 `Idempotency-Key`。
- accepted rewrite 合并必须覆盖 D-20 flat profile 与历史 parse 输出的 bullet 容器：`sections[]`、`experience[]`、`experiences[]`、`projects[]`。`structuredProfile` 被响应省略时，原件 fallback 不得崩溃。
- `ResumeEditTab` 手动编辑 `displayName` / `headline` / `summary` 后调用 `updateResume(resumeId, { displayName?, structuredProfile? })`；422 inline、409 IK conflict、404 cross-user error 都以当前 flat hook 映射。
- Export PDF / copyText 复用 plan 001 flat `exportResume` / `buildResumePlainText` 路径，Rewrites / Edit 切换不退化。
- i18n / a11y / privacy / UI parity gate 仍覆盖 `ResumeDetailView` / `ResumeRewritesTab` / `ResumeEditTab` desktop + mobile；负向 grep 覆盖所有退役版本树、branch、suggestion decision、version operation 词。
- 本 plan 不修改 backend handler / 异步 job / outbox event / AI 调用；tailor 真实运行由 backend-resume D-20 flat handlers 和 P0.077-P0.080 证明，本 plan 只消费 generated client。

## 2 背景

本 plan 是 frontend-resume-workshop 第三批 plan。2026-05 初始版本服务于版本树 / BranchFlow；2026-06-13 D-20 已把产品与 UI 重塑为 flat resume IA，因此当前 Phase 8 负责把旧实现记录收束到新主路径："打开 flat 简历详情 → AI 生成改写建议 → 本地采纳 → 覆盖原简历或另存新简历 → 手动编辑 flat profile"。

- D-20 后不存在原始简历树、主版本、岗位定制版本、分叉流程或 suggestion terminal decision API；历史 Phase 1-7 的 branch/version 叙述仅保留为迁移背景，不能作为当前验收依据。
- Rewrites Tab 当前消费 `requestResumeTailor` / `getResumeTailorRun` 的 ephemeral suggestions；UI 只记录本地 accepted rewrites，落盘时通过 `updateResume` 或 `duplicateResume` 写回 flat `Resume.structuredProfile`。
- Edit Tab 当前通过 `updateResume` 修改 flat resume 的 `displayName` / `structuredProfile.headline` / `structuredProfile.summary`，不再调用 `updateResumeVersion`。
- 与 [backend-resume D-20](../../../backend-resume/spec.md) / [openapi-v1-contract D-26](../../../openapi-v1-contract/spec.md) 协作：generated client 当前 surface 为 `listResumes` / `getResume` / `updateResume` / `duplicateResume` / `exportResume` / `requestResumeTailor` / `getResumeTailorRun`；退役 operation 通过 codegen、fixture、scenario 和 grep gate 保持 0 回流。
- BUG-0123 证明旧 gate 太窄：只看 `sections[].bullets`、root route attr 或有 `structuredProfile` 的 fixture 不足以证明 Phase 8。当前验收必须包含 flat profile bullet 容器、omitted `structuredProfile` fallback、以及 rerun body-level `targetJobId` preservation。

每个 phase 的当前执行含义：Phase 1-7 是历史增量与回归线索；Phase 8 是 D-20 收口 owner，必须用当前代码、scenario wrappers、BDD 文档和负向搜索重新证明 flat contract。

执行本 plan 前必须确认：

- [frontend-resume-workshop/001-listing-routing-and-detail-readonly](../001-listing-routing-and-detail-readonly/plan.md) completed；ResumeDetailView 三 tab 容器已就位（Preview tab 真实、Rewrites / Edit `<ComingSoonTab>` 占位、tab 切换 / URL `tab` param 行为已稳定）。
- [frontend-resume-workshop/002-create-flow-and-onboarding](../002-create-flow-and-onboarding/plan.md) 当前分支实现已落地；`flow=create` 渲染 `ResumeCreateFlow`，Home / Workspace CTA handoff 已可走通。003 Phase 0 只需反查当前代码事实与 plan 002 lifecycle 状态，避免把旧 placeholder 当作仍存在的 create 主路径。
- [backend-resume](../../../backend-resume/spec.md) / [openapi-v1-contract](../../../openapi-v1-contract/spec.md) D-20 contract collapse 已完成，generated client 含当前 flat Resume / ResumeTailor ops，退役 version-tree ops 不得作为 frontend 003 当前依赖。
- UI 真理源 [`ui-design/src/screen-resume-workshop.jsx`](../../../../../ui-design/src/screen-resume-workshop.jsx) 的 `ResumeDetailView / ResumeRewritesTab / RewriteSaveConfirmModal / ResumeEditTab` + [`docs/ui-design/resume-module.md`](../../../../ui-design/resume-module.md) v2.0 + [`docs/ui-design/jd-resume-management.md`](../../../../ui-design/jd-resume-management.md) v2.0 active。

## 3 质量门禁分类

- **Plan 类型**: `code-internal + feature-behavior`。本 plan 当前 Phase 8 实现 flat resume detail / rewrites / edit 用户可见 UI 行为，并消费 generated client flat Resume / ResumeTailor operations。
- **TDD 策略**: 适用。Red-Green-Refactor 入口：
  1. Vitest 组件单测：`ResumeWorkshopScreen.test.tsx` 证明 `flow=branch` 不 materialize；`ResumeDetailView.test.tsx` 覆盖 detail tab、omitted `structuredProfile` fallback、accepted rewrites flat profile merge、rerun body context；`ResumeRewritesTab.test.tsx` 覆盖 accept-only + save modal；`ResumeEditTab.test.tsx` 覆盖 `updateResume` save / 422。
  2. hook / adapter test：`useRequestResumeTailor.test.tsx` 覆盖 `{ resumeId, targetJobId?, mode }` + IK replay/rotation/error mapping；`useResumeTailorRunPolling.test.tsx` 覆盖 queued/generating/ready/failed/timeout/unmount cleanup；`ResumeDetailFixtureParity.test.tsx` 覆盖 flat fixture parity。
  3. flat profile regression test：accepted rewrites 必须写入 `sections[]`、`experience[]`、`experiences[]`、`projects[]` 的 `bullets: string[]`；保存 payload 不得写入 `acceptedRewrites` 临时字段。
  4. route context test：当 route 带 `targetJobId`，detail rerun 必须发送 `targetJobId`；无 `targetJobId` 时才允许 generic rerun；任何路径不得发送 `resumeAssetId` / `resumeVersionId`。
  5. Idempotency-Key contract test：`requestResumeTailor` / `updateResume` / `duplicateResume` / `exportResume` 均通过 `generateIdempotencyKey()` 或对应 hook 生成 IK，request spy 断言 `Idempotency-Key` 出现。
  6. 隐私 grep test：originalBullet / suggestedBullet / matchSummary / structuredProfile / manual edit 文本不出现在 URL / pendingAction params / localStorage / mock transport log / console / toast。
  7. Playwright pixel parity：flat `ResumeDetailView` / `ResumeRewritesTab` / `ResumeEditTab` desktop 1440px + mobile 390x844 DOM anchor + computed style + bounding box + screenshot smoke（baseline 可复现时启用 screenshot diff）。
  8. negative grep test：runtime Resume Workshop source 不出现 retired 模块名、`ResumeBranchFlow` / `branchResumeVersion` / `seedStrategy` / `acceptResumeTailorSuggestion` / `rejectResumeTailorSuggestion` / `updateResumeVersion` / `resumeVersionId` / `resumeAssetId`；不出现 retired tailor mode `inline | rewrite | mirror`；不 import `ui-design/src/screen-resume-workshop.jsx` / `ui-design/src/data.jsx` 作为运行时依赖。

  执行入口：`/implement frontend-resume-workshop/003-branch-rewrites-and-edit` → `/tdd`。

- **BDD 策略**: 适用（Feature plan requires BDD）。`E2E.P0.084` retired branch flow + flat route regression、`E2E.P0.085` flat Rewrites tailor polling + rerun、`E2E.P0.086` accept-only save + Edit Tab `updateResume`、`E2E.P0.087` export/copy + flat detail parity + retired negative，详见 [bdd-plan.md](./bdd-plan.md) / [bdd-checklist.md](./bdd-checklist.md)。

- **替代验证 gate**:
  - `pnpm --filter @easyinterview/frontend test` (Vitest)
  - `pnpm --filter @easyinterview/frontend build` + `pnpm --filter @easyinterview/frontend test:pixel-parity` (Playwright)
  - `git grep -nE "welcome|mistake|growth|drill|followup|STAR|ExperiencesScreen|experiences-route|voice|OnboardingScreen|onboarding=true|ResumeBranchFlow|branchResumeVersion|seedStrategy|acceptResumeTailorSuggestion|rejectResumeTailorSuggestion|updateResumeVersion|resumeVersionId|resumeAssetId" -- frontend/src/app/screens/resume-workshop/`（旧入口 / 版本树 / 退役 operation negative；当前 plan 文档 prose 不纳入 raw zero-hit；不得禁用 D-20 flat profile `experiences[]` 字段）
  - `git grep -nE "(^|[^A-Za-z0-9_])(inline|rewrite|mirror)([^A-Za-z0-9_]|$)" -- frontend/src/app/screens/resume-workshop/tabs/`（retired tailor mode negative；与 B3 D-14 同步）
  - `git grep -nE "ui-design/src/(data|screen-resume-workshop)" -- frontend/src/app/screens/resume-workshop/tabs/ frontend/src/app/screens/resume-workshop/components/`（原型 runtime import negative）
  - `sync-doc-index --check`

### 3.1 Frontend / Backend Operation Matrix

本 plan 走 fixture-backed frontend + real-backend generated-client preflight path：flat detail read / 改写运行 / 改写保存 / 结构化编辑 / 导出与复制共五类操作。这里的 fixture-backed dev/test 只表示可用当前 OpenAPI fixture 做前端确定性验证，不表示可以绕过 backend-resume / OpenAPI 对 operation / fixture / generated artifact 的收敛要求。

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `listResumes` | `openapi/fixtures/Resumes/listResumes.json` `default` | `ResumeListView` / `useResumeAssets` flat list | `backend/internal/resume/handler/list.go` + `cmd/api` `GET /api/v1/resumes` real route ready | `resumes` read | none | E2E.P0.084 |
| `getResume` | `openapi/fixtures/Resumes/getResume.json` `default` / `not-found-404` | `ResumeDetailView` / `useResumeAsset`; detail fallback must tolerate omitted `structuredProfile` | `backend/internal/resume/handler/get.go` + `cmd/api` `GET /api/v1/resumes/{resumeId}` real route ready | `resumes` read | none | E2E.P0.084 + P0.086 + P0.087 |
| `requestResumeTailor` | `openapi/fixtures/ResumeTailor/requestResumeTailor.json` `default` / `idempotency-replay` | `useRequestResumeTailor` via `ResumeRewritesTab` rerun; request body `{ resumeId, targetJobId?, mode }` + IK | `backend/internal/resume/handler/request_tailor.go` + `cmd/api` `POST /api/v1/resume/tailor` real route ready | `resume_tailor_runs(queued)` + `async_jobs(resume_tailor)` | downstream `resume.tailor` | E2E.P0.085 |
| `getResumeTailorRun` | `openapi/fixtures/ResumeTailor/getResumeTailorRun.json` `default` (ready) / `queued` / `generating` / `failed` | `useResumeTailorRunPolling`; read-only no IK; ready suggestions refresh Rewrites surface | `backend/internal/resume/handler/get_tailor_run.go` + `cmd/api` `GET /api/v1/resume/tailor-runs/{tailorRunId}` real route ready | `resume_tailor_runs` read | none in read path | E2E.P0.085 |
| `updateResume` | `openapi/fixtures/Resumes/updateResume.json` `default` / `idempotency-replay` / `validation-error-422` | `useResumeSave.overwrite`; Rewrites overwrite and Edit Tab save; flat profile merge covers sections/experience/experiences/projects | `backend/internal/resume/handler/update.go` + `cmd/api` `PATCH /api/v1/resumes/{resumeId}` real route ready | `resumes` UPDATE | none | E2E.P0.086 |
| `duplicateResume` | `openapi/fixtures/Resumes/duplicateResume.json` `default` / `idempotency-replay` / `validation-error-422` | `useResumeSave.saveAsNew`; `RewriteSaveConfirmModal` save-as-new path | `backend/internal/resume/handler/duplicate.go` + `cmd/api` `POST /api/v1/resumes/{resumeId}/duplicate` real route ready | `resumes` INSERT from flat source snapshot | none | E2E.P0.086 |
| `exportResume` | `openapi/fixtures/Resumes/exportResume.json` `p0-501-not-available` | plan 001 `ResumeDetailExport` path reused from Rewrites/Edit header; IK + 501 friendly toast | P0 explicit 501 stub; P1 backend-resume export owner | none in P0 | none | E2E.P0.087 |

### 3.2 上游依赖 gate（D-20 当前事实）

- 反查 [backend-resume](../../../backend-resume/spec.md) 与 [openapi-v1-contract](../../../openapi-v1-contract/spec.md) D-20 当前事实：flat `listResumes / getResume / updateResume / duplicateResume / exportResume / requestResumeTailor / getResumeTailorRun` generated client + fixtures + real handlers ready。
- `make codegen-check` 必须证明退役 `branchResumeVersion / acceptResumeTailorSuggestion / rejectResumeTailorSuggestion / updateResumeVersion / getResumeVersion / listResumeVersions` generated surface 不回流。
- P0.084-P0.087 trigger 必须前置 `frontendOwners.realApiMode.test.ts`，verify 检查 real-mode marker、默认 backend base URL、无 fixture `Prefer` header 和 side-effect IK。
- 若 flat fixture / generated client / handler 任一缺失，Phase 8 暂停并转回对应 owner；frontend 003 不私造客户端协议。

## 4 实施步骤

### Phase 0: 上游依赖 gate + retired drift baseline

#### 0.1 上游 fixture / handler 状态确认
- 确认 backend-resume/002 Phase 4..8 当前事实仍成立：`branchResumeVersion / requestResumeTailor / getResumeTailorRun / acceptResumeTailorSuggestion / rejectResumeTailorSuggestion / updateResumeVersion` generated server interface、real handler 与 `cmd/api` route 均存在；
- 确认 fixture 名称与 error envelope 形态：accept / reject fixture 必须包含 `default / idempotency-replay / already-decided-409`，且 409 body 为 `error.code='VALIDATION_FAILED'` + `error.details.reason='SUGGESTION_ALREADY_DECIDED'`；requestTailor 必须含 `default / idempotency-replay` 且 request header 带 `Idempotency-Key`；getTailorRun 必须含 `queued / generating / default(ready) / failed`。若任一事实缺失，Phase 4 / 5 / E2E.P0.085-P0.086 暂停并转回 backend owner 修复，不允许 frontend plan 以旧 envelope 或 synthetic schema 收口。

#### 0.2 retired drift baseline
- `git grep -nE "(^|[^A-Za-z0-9_])(inline|rewrite|mirror)([^A-Za-z0-9_]|$)" -- frontend/src/app/screens/resume-workshop/`：0 命中前置；
- `git grep -nE "welcome|mistake|growth|drill|followup|STAR|ExperiencesScreen|experiences-route|voice|OnboardingScreen|onboarding=true" -- frontend/src/app/screens/resume-workshop/`：0 命中前置（plan 001 / 002 阶段已 enforce，Phase 7.10 收口时再次验证；D-20 flat profile `experiences[]` 字段允许存在）。

### Phase 1: ResumeBranchFlow 容器 + 路由 + auth gate

#### 1.1 替换 plan 001 中 `flow=branch` 的 `<NotImplementedPlaceholder>`
- 修订 `frontend/src/app/screens/resume-workshop/ResumeWorkshopScreen.tsx`：当 `flow === 'branch'` 时渲染新增的 `ResumeBranchFlow`，传入 `branchOriginalId` 解析得到的 `original` + `master` 上下文。
- placeholder 仅保留给 `flow === 'create'` 不可达的边界 fallback（plan 002 已替换主路径）。

#### 1.2 实现 `frontend/src/app/screens/resume-workshop/branch/ResumeBranchFlow.tsx`
- 路由 param 解析：`branchOriginalId` 必填；缺失时显示 `<NotImplementedPlaceholder>` 或返回 list（与 UI 真理源一致）；
- 数据获取：复用 plan 001 `listResumes` + `listResumeVersions` adapter 找到 `original` 与对应 MASTER `version`；如 original 不存在 / cross-user 隔离 → 渲染 NotFound CTA；
- 表单 state：`name / target / focus / seed`；表单校验 `canSubmit = name.trim().length > 0 && target.trim().length > 0`；
- 渲染：源级复刻 UI 真理源 BranchFlow header + BRANCHING FROM 卡 + 版本名称 / 目标岗位 / 公司 input + 侧重方向 chip group + Bullet 初始化 3 卡 + 底部 actions（取消 / 创建版本）。

#### 1.3 auth gate
- 未登录访问 `resume_versions?flow=branch&branchOriginalId={id}` 渲染 auth gate；pendingAction 只携带 `{ route: 'resume_versions', params: { flow: 'branch', branchOriginalId } }`，不携带表单 draft（name / target / focus / seed）；
- 登录恢复：进入 BranchFlow 默认 seed='copy_master' + focus='platform'，与 UI 真理源默认一致。

#### 1.4 Vitest 组件单测
- 表单校验 / canSubmit / 3 seedStrategy 切换 / focus 选项切换 / cancel / Vitest mock client 0 个 protected API 请求 in auth gate；至少 ≥ 8 case PASS。

### Phase 2: branchResumeVersion 三 seedStrategy 提交 + IK + nav 行为

#### 2.1 实现 `branch/hooks/useResumeBranchSubmit.ts`
- 入参：`{ parentVersionId, targetJobId, seedStrategy, displayName, focusAngle }`；
- 行为：`generateIdempotencyKey()` → generated client `branchResumeVersion(payload, { idempotencyKey })`；
- 三态响应路径：
  - `copy_master / blank` 同步 201 + `ResumeVersion` → toast + nav；
  - `ai_select` 202 + `BranchResumeVersionAccepted` → toast + nav + 返回 `tailorRunId` 供 Rewrites Tab polling 启动；
- 失败路径：
  - 422 `VALIDATION_FAILED`（缺字段 / 不可识别 seed）：inline 错误；
  - 404 parent / targetJob 不存在：toast `底稿来源或目标岗位不可用 · 请返回上一步` + 不 nav；
  - IK replay：返回首次 ResumeVersion；不重复 nav；
- IK request spy 断言 `Idempotency-Key` header 出现且与同一表单 retry 复用至成功或 422 inline。

#### 2.2 BranchFlow 提交触发
- "创建版本" 按钮 click → useResumeBranchSubmit；成功后：
  - `copy_master` → `nav("resume_versions", { versionId: newId, tab: 'rewrites' })`（与 [resumeDefaultTab](../../../../../ui-design/src/screen-resume-workshop.jsx) TARGETED 默认对齐）；
  - `blank` → `nav("resume_versions", { versionId: newId, tab: 'edit' })`（空白起步默认进入手动编辑面板，避免在空 suggestions 列表下显示 Rewrites Tab）；
  - `ai_select` → 同 copy_master nav 路径 + 在 Rewrites Tab 启动 polling（hook 内置）；
- 失败：保留表单 state 不重置。

#### 2.3 路径选项 / focus enum 映射
- UI 真理源 focus options（platform / collaboration / fullstack / leadership / custom）映射为 `focusAngle` 字符串（UI 真理源 label 已经包含可发送字面量）；custom 选项保留 `focusAngle` 为空字符串或 UI 自定义输入（P0 不实现 custom 输入框；维持选项可选，custom 时仅作为 UI label）；
- targetJobId：UI 真理源 BranchFlow 目前接受 free-form `target` 字符串；本 plan P0 仍以字符串形式承载 `targetJobId`（与 OpenAPI `BranchResumeVersionRequest.targetJobId` 字段对齐）；如未来引入 TargetJob 选择器由 follow-up plan 落地，本 plan 不阻塞。

#### 2.4 Vitest 单测
- 三 seedStrategy 提交 + nav target；replay；422 inline；404 parent / targetJob；至少 ≥ 8 case PASS。
- fixture parity test：`branchResumeVersion.json` 六 scenario 与 hook 行为字节匹配。

### Phase 3: Rewrites Tab UI + getResumeVersion 投影

#### 3.1 实现 `frontend/src/app/screens/resume-workshop/tabs/ResumeRewritesTab.tsx`
- 源级复刻 UI 真理源 Rewrites Tab：scope banner + 左侧 bullet 列表（每行：section label / status 圆点 / status 文案 / 截断的 rewritten 预览）+ 右侧 diff 详情 Card（顶部 toolbar / Original / Rewritten 或 Manual edit / WHY THIS CHANGE）+ 顶部计数 chip；
- 数据来源：plan 001 `useResumeVersion(versionId)` 返回的 `version.suggestions[]` + plan 001 adapter `mapBulletSuggestionToUi` 投影；本 plan 扩展 adapter 含 `status / decidedAt / source('ai' | 'manual')`；
- 选中 bullet：维护本地 `selectedBulletId` state；切换时取消任何 inline edit；
- "重新运行改写" button：触发 Phase 5 `useRequestResumeTailor` hook。

#### 3.2 计数与排序
- 计数：`accepted / pending / rejected` 三类；显式从 fixture body 派生（`suggestions.filter(s => s.status === ...)`），不写死；
- 排序：UI 真理源默认按数据顺序展示，本 plan 维持 `suggestions[]` 原始顺序；如需 future 排序由 spec 修订决定。

#### 3.3 隐私
- DOM 渲染原文与 rewritten 文本；URL / pendingAction / localStorage / mock transport log / telemetry 不包含 originalBullet / suggestedBullet 文本；
- 列表行截断展示（90 字符）作为视觉简化，但完整字段仅在 selected 详情 Card 渲染。

#### 3.4 Vitest 单测
- 渲染 happy / empty suggestions / scope banner i18n / 计数派生 / 选中切换 / 隐私 grep。

### Phase 4: 单条 suggestion accept / reject / manual edit 终态

#### 4.1 实现 `tabs/hooks/useAcceptResumeTailorSuggestion.ts`
- 入参：`{ versionId, suggestionId }`；
- 行为：`generateIdempotencyKey()` → generated client `acceptResumeTailorSuggestion(versionId, suggestionId, { idempotencyKey })`；request body 必须为 `undefined`，不得发送 `manualEditText`；
- 成功：返回更新后的 `ResumeVersion`；UI 行 status 变 accepted；
- 失败：
  - 409 `VALIDATION_FAILED` + `detail.reason='SUGGESTION_ALREADY_DECIDED'` → toast `该 bullet 已决定 · 如需重做请新建 suggestion 或 branch`；
  - cross-user 404 → toast generic；
  - 422 → inline；
- IK request spy 断言 header。

#### 4.2 实现 `tabs/hooks/useRejectResumeTailorSuggestion.ts`
- 同 accept hook 形态，调 `rejectResumeTailorSuggestion`；reject 后 UI 行 status 变 rejected 并 line-through。

#### 4.3 inline manual edit
- 在 Rewrites Tab diff Card 中点击 "编辑" → 切换到 textarea + Cancel / Save manual edit 两 button；
- "保存人工改写" 触发：
  - 当前唯一路径：Phase 6 `useUpdateResumeVersion` 先把 manual edit 文本 patch 到 `version.structuredProfile.manualEdits[]`（建议 shape：`{ suggestionId, text, savedAt }`，存放在 OpenAPI `structuredProfile` additionalProperties 内），成功后再调用 bodyless `useAcceptResumeTailorSuggestion({ versionId, suggestionId })` 标记终态；
  - 若 update 成功但 accept 失败：保留 inline manual edit 文本 + saved-manual-pending retry CTA，只重试 bodyless accept，不重复写 manual edit；
- UI 上仍按 UI 真理源行为：保存后 status 设为 accepted（manual 标记），不再 pending；
- Vitest 测试覆盖 update→accept 成功、update 成功但 accept 409 / 404 / network failure 的 retry 状态，以及 update 422 不触发 accept。

#### 4.4 状态机契约
- terminal：accepted / rejected 都是终态；再次点击 accept / reject 走 IK replay；不同 fingerprint 同 key 409；
- accept 不自动 patch `version.structured_profile`（D-12 同步）；UI 不显示 "已自动写入 v1" 这类误导文案；
- 改 suggestion 状态不影响 sibling versions / master / original。

#### 4.5 Vitest 单测
- accept happy / replay / 409 already-decided / 422；
- reject happy / replay / 409；
- manual edit save → `updateResumeVersion` explicit patch + bodyless accept 两步路径；
- accept 不自动改 structured_profile DOM 断言；
- IK header on 三类 op；至少 ≥ 12 case PASS。

### Phase 5: requestResumeTailor + tailor run polling

#### 5.1 实现 `tabs/hooks/useRequestResumeTailor.ts`
- 入参：`{ resumeAssetId, targetJobId, mode: 'gap_review' | 'bullet_suggestions' }`；
- 行为：`generateIdempotencyKey()` → generated client `requestResumeTailor(payload, { idempotencyKey })`；返回 `{ tailorRunId, job(jobType=resume_tailor, status=queued) }`；
- 触发轮询 `useResumeTailorRunPolling(tailorRunId)`。

#### 5.2 实现 `tabs/hooks/useResumeTailorRunPolling.ts`
- 入参：`tailorRunId`；
- 行为：指数退避轮询 `getResumeTailorRun(tailorRunId)`（初始 1500ms / backoff 1.4x / max 12 attempt / ~60s 上限）；
- 退出：`status='ready'` → 触发 `getResumeVersion(versionId)` 刷新 suggestions[]；`status='failed'` → 渲染失败 banner + 重试 CTA（再次调 requestResumeTailor）；timeout → 同 failed；
- fixture-backed harness：当前 fixture 已包含 `queued / generating / default(ready) / failed`；hook test 通过 deterministic scenario sequence 验证 `queued → generating → ready` / `failed`，不再把 status variant 缺失作为 backend follow-up。

#### 5.3 Rewrites Tab UI 集成
- 启动 ai_select branch 后 Rewrites Tab 渲染 polling 进度 banner（"AI 正在生成 bullet · 估计 30s 内完成"）；
- 用户手动触发 "重新运行改写" 后同样进入 polling 状态；
- ready 后 banner 消失 + 列表刷新；failed 后 banner 红 + 重试 CTA。

#### 5.4 Vitest 单测
- happy: queued → generating → ready；
- failed: queued → generating → failed；
- timeout: queued → generating × 12 → timeout；
- requestResumeTailor IK header；
- getResumeTailorRun 不携带 IK header；
- 至少 ≥ 8 case PASS。

### Phase 6: Edit Tab + updateResumeVersion 保存

#### 6.1 实现 `frontend/src/app/screens/resume-workshop/tabs/ResumeEditTab.tsx`
- 源级复刻 UI 真理源 ResumeEditTab：top banner（master 与 targeted 不同提示）+ "保存改动" 按钮 + Headline input + Summary textarea + Experience section（含 Add button 新增 item placeholder；P0 add 行为可仅作为 toast "敬请期待"，不触发真实新增）+ Skills section；
- 数据来源：plan 001 `useResumeVersion(versionId)` 返回 `version.structuredProfile`（含 headline / summary / experience[] / skills[]）+ adapter `mapStructuredProfileToEditTabFields`；
- 本地 state：`{ headline, summary }`（P0 实际 editable 字段）+ 可选 Experience / Skills 编辑（P0 仅 placeholder 渲染，allow follow-up plan）；本 plan 至少打通 headline + summary 实际 patch。

#### 6.2 实现 `tabs/hooks/useUpdateResumeVersion.ts`
- 入参：`{ versionId, displayName?, structuredProfile?, focusAngle?, matchScore? }`；
- mapper：过滤不可编辑字段（`versionType / resumeAssetId / parentVersionId / targetJobId / seedStrategy`）；如调用方传入则 throw + lint error；
- 行为：`generateIdempotencyKey()` → generated client `updateResumeVersion(versionId, payload, { idempotencyKey })`；
- 成功：返回更新后的 `ResumeVersion` → 触发 `getResumeVersion(versionId)` refetch + toast `已保存改动`；
- 失败：422 inline；409 IK conflict toast；cross-user 404 toast generic；
- IK request spy 断言 header。

#### 6.3 Vitest 单测
- happy headline + summary 提交；
- 不可编辑字段过滤；
- IK header + replay；
- 422 / 409 错误映射；
- master vs targeted scope banner 文案；
- 至少 ≥ 8 case PASS。

### Phase 7: i18n + a11y + 隐私 + UI parity + BDD + 旧入口负向

#### 7.1 i18n
- 新增 key 空间：
  - `resumeWorkshop.branch.*`（标签 / scope / form labels / cta / errors）；
  - `resumeWorkshop.rewrites.*`（scope banner / status enum / toolbar / why）；
  - `resumeWorkshop.edit.*`（top banner master vs targeted / labels / save button）；
  - `resumeWorkshop.tailor.*`（polling banner / retry / mode labels）；
- EN / ZH 切换关键文案 + Accept-Language header 携带 7 个 op 请求（branchResumeVersion / requestResumeTailor / getResumeTailorRun / acceptResumeTailorSuggestion / rejectResumeTailorSuggestion / updateResumeVersion / exportResumeVersion）。

#### 7.2 a11y
- BranchFlow：表单 field set + `aria-label` / `aria-required` / `aria-invalid`；focus 落在第一个 input（version name）；
- Rewrites Tab：bullet 列表 `role="listbox"` + 单行 `role="option"` + `aria-selected`；diff Card 内 buttons 标签 `aria-label="拒绝/编辑/采纳"`；scope banner `role="status"` + `aria-live="polite"`；
- Edit Tab：input / textarea 标签关联 + save 按钮 `aria-disabled` 状态；
- 键盘：tab 切换、enter 触发主按钮、ESC 取消 inline edit；
- Playwright axe-core check PASS。

#### 7.3 隐私红线
- originalBullet / suggestedBullet / matchSummary.strengths / matchSummary.gaps / structuredProfile / manualEditText / form draft 不出现在：console / URL query / route params / pendingAction params / localStorage / sessionStorage / mock transport log / telemetry / error toast 内容；
- pendingAction 在所有 auth gate 中仅保留 route + 必要 params（`flow / branchOriginalId / versionId / tab`），不携带 form draft / suggestion 内容。

#### 7.4 UI parity gate
- 新增 `frontend/tests/pixel-parity/resume-workshop-branch-rewrites-edit.spec.ts` 覆盖三屏 desktop + mobile；
- 关键元素 bounding box：BranchFlow form grid / Rewrites Tab 左右栏 / Edit Tab card 间距；
- screenshot 只作为 smoke / 稳定 baseline gate；clean checkout PASS 不依赖未跟踪 baseline。

#### 7.5 Export PDF / copyText 一致性
- 在 Rewrites Tab + Edit Tab 顶部 header 中保留 plan 001 的 Export PDF + 复制纯文本按钮；切换 tab 不影响行为；
- 测试断言 plan 001 P0.037 中关于 Export PDF P0 stub toast / Idempotency-Key header / copyText clipboard 真实可用的 evidence 在本 plan 三屏上仍然成立。

#### 7.6 BDD 场景验证
- `test/scenarios/e2e/p0-084-resume-branch-flow-three-seed-strategies/` 全 PASS（covers C-11 branch 主路径 + 3 seed + IK + 422 + cross-user）。
- `test/scenarios/e2e/p0-085-resume-rewrites-tab-tailor-run-polling/` 全 PASS（covers C-11 Rewrites Tab 主路径 + ai_select branch dispatch handoff + 重新运行改写 + tailor polling）。
- `test/scenarios/e2e/p0-086-resume-suggestion-accept-reject-edit-and-update-version/` 全 PASS（covers C-11 终态状态机 + manual edit fallback + Edit Tab updateResumeVersion 主路径 + IK）。
- `test/scenarios/e2e/p0-087-resume-detail-export-copy-consistency-and-parity/` 全 PASS（covers C-11 三屏 UI parity + plan 001 Export / copyText 不退化 + 隐私 + 旧入口负向）。
- 在 `test/scenarios/e2e/INDEX.md` 追加 P0.084 + P0.085 + P0.086 + P0.087 行。

#### 7.7 旧入口 / retired tailor mode 负向 grep
- `git grep -nE "welcome|mistake|growth|drill|followup|STAR|ExperiencesScreen|experiences-route|voice|OnboardingScreen|onboarding=true" -- frontend/src/app/screens/resume-workshop/branch/ frontend/src/app/screens/resume-workshop/tabs/`：0 命中（不禁止 D-20 flat profile `experiences[]` 字段）；
- `git grep -nE "(^|[^A-Za-z0-9_])(inline|rewrite|mirror)([^A-Za-z0-9_]|$)" -- frontend/src/app/screens/resume-workshop/branch/ frontend/src/app/screens/resume-workshop/tabs/`：0 命中（B3 D-14 同步）；
- `git grep -nE "ui-design/src/(data|screen-resume-workshop)" -- frontend/src/app/screens/resume-workshop/branch/ frontend/src/app/screens/resume-workshop/tabs/`：0 命中。

#### 7.8 spec / history / INDEX 同步
- 确认 frontend-resume-workshop spec.md / history.md / `docs/spec/INDEX.md` 已由本 L1 设计结晶同步到 1.1，并且 §6 C-11 / §7 plan 003 行指向当前 active plan；实施阶段不得为了 checklist 收口重复 bump spec 版本，除非发现新的设计事实需要原地修订。
- 确认 §3.2 accept/reject 口径为 UI 真理源 inline action + terminal-state feedback，不引入未在 `ui-design/` 出现的独立 ConfirmDialog。
- 确认 `docs/spec/frontend-resume-workshop/plans/INDEX.md` 已包含 003 active 行，且 Header / INDEX 投影一致。
- `sync-doc-index --check` PASS。

### Phase 8: D-20 简历扁平化 collapse（删 branch + rewrites 采纳保存）

> product-scope D-20 / spec D-8。依赖 B2 004 Phase 7（contract collapse）+ generated client 重生。删除 `flow=branch` / `ResumeBranchFlow` / seedStrategy；`ResumeRewritesTab` 改写仅「采纳」（删除逐条拒绝/编辑 + `acceptResumeTailorSuggestion`/`rejectResumeTailorSuggestion` op）+ tailor run polling（`requestResumeTailor`→`getResumeTailorRun` ephemeral）+ `RewriteSaveConfirmModal` 覆盖（`updateResume`）/ 另存（`duplicateResume`）；`ResumeEditTab` 提交 `updateResume`。详见 spec D-8。

#### 8.1 Flat route / retired branch gate

- `parseResumeWorkshopParams` / `ResumeWorkshopScreen` 只 materialize `flow=list | create` + `resumeId` + `tab` + optional `tailorRunId` / `targetJobId`。
- `flow=branch` 或未知 flow 回落到 flat list；runtime 不渲染 `resume-branch-flow`。
- `ResumeBranchFlow` / `branchResumeVersion` / `seedStrategy` / `resumeVersionId` / `resumeAssetId` / `listResumeVersions` / `getResumeVersion` 在 runtime Resume Workshop source 中保持 0 命中（test 文件和历史文档除外）。

（验证：`ResumeWorkshopScreen.test.tsx` + P0.084 verify retired grep）

#### 8.2 Rewrites accept-only + save modal gate

- `ResumeRewritesTab` 只提供本地「采纳」；不出现 reject / inline edit / server-side suggestion terminal decision。
- `RewriteSaveConfirmModal` 覆盖（`updateResume`）与另存（`duplicateResume`）两条保存路径均可用，保存中 / 取消 / modal item / mode toggle 有 DOM anchor。
- 保存 request 带 `Idempotency-Key`，不发送 `acceptResumeTailorSuggestion` / `rejectResumeTailorSuggestion` / `updateResumeVersion`。

（验证：`ResumeRewritesTab.test.tsx` + `ResumeDetailView.test.tsx` + P0.086 verify）

#### 8.3 Flat structuredProfile merge gate

- accepted rewrite merge 覆盖 `structuredProfile.sections[]`、`experience[]`、`experiences[]`、`projects[]` 的 `bullets: string[]`；未匹配 bullet 保持不变。
- 保存 payload 不写入 UI-only `acceptedRewrites` / modal state。
- flat `getResume` response 省略 `structuredProfile` 且没有 source text 时，原件 preview fallback 不崩溃。

（验证：`ResumeDetailView.test.tsx` 三条 regression：sections overwrite、flat bullet arrays、omitted profile fallback）

#### 8.4 Tailor rerun route context gate

- route 解析出的 optional `targetJobId` 必须从 `ResumeWorkshopScreen` 透传到 detail container 与 Rewrites Tab rerun hook。
- 有 `targetJobId` 时 `requestResumeTailor` body 为 `{ resumeId, targetJobId, mode }`；无 `targetJobId` 时才允许 `{ resumeId, mode }` generic rerun。
- rerun body 不恢复 `resumeAssetId` / `resumeVersionId`。

（验证：`ResumeDetailView.test.tsx` route-context regression + `useRequestResumeTailor.test.tsx` IK/body contract）

#### 8.5 Edit Tab + export/copy non-regression gate

- `ResumeEditTab` 保存通过 `updateResume` 覆盖 flat resume `displayName` / `structuredProfile.headline` / `structuredProfile.summary`；422 / 409 / 404 error mapping 仍可见。
- Export PDF 使用 `exportResume` P0 501 friendly toast；copyText 使用 `buildResumePlainText` clipboard path；Rewrites / Edit tab 切换不隐藏 header export。

（验证：`ResumeEditTab.test.tsx` + `ResumeDetailExport.test.tsx` + P0.087 verify）

#### 8.6 UI parity / privacy / BDD wrapper gate

- P0.084-P0.087 trigger/verify 使用当前 flat test files，verify 显式拒绝 no-test / fail marker，并检查 real-backend generated-client preflight marker。
- P0.087 Playwright parity 跑 `frontend/tests/pixel-parity/resume-workshop-branch-rewrites-edit.spec.ts`，证明 flat detail/Rewrites/Edit desktop + mobile DOM / style / bounding / screenshot smoke / axe。
- 隐私 gate 覆盖 originalBullet / suggestedBullet / matchSummary / structuredProfile / manual edit 文本不进入 URL / pendingAction / localStorage / mock transport log / toast。

（验证：P0.084-P0.087 `setup → trigger → verify → cleanup` + focused/full frontend gates）

## 5 验收标准

- 本计划列出的 §4 所有 Phase task 全部完成；
- §3 替代验证 gate 全部通过；
- spec §6 C-11 PASS（D-20 flat Rewrites Tab + Edit Tab + `RewriteSaveConfirmModal` + exportPDF/copyText 一致性）；C-1..C-10 不退化；
- BDD E2E.P0.084 + P0.085 + P0.086 + P0.087 按当前 flat scenario contract PASS；
- BUG-0123 类 gate 已固化：omitted `structuredProfile` fallback、accepted rewrites 写入 flat bullet containers、route `targetJobId` 进入 rerun body；
- UI parity gate 已接入 `frontend/tests/pixel-parity/resume-workshop-branch-rewrites-edit.spec.ts`，clean checkout PASS 不依赖本地未跟踪 screenshot baseline；
- engineering-roadmap §5.2 `frontend-resume-workshop` 状态保持 active；
- spec.md 1.3 / history.md / plans/INDEX.md / docs/spec/INDEX.md 同步至最新；除非本 plan 实施中引入新的设计事实，否则不重复 bump spec 版本；
- 上游 gate 已满足：D-20 flat `listResumes / getResume / updateResume / duplicateResume / exportResume / requestResumeTailor / getResumeTailorRun` generated client + fixtures + real handler 可用；退役 version-tree operations 不回流。

## 6 风险与应对

| 风险 | 应对 |
|------|------|
| R1: backend / OpenAPI fixture / generated client 回到 version-tree shape，导致 frontend flat hooks 与真实后端不一致 | Phase 8.1 / 8.6 通过 codegen-check、real-mode gate 和 runtime retired grep 阻断；缺失则停止 Phase 8 并回到对应 owner 修复 |
| R2: accepted rewrite 保存只覆盖 legacy `sections[]`，漏掉 D-20 flat `experience[]` / `projects[]` | Phase 8.3 固化 flat profile merge regression，必须覆盖 `sections`、`experience`、`experiences`、`projects` |
| R3: `getResume` response 省略 `structuredProfile` 时 detail fallback 崩溃 | Phase 8.3 固化 omitted profile fallback regression；fixture 或 component test 任一失败都不能收口 |
| R4: `targetJobId` 只停留在 root route data attribute，rerun body 丢失 JD context | Phase 8.4 要求 body-level assertion；有 `targetJobId` 时 `requestResumeTailor` 必须携带，且不恢复 old ids |
| R5: IK header fixture 与 generated client `opts.idempotencyKey` 行为漂移 | Phase 8.2 / 8.5 用 hook/spy tests 断言 `requestResumeTailor`、`updateResume`、`duplicateResume`、`exportResume` 的 IK 行为 |
| R6: `getResumeTailorRun` status sequence 只在单 fixture scenario 中表达，polling 测试若继续使用 synthetic mock 可能绕过真实 schema | Phase 8.4 / P0.085 使用 `queued / generating / default(ready) / failed` fixture variants 组成 deterministic sequence；只允许 mock 调度顺序，不 mock response schema |
| R7: Edit Tab P0 仅打通 headline + summary，Experience / Skills 列表的 add / edit / remove 未实现可能让用户期待落空 | Phase 8.5 显式锁定 P0 保存字段；如 UI 真理源扩展 Experience / Skills 编辑，先更新 `ui-design/` 与 spec，再新增 plan |
| R8: flat detail + Rewrites + Edit pixel parity baseline 数量增加，可能引起本地 gate 时间膨胀 | Phase 8.6 复用 frontend-shell/003 pipeline；screenshot 仅 smoke + DOM/style/bounding box 优先；新机器先跑 `test:pixel-parity:install` 缓存浏览器 |
| R9: 用户在 tailor polling 后切到 Edit Tab 或返回 list，polling hook 资源泄漏 | `useResumeTailorRunPolling` hook cleanup + P0.085 unmount fake-timer gate 证明无后续 `getResumeTailorRun` |
| R10: i18n key namespace 与 plan 002 冲突 | 继续使用 `resumeWorkshop.rewrites.*` / `.edit.*` / `.tailor.*`，与 plan 002 `.create.*` / `.parsing.*` / `.preview.*` 不交叉；类型检查捕获 key drift |
