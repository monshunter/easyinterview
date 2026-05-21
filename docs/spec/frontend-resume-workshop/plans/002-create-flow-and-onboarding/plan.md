# Frontend Resume Workshop Create Flow and Onboarding

> **版本**: 1.0
> **状态**: completed
> **更新日期**: 2026-05-21

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [frontend-resume-workshop spec](../../spec.md) §6 C-10（CreateFlow 三 tab + Onboarding）以及 §3.2 待确认事项中已锁定的 guided 模式 P0 落地、首页 / Workspace `1 分钟创建` 串通项落到 `frontend/` 实现：

- 替换 [001-listing-routing-and-detail-readonly](../001-listing-routing-and-detail-readonly/plan.md) 阶段在 `ResumeWorkshopScreen` 中对 `flow=create` 渲染的 `<NotImplementedPlaceholder>`，源级复刻 [`ui-design/src/screen-resume-workshop.jsx`](../../../../../ui-design/src/screen-resume-workshop.jsx) 中以下组件：
  - `ResumeCreateFlow`（路由容器 + `stage ∈ {input | parsing | preview}` 状态机 + `createMode ∈ {upload | paste | guided}` 切换）
  - `ResumeParseFlow`（Agent Parsing 进度态：7 段步骤动画 + 取消返回）
  - `ResumePreviewConfirm`（结构化草稿预览 + 右栏 "会保存什么" / "解析备注" + 返回 / 确认保存 v1）
- 实现 Upload tab 双步上传契约：第一步 generated client `createUploadPresign({purpose: 'resume', fileName, contentType, byteSize})` + Idempotency-Key → 第二步浏览器 `fetch(uploadUrl, { method, headers, body: file })` 直传对象存储；上传完成后 generated client `registerResume({ sourceType: 'upload', fileObjectId, title, language })` + IK，进入 Agent Parsing；
- 实现 Paste tab 单步契约：textarea 非空时 `registerResume({ sourceType: 'paste', rawText, title, language })` + IK；
- 实现 Guided tab 5 步问答 + 提交契约：`registerResume({ sourceType: 'guided', guidedAnswers, title, language })` + IK；guidedAnswers 以 jsonb 对象提交（按 UI 真理源 5 个 key：`recentRole / direction / proofProject / metrics / target`），与 `RegisterResumeRequest.guidedAnswers` schema 字节兼容；
- Agent Parsing 阶段消费 `getResume(resumeAssetId)` 进行 `parseStatus` 轮询（队列态：`queued | generating` 触发动画继续；终态 `ready` → 跳 Preview Confirm；终态 `failed` → 失败 UI + 重试 / 返回输入；与 [backend-resume/001 C-3 / C-4](../../../backend-resume/spec.md) 行为对齐；frontend 不直接消费 `resume.parse.completed` outbox event）；
- Preview Confirm 提交契约：`confirmResumeStructuredMaster(resumeAssetId, { structuredProfile, displayName, language })` + IK → 201 `ResumeVersion(versionType='structured_master')` → 返回 list 并 toast；409 `RESUME_STRUCTURED_MASTER_ALREADY_EXISTS` → 显式提示 "已存在主版本"、不重复提交；422 `VALIDATION_FAILED` → inline 错误 + 保留草稿；
- 串通入口：[`ui-design/src/screen-home.jsx`](../../../../../ui-design/src/screen-home.jsx) `还没有简历？1 分钟创建` CTA + [`ui-design/src/screen-workspace.jsx`](../../../../../ui-design/src/screen-workspace.jsx) `WorkspaceMissingResumeState` CTA 均按 `nav("resume_versions", { flow: "create" })` 路由 → 002 落地后实际渲染 `ResumeCreateFlow`；新增集成测试覆盖两条 CTA → 路由 → 容器主路径，确保 placeholder → real component 的转场不破坏既有 Home / Workspace 行为；
- i18n（resumeWorkshop.create.* / resumeWorkshop.parsing.* / resumeWorkshop.preview.* key 空间）+ a11y（tab 切换 / focus 管理 / ESC 取消 / aria-label）+ 隐私红线（rawText / guidedAnswers / parsedSummary / file binary / structuredProfile 内容不出现在 console / URL / pendingAction / localStorage / telemetry / mock transport log）+ UI parity gate（Vitest + Playwright pixel parity 对 `ResumeCreateFlow` / `ResumeParseFlow` / `ResumePreviewConfirm` 三屏 desktop + mobile 断言）；
- 不实现 Branch Flow / Rewrites Tab / Edit Tab（归 [003-branch-rewrites-and-edit](../003-branch-rewrites-and-edit/plan.md)）；不实现 exportResumeVersion 真实 PDF 行为（仍按 plan 001 P0 toast 兜底）；不实现 backend handler / 异步 job / AI 解析（归 [backend-upload/001](../../../backend-upload/plans/001-file-objects-and-presign-baseline/plan.md) + [backend-resume/001](../../../backend-resume/plans/001-asset-register-parse-and-listing/plan.md) + [backend-resume/002](../../../backend-resume/plans/002-versions-tailor-runs-and-save-v1/plan.md)）。

## 2 背景

本 plan 是 frontend-resume-workshop 第二批 plan，承担 P0 用户路径 "首次（或追加）新建简历 → 选择 upload / paste / guided 输入 → 等待 Agent 解析 → 预览确认 → 保存为 v1 主版本 → 回到列表" 的前端端到端。这是 Resume Workshop 阶段 2 的关键解锁项：

- 解锁首页 `还没有简历？1 分钟创建` 与 Workspace `WorkspaceMissingResumeState` 的实际可用路径；plan 001 阶段两条 CTA 仍命中 `<NotImplementedPlaceholder>`，本 plan 完成后用户可在 ≤ 1 分钟范围内从入口走通 `输入 → 解析 → 预览 → 保存 v1`。
- 与 [backend-upload/001](../../../backend-upload/plans/001-file-objects-and-presign-baseline/plan.md) 的 `createUploadPresign` + register state machine 共同闭合 upload 模式；mock-first 阶段可消费 `openapi/fixtures/Uploads/createUploadPresign.json` `default` + 计划新增 scenario，real backend 已 ready（参见近期 work journal `feat(backend-upload): land object store register service`），切真不需要重构 adapter。
- 与 [backend-resume/001](../../../backend-resume/plans/001-asset-register-parse-and-listing/plan.md) 已实现的 `registerResume` / `getResume` / `resume.parse` async job 行为字节对齐；frontend 通过 `getResume` 轮询 `parseStatus` 完成 Agent Parsing 终态判断（与 `openapi/fixtures/Resumes/getResume.json` `default` ready scenario 配合，并按 §3.1 operation matrix 中 `getResume.parse-status-*` 列出的 scenario 升级路径补齐）。
- 与 [backend-resume/002](../../../backend-resume/plans/002-versions-tailor-runs-and-save-v1/plan.md) Phase 1 同步交付的 D-10 additive `confirmResumeStructuredMaster` operation 协作：本 plan Preview Confirm 提交路径必须等待 backend-resume/002 Phase 1 落地（fixture + generated client method）才能开启 mock-first；本 plan Phase 0 显式声明等待 gate，不私造客户端协议。
- 与 [frontend-workspace-and-practice/001 ResumePickerModal](../../../frontend-workspace-and-practice/plans/001-workspace-and-interview-context/plan.md) / [frontend-home-job-picks-and-parse/001](../../../frontend-home-job-picks-and-parse/plans/001-home-jd-import-and-parse/plan.md) 解锁路径互补：Home / Workspace 两条入口的 route handoff 行为不变（plan 001 时已为 `flow=create` 预留 placeholder），本 plan 只换内部渲染。

每个 phase 是可独立验证的纵向切片：Phase 1 起来就有 `ResumeCreateFlow` 容器渲染 + auth gate；Phase 2 起来就有 Upload tab 双步上传链路；Phase 3 起来就有 Paste / Guided tab 单步提交链路；Phase 4 起来就有 ResumeParseFlow 动画 + `getResume` 轮询；Phase 5 起来就有 ResumePreviewConfirm + `confirmResumeStructuredMaster` 保存；Phase 6 起来就有 Home / Workspace CTA 串通 + i18n + a11y + 隐私 + UI parity gate + BDD + 旧入口负向 grep。

执行本 plan 前必须确认：

- [frontend-resume-workshop/001-listing-routing-and-detail-readonly](../001-listing-routing-and-detail-readonly/plan.md) completed（路由替换 + ResumeWorkshopScreen 容器 + flow=create `<NotImplementedPlaceholder>` 行为基线已就位；adapter 层 `frontend/src/app/screens/resume-workshop/adapters/` 已 exist）。
- [openapi-v1-contract/004-resume-additive-coverage](../../../openapi-v1-contract/plans/004-resume-additive-coverage/plan.md) Phase 1-5 已完成（registerResume / getResume / createUploadPresign generated client artifacts 全部就位）。
- [backend-resume/002-versions-tailor-runs-and-save-v1](../../../backend-resume/plans/002-versions-tailor-runs-and-save-v1/plan.md) Phase 1 已完成（D-10 `confirmResumeStructuredMaster` operation + 4 scenario fixture + B1 错误码 `RESUME_STRUCTURED_MASTER_ALREADY_EXISTS` + generated client method 已 ready）；如未 ready，本 plan Phase 5 暂以 `<ComingSoonPreviewConfirm>` 占位，不私造协议。Phase 6 串通必须基于 confirm 真实可用后再 close。
- [backend-upload/001-file-objects-and-presign-baseline](../../../backend-upload/plans/001-file-objects-and-presign-baseline/plan.md) Phase 0-5 已落地（createUploadPresign handler + state machine + fixture）；mock-first 阶段允许仅依赖 fixture，切真不阻塞本 plan。
- UI 真理源 [`ui-design/src/screen-resume-workshop.jsx`](../../../../../ui-design/src/screen-resume-workshop.jsx)（`ResumeCreateFlow` / `ResumeParseFlow` / `ResumePreviewConfirm` 三组件）+ [`docs/ui-design/resume-onboarding.md`](../../../../ui-design/resume-onboarding.md) v1.5 + [`docs/ui-design/resume-module.md`](../../../../ui-design/resume-module.md) v1.7 active。
- [frontend-shell auth pending action contract](../../../frontend-shell/spec.md) 已就位；CreateFlow auth gate 通过 pendingAction 保留 `flow=create` + `createMode`，不携带 raw text / guidedAnswers。

## 3 质量门禁分类

- **Plan 类型**: `code-internal + feature-behavior`。本 plan 实现前端容器 + 三种输入模式提交 + 轮询 + 确认保存 v1；用户可见 UI 行为，且与多个后端 owner 跨契约。
- **TDD 策略**: 适用。Red-Green-Refactor 入口：
  1. Vitest 组件单测：`ResumeCreateFlow` stage / mode 状态机 / param 解析 / auth gate；`ResumeParseFlow` 动画 + 轮询取消；`ResumePreviewConfirm` 草稿渲染 + 提交错误处理；新增 hooks（`useResumeRegistration` / `useResumeParsingPolling` / `useResumeStructuredMasterConfirm`）；
  2. adapter unit test：Upload `RegisterResumeRequest{sourceType:'upload'}` / Paste `{sourceType:'paste', rawText}` / Guided `{sourceType:'guided', guidedAnswers}` 三种 payload 形态；Preview Confirm `ConfirmResumeStructuredMasterRequest` payload；
  3. fixture parity test：组件渲染从 `Uploads/createUploadPresign.json default` + `Resumes/registerResume.json default / paste-text / guided-answers` + `Resumes/getResume.json default / master-default / not-found` + `Resumes/confirmResumeStructuredMaster.json default / idempotency-replay / already-exists-409 / validation-422` 时 DOM testid 覆盖锚点 + 数量断言从 fixture 派生；
  4. Idempotency-Key contract test：`createUploadPresign` / `registerResume` / `confirmResumeStructuredMaster` 三个 op 通过 `frontend/src/lib/conventions/idempotency.ts::generateIdempotencyKey()` 生成 IK，request spy 断言 `Idempotency-Key` header 出现；replay 行为通过 fixture 验证；
  5. 隐私 grep test：raw text / file binary / guidedAnswers / parsedSummary / structured_profile 内容不出现在 URL / pendingAction params / localStorage / mock transport log / console output；
  6. auth boundary test：未登录访问 `resume_versions?flow=create` 不触发 createUploadPresign / registerResume / getResume / confirmResumeStructuredMaster；登录恢复只携带 `flow=create` + 可选 `createMode`，不携带文本；
  7. Playwright pixel parity：`ResumeCreateFlow`（3 mode 各自） + `ResumeParseFlow` + `ResumePreviewConfirm` 五个屏幕 desktop 1440px + mobile 390x844 DOM anchor + computed style + bounding box + screenshot smoke（仅在 baseline 可复现/维护时使用 screenshot diff）；
  8. negative grep test：`frontend/src/app/screens/resume-workshop/create/` 不出现 retired 模块名（welcome / mistake / growth / plan / drill / followup / 旧 STAR / 旧 experiences / voice）；不出现旧 onboarding 字面量 `OnboardingScreen` / `onboarding=true`；不 import `ui-design/src/screen-resume-workshop.jsx` / `ui-design/src/data.jsx` 作为运行时依赖。

  执行入口：`/implement frontend-resume-workshop/002-create-flow-and-onboarding` → `/tdd`。

- **BDD 策略**: 适用（Feature plan requires BDD）。`E2E.P0.081` create-flow-upload-paste-guided-happy + `E2E.P0.082` create-flow-parsing-failure-and-retry + `E2E.P0.083` create-flow-preview-confirm-and-cta-handoff，详见 [bdd-plan.md](./bdd-plan.md) / [bdd-checklist.md](./bdd-checklist.md)。

- **替代验证 gate**:
  - `pnpm --filter @easyinterview/frontend test` (Vitest)
  - `pnpm --filter @easyinterview/frontend build` + `pnpm --filter @easyinterview/frontend test:pixel-parity` (Playwright；首次或新机器先跑 `pnpm --filter @easyinterview/frontend test:pixel-parity:install`)
  - `pnpm --filter @easyinterview/frontend lint` (ESLint + UI parity rules)
  - `pnpm --filter @easyinterview/frontend build`
  - `git grep -nE "welcome|mistake|growth|drill|followup|STAR|experiences|voice|OnboardingScreen" -- frontend/src/app/screens/resume-workshop/create/`（旧入口 negative；当前 plan 文档 prose 不纳入 raw zero-hit）
  - `git grep -nE "ui-design/src/(data|screen-resume-workshop)" -- frontend/src/app/screens/resume-workshop/create/`（原型 runtime import negative）
  - `git grep -nE "rawText|guidedAnswers|parsedSummary|parsedTextSnapshot|structuredProfile" -- frontend/src/app/screens/resume-workshop/create/ | git grep -vE "(\\\.test\\\.|adapters/|hooks/use|api/)"` 用于辅助检视未被允许直接渲染或对外输出敏感内容；具体 grep 表达式由实施时按 §3 隐私红线条款落定，不强制 raw 0 命中（隐私 grep 主断言在 Vitest / Playwright 测试内执行）
  - `sync-doc-index --check`

### 3.1 Frontend / Backend Operation Matrix

本 plan 走 mock-first 与切真混合 frontend path：上传 / 注册 / 解析轮询 / 确认保存 v1 共四类操作。fixture 覆盖必须显式标注；any operation that 在本 plan kickoff 时缺少 default / 边界 scenario 必须在 §6 风险与 Phase 0 dependency gate 中显式声明。

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `createUploadPresign` | `openapi/fixtures/Uploads/createUploadPresign.json` `default`；本 plan 不新增 scenario，validation/IK replay 行为通过 Vitest 模拟 generated client error path 验证 | `useResumePresignUpload` hook（Phase 2）：传入 `purpose='resume'` + file metadata + `generateIdempotencyKey()` → 拿到 `uploadUrl / method / headers / expiresAt / fileObjectId` | `backend-upload/001` already-landed (per work journal `feat(backend-upload): land object store register service`); fixture-backed allowed in mock-first dev | `file_objects(upload_status='pending')` | none | E2E.P0.081 |
| (Browser PUT to signed URL) | n/a（fixture 模拟成功）；mock transport 不真实上传，Vitest spy 校验 `fetch(uploadUrl, { method, headers, body: file })` 调用形态 | `useResumePresignUpload.uploadBinary` 阶段；上传失败映射 toast "上传失败 · 请重试" | n/a (object store provider) | object store（uploaded object） | none | E2E.P0.081 |
| `registerResume` | `openapi/fixtures/Resumes/registerResume.json` `default`（upload）/ `paste-text` / `guided-answers`；本 plan 不在 fixture 层新增 validation 场景，验证由 generated client mapper 单测 + handler 422 path test 覆盖 | `useResumeRegistration` hook：根据 `createMode` 选择对应 payload 形态（`sourceType: 'upload' | 'paste' | 'guided'`）+ IK；返回 `ResumeAssetWithJob{resumeAssetId, job(jobType=resume_parse, status=queued)}` | `backend-resume/001` already-landed (per `feat(backend-resume): close resume baseline verification`); fixture-backed allowed in dev mock | `resume_assets`（含 sourceType / fileObjectId / rawText / guidedAnswers / parseStatus='queued'） | `resume.parse` async (downstream — not invoked from frontend) | E2E.P0.081 |
| `getResume` | `openapi/fixtures/Resumes/getResume.json` `default` (parseStatus='ready') / `master-default` / `not-found`；本 plan **不依赖** 缺失的 `queued` / `generating` / `failed` parseStatus scenario：轮询逻辑由 hook 在 mock 层根据 attempt 序列模拟 `queued → generating → ready` 过渡；frontend Phase 4 必须在 verify 中明确标注 "现阶段 fixture 未提供逐 attempt parseStatus scenario，所以使用 mock client deterministic stepping"，并在 retrospective 中提议由 backend-resume/001 followup 或 backend-resume/002 followup plan 在 fixture 层补齐 | `useResumeParsingPolling` hook（Phase 4）：以指数退避（默认 1.2x，最大 8 attempt，~30s 上限）轮询 `getResume(resumeAssetId)`；终态 ready / failed 退出；fixture-backed 测试通过 mock client `__nextAttemptScenario` API（mock-contract-suite 已支持 attempt-aware 替代时切到原生支持，否则保留 deterministic stepping） | `backend-resume/001` already-landed; fixture-backed allowed | `resume_assets` (read) | downstream `resume.parse` (frontend not the consumer) | E2E.P0.082 |
| `confirmResumeStructuredMaster` (D-10 NEW; cross-owner change from backend-resume/002 Phase 1) | `openapi/fixtures/Resumes/confirmResumeStructuredMaster.json` `default` / `idempotency-replay` / `already-exists-409` / `validation-422`（**必须等待 [backend-resume/002 Phase 1](../../../backend-resume/plans/002-versions-tailor-runs-and-save-v1/plan.md#phase-1-b2-d-18-additive-confirmresumestructuredmaster--b1-错误码增补) 落地；fixture 本 plan 不自带，consumer plan 仅消费**） | `useResumeStructuredMasterConfirm` hook（Phase 5）：传入 resumeAssetId + structuredProfile（来自 getResume 的 parsedSummary 投影，由 adapter 落到 `ConfirmResumeStructuredMasterRequest` schema）+ displayName + language + IK；成功返回 `ResumeVersion(versionType='structured_master')` → toast + nav 回 list；409 `RESUME_STRUCTURED_MASTER_ALREADY_EXISTS` → 显式提示 + nav 到对应已存在 master detail（通过 `listResumeVersions` 查找）；422 `VALIDATION_FAILED` → inline 错误 | `backend-resume/002` Phase 1 in-flight；mock-first 直到 fixture 落地；real handler 由 backend-resume/002 Phase 2 实现 | `resume_versions(version_type='structured_master')` | none in confirm path | E2E.P0.083 |
| `listResumes` | `openapi/fixtures/Resumes/listResumes.json` `default` / `empty` / `paginated`（plan 001 已消费） | 回到 list 后 plan 001 的 `ResumeListView` 自然刷新；本 plan 集成测试断言新建 entry 出现在最近 entry 之首 | `backend-resume/001` already-landed | `resume_assets` (read) | none | E2E.P0.083 |

### 3.2 上游依赖 gate（必须在本 plan 落地前确认）

- 等待 [backend-resume/002 Phase 1](../../../backend-resume/plans/002-versions-tailor-runs-and-save-v1/plan.md#phase-1-b2-d-18-additive-confirmresumestructuredmaster--b1-错误码增补) 同步交付：openapi.yaml 新增 `POST /api/v1/resumes/{resumeAssetId}/structured-master` operation + `ConfirmResumeStructuredMasterRequest` schema + 4 个 fixture scenario + B1 错误码 `RESUME_STRUCTURED_MASTER_ALREADY_EXISTS` + Go/TS generated client artifacts。本 plan Phase 0 必须显式 grep `openapi/fixtures/Resumes/confirmResumeStructuredMaster.json` 与 generated client `confirmResumeStructuredMaster` symbol，缺失则停止 Phase 5 实施并升级 blocker。
- 验证 plan 001 已替换 `PlaceholderScreen` 行为：`resume_versions?flow=create` 在 plan 001 阶段渲染 `<NotImplementedPlaceholder>`；本 plan Phase 1 替换为 `ResumeCreateFlow`。
- 验证 [frontend-home-job-picks-and-parse/001](../../../frontend-home-job-picks-and-parse/plans/001-home-jd-import-and-parse/plan.md) 与 [frontend-workspace-and-practice/001](../../../frontend-workspace-and-practice/plans/001-workspace-and-interview-context/plan.md) CTA route 行为：两个 plan 已为 `resume_versions?flow=create` 留路由；本 plan Phase 6 只校验链路。

## 4 实施步骤

### Phase 1: ResumeCreateFlow 容器 + stage / mode 状态机 + auth gate

#### 1.1 替换 plan 001 中 `flow=create` 的 `<NotImplementedPlaceholder>`
- 修订 `frontend/src/app/screens/resume-workshop/ResumeWorkshopScreen.tsx`：当 `flow === 'create'` 时渲染新增的 `ResumeCreateFlow` 而非 placeholder。
- placeholder 组件保留（仅供 `flow === 'branch'` 在 plan 003 落地前继续使用）。

#### 1.2 实现 `frontend/src/app/screens/resume-workshop/create/ResumeCreateFlow.tsx`
- 状态：`stage ∈ {'input' | 'parsing' | 'preview'}` + `createMode ∈ {'upload' | 'paste' | 'guided'}` + 输入态本地 state（`pickedFile / rawText / guidedAnswers[5]`）+ 来自 hooks 的 `currentAssetId` 与 `parseStatus`。
- 渲染：Header（含 "返回简历工坊" 按钮 + label + h1 + 说明）+ 主区两栏（Card 左 - 三 tab + tab content；右栏 - "会保存什么" + "接下来" 两个 Card），按 UI 真理源 DOM/Style 1:1 复刻；不使用 ui-design 的 `buildResumeData` runtime；i18n key 走 `resumeWorkshop.create.*`。

#### 1.3 auth gate
- 未登录态：渲染 plan 001 的 auth gate / 登录引导；pendingAction 只携带 `{ route: 'resume_versions', params: { flow: 'create', createMode? } }`；不携带 rawText / guidedAnswers / file binary。
- 登录恢复：进入 `flow=create`，按 createMode 默认 `upload`；不预填用户上次输入。

#### 1.4 Vitest 组件单测
- stage / mode 转换：input ↔ parsing ↔ preview 与 back / cancel；mode 三态切换；最少 8 个 case PASS。
- auth gate：未登录态 0 个 protected API 请求 spy；pendingAction params 不含 raw text。

### Phase 2: Upload tab 双步上传 + IK 契约

#### 2.1 实现 `frontend/src/app/screens/resume-workshop/create/UploadTab.tsx`
- 源级复刻 UI 真理源 `createMode === 'upload'` 分支：dropzone 视觉、icon、`<input type=file accept=".pdf,.docx,.md,.txt">` 触发逻辑、`pickedFile` 名称展示。
- 客户端 pre-check：文件扩展名 ∈ {pdf, docx, md, txt}（不通过则 inline error，不进入双步契约）；文件大小不可超 `upload.maxBytes.resume` 默认 10MB（[backend-upload D-7](../../../backend-upload/spec.md#31-已锁定决策)），超限 inline error；frontend 不预设环境覆盖路径。

#### 2.2 实现 `frontend/src/app/screens/resume-workshop/create/hooks/useResumePresignUpload.ts`
- Step 1：generated client `createUploadPresign({ purpose: 'resume', fileName, contentType, byteSize })` + `Idempotency-Key`（每次新 IK；同一 file 重试复用同一 IK，直到成功或 TTL 过期；TTL 过期后重新生成 IK）；
- Step 2：浏览器 `fetch(uploadUrl, { method, headers, body: file, mode: 'cors' })`；失败映射 toast "上传失败 · 请重试"；成功后 hook 返回 `{ fileObjectId, expiresAt }`；
- 错误路径：
  - presign `VALIDATION_FAILED`（purpose / size）：inline error；
  - PUT 失败（network / 4xx / 5xx）：toast + 可重试（同一 file 复用 fileObjectId）；
  - presign URL TTL 过期（`expiresAt < now`）：自动重 presign + IK。

#### 2.3 衔接到 register 阶段
- Upload tab `handleSubmit` 触发：UploadTab 不直接调用 registerResume；由父 ResumeCreateFlow 在拿到 `fileObjectId` 后调用 Phase 3 中 `useResumeRegistration` 的 `register({ sourceType: 'upload', fileObjectId, title, language })`。

#### 2.4 Vitest 单测
- 双步流程 happy + 失败 retry + TTL 过期重 presign + IK header 断言（request spy）；
- 客户端 pre-check 边界（扩展名 / 大小）。

### Phase 3: Paste + Guided tab + `useResumeRegistration` hook

#### 3.1 实现 `frontend/src/app/screens/resume-workshop/create/PasteTab.tsx`
- 源级复刻 UI 真理源 `createMode === 'paste'` 分支：textarea + 下方说明 + accent submit button；submit disabled 当 `rawText.trim().length === 0`；
- submit → 父 ResumeCreateFlow 调 register `{ sourceType: 'paste', rawText, title, language }`；title 由 `frontend/src/app/screens/resume-workshop/create/util/title.ts` 派生（默认 = lang === en ? "Pasted resume" : "粘贴的简历"，长度 ≤ 80）。

#### 3.2 实现 `frontend/src/app/screens/resume-workshop/create/GuidedTab.tsx`
- 源级复刻 UI 真理源 `createMode === 'guided'` 分支：左栏 5 步导航 + 右栏当前 step question + textarea + 上一步 / 下一步 / Generate v1 按钮；按 UI 真理源 5 个 question key（`recentRole / direction / proofProject / metrics / target`）落 i18n；
- 提交：把 `guidedAnswers[0..4]` 映射为 `{ recentRole, direction, proofProject, metrics, target }` jsonb 对象（空字符串原样保留，不修剪用户答案）；触发 register `{ sourceType: 'guided', guidedAnswers, title, language }`。

#### 3.3 实现 `frontend/src/app/screens/resume-workshop/create/hooks/useResumeRegistration.ts`
- 入参：`{ sourceType, fileObjectId? | rawText? | guidedAnswers?, title, language }`；
- 行为：`generateIdempotencyKey()` → generated client `registerResume`；
- 成功：返回 `ResumeAssetWithJob{resumeAssetId, job}`，触发 parent state `stage = 'parsing'`；
- 失败：422 → inline；5xx / network → toast + 保留输入；
- IK request spy 断言 `Idempotency-Key` header 出现且与同一 file / paste / guided 行为内复用至成功或边界重置。

#### 3.4 Vitest 单测
- Paste happy / disabled when empty / IK header；
- Guided 5 step nav + submit on last step + payload shape；
- registerResume mapper：三 sourceType × payload shape 对齐 fixture 字段；
- 422 / 5xx 失败映射。

### Phase 4: Agent Parsing stage（`ResumeParseFlow` + `useResumeParsingPolling`）

#### 4.1 实现 `frontend/src/app/screens/resume-workshop/create/ResumeParseFlow.tsx`
- 源级复刻 UI 真理源 7 段步骤 ticker：每段 ~700ms 推进；最后一段后跳转到 Preview Confirm（实际跳转由 hook polling 触发，UI 动画与轮询解耦：动画作为视觉占位，轮询完成才允许 transition）。
- "取消并返回修改" 按钮：取消 polling + 回到 input stage 并保留用户输入（rawText / guidedAnswers / pickedFile）。

#### 4.2 实现 `frontend/src/app/screens/resume-workshop/create/hooks/useResumeParsingPolling.ts`
- 入参：`resumeAssetId`；
- 行为：指数退避轮询 `getResume(resumeAssetId)`（默认初始 1500ms / backoff 1.4x / 最大 8 attempt / 上限 ~30s）；
- 退出条件：
  - `parseStatus === 'ready'` → 返回 asset payload（含 parsedSummary）；
  - `parseStatus === 'failed'` → 返回 `{ status: 'failed', errorCode }`；UI 显示失败态 + 重试（重启动 polling）/ 回输入；
  - cancel → 终止；
  - 上限 attempt 超出 → 视为 timeout failure，等价 failed 路径但 errorCode='PARSE_TIMEOUT'；
- mock harness：本 plan 不依赖 fixture 提供 `queued` / `generating` / `failed` parseStatus scenario；在 Vitest / Playwright 中通过 mock client 的 attempt-aware override 或 hook 内 stub 模拟终态转移；retrospective 中提议 backend-resume followup 在 fixture 层补齐。

#### 4.3 隐私红线
- parsing 阶段 `getResume` response 中 `originalText / parsedTextSnapshot / guidedAnswers / parsedSummary` 字段不渲染在 ParseFlow UI（仅 ready 后到 PreviewConfirm 才渲染 parsedSummary 结构化字段，且不在 URL / pendingAction / localStorage）；
- URL 不携带 polling 进度；history 不堆叠 stage 转移。

#### 4.4 Vitest 单测
- happy：`queued → generating → ready` → 触发 stage='preview'；
- failed：parseStatus=failed → 失败态 + 重试 button；
- cancel：返回 input stage 且保留输入；
- timeout：8 attempt 后等价 failed；
- IK：getResume 不属于 side-effect op，不要求 IK header；request spy 断言无 IK；
- 隐私：DOM 不渲染 parsedTextSnapshot / parsedSummary 字段。

### Phase 5: ResumePreviewConfirm + `confirmResumeStructuredMaster` 保存 v1

#### 5.1 Gate：等待 [backend-resume/002 Phase 1](../../../backend-resume/plans/002-versions-tailor-runs-and-save-v1/plan.md#phase-1-b2-d-18-additive-confirmresumestructuredmaster--b1-错误码增补) 落地
- 检查 `openapi/fixtures/Resumes/confirmResumeStructuredMaster.json` 存在且含 `default / idempotency-replay / already-exists-409 / validation-422` scenario；
- 检查 generated client 已暴露 `confirmResumeStructuredMaster` method 与 `ConfirmResumeStructuredMasterRequest` / `ResumeVersion` 类型；
- 不满足 gate → Phase 5 推迟，UI 渲染 `<ComingSoonPreviewConfirm>`（含说明 + 返回 input 按钮）；不私造客户端协议；
- 满足 gate → 继续。

#### 5.2 实现 `frontend/src/app/screens/resume-workshop/create/ResumePreviewConfirm.tsx`
- 源级复刻 UI 真理源：header（label "PREVIEW · CONFIRM TO SAVE AS V1" + h1 + source 标签 + `已解析` 状态 + 上一步 / 确认保存按钮）+ 主体两栏（左：草稿主体；右：会保存什么 + 解析备注）；
- 草稿数据来源：`useResumeAsset(resumeAssetId)` 拉取最新 `parseStatus='ready'` 的 resume_asset，通过 adapter `mapParsedSummaryToStructuredProfileDraft` 把 `parsedSummary` jsonb 投影到 `ConfirmResumeStructuredMasterRequest.structuredProfile`；
- 渲染字段集合按 UI 真理源 draft 形态：identity（name/title/location/contact）/ summary / experience / projects / skills / education。
- `displayName` 默认 = `parsedSummary.displayName || mode-derived title`；用户在 Preview Confirm 可不编辑该字段（隐式默认）；如未来支持显式编辑由后续 plan 落地（不阻塞本 plan）。

#### 5.3 实现 `frontend/src/app/screens/resume-workshop/create/hooks/useResumeStructuredMasterConfirm.ts`
- 入参：`{ resumeAssetId, structuredProfile, displayName, language }`；
- 行为：`generateIdempotencyKey()` → generated client `confirmResumeStructuredMaster(resumeAssetId, payload, { idempotencyKey })`；
- 成功：返回 `ResumeVersion`；触发 toast `已保存 v1 主版本 · 进入简历工坊` / `Saved v1 master · back to workshop`；nav `resume_versions` 默认 list；
- 失败处理：
  - `409 RESUME_STRUCTURED_MASTER_ALREADY_EXISTS`：toast `已存在主版本 · 跳转查看`；调用 `listResumeVersions(resumeAssetId)` 找到 `versionType='structured_master'` 行 → nav `resume_versions?versionId={masterId}&tab=preview`；
  - `422 VALIDATION_FAILED`：inline error + 保留草稿；
  - 其他：toast generic + 保留草稿；
- IK request spy 断言 `Idempotency-Key` header 出现；replay 行为通过 fixture `idempotency-replay` 验证。

#### 5.4 Vitest 单测
- happy：fixture `default` → 201 → toast + nav；
- replay：同 IK 二次提交 → 同 ResumeVersion 返回，不重复 nav；
- 409：fixture `already-exists-409` → 兜底 nav；
- 422：fixture `validation-422` → inline；
- 隐私：DOM 渲染 structuredProfile 但 URL / pendingAction / localStorage 不含。

### Phase 6: Home / Workspace CTA 串通 + i18n + a11y + 隐私 + UI parity + BDD + 旧入口负向

#### 6.1 Home `1 分钟创建` CTA 集成
- 集成测试：在 `frontend/src/app/scenarios/p0-081-resume-create-flow-upload-paste-guided-happy.test.tsx` 或等价位置覆盖：从 `home` route 渲染 `还没有简历？1 分钟创建` 按钮 → click → URL 切到 `resume_versions?flow=create` → 渲染 `ResumeCreateFlow`（断言 testid `resume-create-flow`）。
- 旧 `onboarding` route alias 不复活：grep `frontend/src/app/normalizeRoute.ts` 确认 alias map 仅折回 `resume_versions`；测试断言访问 `/onboarding` 折回 `/resume_versions`，仍不直接渲染旧 OnboardingScreen。

#### 6.2 Workspace `WorkspaceMissingResumeState` CTA 集成
- 集成测试：在 `frontend/src/app/scenarios/p0-018-workspace-default-render.test.tsx` 或等价 plan 001 / 002 workspace 测试中追加：missing-resume 空态 → 点击 "创建简历" CTA → URL 切到 `resume_versions?flow=create` → 渲染 `ResumeCreateFlow`；
- pendingAction：未登录态下点击 CTA 后 pendingAction 只保留 `{ route: 'resume_versions', params: { flow: 'create' } }`，不携带任何 workspace context；登录恢复后命中 `ResumeCreateFlow` upload 模式默认 tab。

#### 6.3 i18n
- 新增 key 空间：`resumeWorkshop.create.tabs.{upload,paste,guided}`、`resumeWorkshop.create.upload.*`、`resumeWorkshop.create.paste.*`、`resumeWorkshop.create.guided.*`、`resumeWorkshop.create.sidebar.{whatSaved,whatNext}`、`resumeWorkshop.parsing.*`、`resumeWorkshop.preview.*`、`resumeWorkshop.create.errors.{validation,sizeExceeded,uploadFailed,parseTimeout,alreadyExists}`；
- 测试覆盖 EN / ZH 切换的关键文案 + Accept-Language header 携带（createUploadPresign / registerResume / getResume / confirmResumeStructuredMaster 四个 op 的 generated client 请求）。

#### 6.4 a11y
- tab 切换（upload / paste / guided）使用 `role="tablist"` + `role="tab"` + `aria-selected`；
- focus 管理：进入 ResumeCreateFlow 时 focus 落在 "返回简历工坊" 或 tab list；进入 ResumeParseFlow 时 focus 落在 "取消并返回修改"；进入 ResumePreviewConfirm 时 focus 落在 "确认并保存 v1"；
- ESC：在 ResumeCreateFlow / ResumeParseFlow 中视为返回上一 stage（不是 nav 出 resume_versions）；
- 键盘：guided 5 step 之间可用 Tab + Shift+Tab 导航 + ←/→ 切换 step；
- aria-live：parsing ticker / preview confirm success / 409 / 422 错误 toast 都通过 `aria-live="polite"` 或 toast manager 提示。

#### 6.5 隐私红线
- rawText / guidedAnswers / pickedFile binary / parsedTextSnapshot / parsedSummary / structuredProfile 内容不出现在：
  - console.log / console.warn / console.error；
  - URL query / hash / route params / pendingAction params；
  - localStorage / sessionStorage / IndexedDB；
  - mock transport log / generated client request logger / observability sink；
  - error toast 内容（user-visible 错误只保留 enum / generic 文案）；
- Vitest spy + Playwright DOM/network sniff 联合 grep；
- 仅允许 clipboard / blob 输出（本 plan 不实现 copy；保留给 plan 003 + 已落地的 plan 001 PreviewTab）。

#### 6.6 UI parity gate
- 复用 [frontend-shell/003-ui-design-pixel-parity-gate](../../../frontend-shell/plans/003-ui-design-pixel-parity-gate/plan.md) 框架；
- 新增 `frontend/tests/pixel-parity/resume-workshop-create.spec.ts` 覆盖 5 个屏幕：CreateFlow Upload tab / CreateFlow Paste tab / CreateFlow Guided tab / ParseFlow / PreviewConfirm；
- desktop 1440px + mobile 390x844 viewport DOM anchor + computed style + bounding box + screenshot smoke；
- 仅在 baseline 可由 clean checkout 稳定取得 / 本 plan 显式维护时启用 screenshot diff regression；常规 PASS 证据靠 DOM/style/bounding box / 非空截图 buffer。

#### 6.7 BDD 场景验证
- 执行 `test/scenarios/e2e/p0-081-resume-create-flow-upload-paste-guided-happy/` 全 PASS（covers C-10 主路径 + 三 sourceType）。
- 执行 `test/scenarios/e2e/p0-082-resume-create-flow-parsing-failure-and-retry/` 全 PASS（parse failed / parse timeout / cancel-and-return）。
- 执行 `test/scenarios/e2e/p0-083-resume-create-flow-preview-confirm-and-cta-handoff/` 全 PASS（preview confirm happy + 409 already-exists + 422 validation + Home / Workspace CTA handoff + auth pending action）。
- 在 `test/scenarios/e2e/INDEX.md` 追加 P0.081 + P0.082 + P0.083 行（关联需求 `frontend-resume-workshop C-10`，状态 Ready，automated）。

#### 6.8 旧入口负向 grep
- `git grep -nE "welcome|mistake|growth|drill|followup|STAR|experiences|voice|OnboardingScreen|onboarding=true" -- frontend/src/app/screens/resume-workshop/create/`：0 命中；
- `git grep -nE "ui-design/src/(data|screen-resume-workshop)" -- frontend/src/app/screens/resume-workshop/create/`：0 命中（不允许 runtime import prototype data / component source）。

#### 6.9 spec / history / INDEX 同步
- 确认 frontend-resume-workshop spec.md / history.md / `docs/spec/INDEX.md` 已由本 L1 设计结晶同步到 1.1，并且 §3.1 D-4 / §6 C-10 / §7 plan 002 行指向当前 active plan；实施阶段不得为了 checklist 收口重复 bump spec 版本，除非发现新的设计事实需要原地修订。
- 确认 `docs/spec/frontend-resume-workshop/plans/INDEX.md` 已包含 002 active 行，且 Header / INDEX 投影一致。
- `sync-doc-index --check` PASS。

## 5 验收标准

- 本计划列出的 §4 所有 Phase task 全部完成；
- §3 替代验证 gate 全部通过；
- spec §6 C-10 PASS（CreateFlow 三 tab + Onboarding handoff）；C-1..C-9 / C-11 不退化（plan 001 / 003 责任范围保持）；
- BDD E2E.P0.081 + P0.082 + P0.083 PASS；
- UI parity gate 已接入 `frontend/tests/pixel-parity/resume-workshop-create.spec.ts`，clean checkout PASS 不依赖本地未跟踪 screenshot baseline；
- engineering-roadmap §5.2 `frontend-resume-workshop` 状态保持 active（plan 001 已升级；本 plan 不改动 §5.2 状态）；
- spec.md 1.1 / history.md / plans/INDEX.md / docs/spec/INDEX.md 同步至最新；
- 上游 gate 已满足：backend-resume/002 Phase 1 落地的 confirmResumeStructuredMaster fixture + generated client + 错误码均可消费。

## 6 风险与应对

| 风险 | 应对 |
|------|------|
| R1: backend-resume/002 Phase 1 (D-10 `confirmResumeStructuredMaster`) 未及时落地，导致 Preview Confirm 提交路径无 fixture / 无 client method 可用 | Phase 0 / Phase 5.1 显式 gate：未满足时 Phase 5 渲染 `<ComingSoonPreviewConfirm>`，并升级 blocker。Plan 不私造客户端协议，不绕过 D-10 schema |
| R2: `getResume` fixture 当前仅含 `default` (ready)；没有 `queued / generating / failed / parse-timeout` scenario，轮询测试只能依靠 mock client deterministic stepping | Phase 4 hook 显式标注：在 fixture 层补齐前以 mock client `__nextAttemptScenario` 模拟终态转移；在 retrospective 中向 backend-resume followup 提议补 fixture；plan 不直接修改 B2 fixture（属于 cross-owner) |
| R3: createUploadPresign 仅 `default` scenario，validation / IK replay 行为依赖 generated client error path mock；可能漏覆盖签名 URL TTL 边界 | Phase 2 单测覆盖 TTL 过期 → 重 presign 路径；mock client 在 attempt 序列内模拟 `expiresAt` 推进；如真实切真后发现 fixture 不足，由 backend-upload followup 补齐 |
| R4: pendingAction 携带 raw text / file binary 导致隐私泄露 | Phase 1.3 + 6.5 显式断言 pendingAction params 集合为 `{ route, params: { flow, createMode? } }`；Vitest grep + Playwright network sniff 联合验证 |
| R5: 双步上传 PUT 阶段失败 retryable 与 fileObjectId 复用语义不清晰，导致孤儿 `file_objects.pending` 行 | Phase 2.2 hook：同一 file retry 在 TTL 内复用 fileObjectId（不创建新行）；TTL 过期或用户切换 file 时生成新 fileObjectId（新 IK）；retrospective 中观察是否需要触发 backend-upload pending GC（[backend-upload §3.2 待确认事项](../../../backend-upload/spec.md#32-待确认事项)） |
| R6: guided 5 step navigation 在 mobile 视图被压缩或 a11y 焦点丢失 | Phase 6.4 a11y 测试覆盖 mobile viewport + 键盘导航；Phase 6.6 mobile pixel parity 测试覆盖 guided 步骤面板 |
| R7: home + workspace 两条 CTA route 行为已在 plan 001 时与 placeholder 一起验证，002 替换后存在 race condition（如 ResumeCreateFlow 默认 createMode 与 placeholder 文案差异）破坏既有 BDD | Phase 6.1 + 6.2 集成测试更新既有 P0.018 / P0.019 / P0.014 等 home / workspace 场景中关于 `resume_versions?flow=create` 渲染断言，使其指向 ResumeCreateFlow；不允许 placeholder 与 real component 同时存在断言 |
| R8: confirmResumeStructuredMaster 409 路径需要二次调 listResumeVersions 找已存在 master，可能引起 race（用户同时打开两个 tab） | Phase 5.3 fallback：409 时显式 toast "已存在主版本"，并在 hook 内调 listResumeVersions 收敛到第一个 `versionType='structured_master'` 行；如不存在（数据漂移）则降级到 list view + toast；不阻塞其他 plan |
| R9: i18n key 空间膨胀与 plan 001 / 003 冲突 | Phase 6.3 key 空间严格落在 `resumeWorkshop.create.*` / `.parsing.*` / `.preview.*`；plan 001 / 003 owner namespace 不交叉；CI lint 在 ESLint i18n rule 下捕获重复 |
