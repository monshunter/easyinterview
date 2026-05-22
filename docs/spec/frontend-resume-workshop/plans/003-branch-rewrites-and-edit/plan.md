# Frontend Resume Workshop Branch, Rewrites and Edit

> **版本**: 1.1
> **状态**: completed
> **更新日期**: 2026-05-23

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

> 2026-05-23 L2 real-backend gate remediation：P0.084-P0.087 trigger 已前置 `frontendOwners.realApiMode.test.ts`，verify 检查 `VITE_EI_API_MODE=real`、默认 backend base URL 与测试文件 marker；branch/rewrites/edit 的 fixture-backed UI variants 继续保留，但真实 resume-version / tailor / suggestion / export generated-client routing 由集中 gate 证明。

## 1 目标

把 [frontend-resume-workshop spec](../../spec.md) §6 C-11（BranchFlow + Rewrites Tab + Edit Tab + exportPDF / copyText 真实可用）落到 `frontend/` 实现：

- 替换 [001-listing-routing-and-detail-readonly](../001-listing-routing-and-detail-readonly/plan.md) 阶段在 `ResumeWorkshopScreen` 中对 `flow=branch` 渲染的 `<NotImplementedPlaceholder>`，源级复刻 [`ui-design/src/screen-resume-workshop.jsx`](../../../../../ui-design/src/screen-resume-workshop.jsx) 中以下组件：
  - `ResumeBranchFlow`（路由容器 + BRANCHING FROM 卡 + 版本名 / 目标岗位 / 侧重方向 / Bullet 初始化 3 路 seedStrategy + 创建动作）
  - `ResumeRewritesTab`（仅 TARGETED 版本可见：scope banner + bullet 列表 + diff 详情 + accept / reject / edit inline + accepted / pending / rejected 计数 + tailor run re-run）
  - `ResumeEditTab`（MASTER 与 TARGETED 都可见：headline + summary + experience + skills section + save changes）
- 实现 BranchFlow 提交契约：generated client `branchResumeVersion({ parentVersionId, targetJobId, seedStrategy, displayName, focusAngle })` + `Idempotency-Key`。三 seedStrategy 三种响应路径：
  - `copy_master` → 201 + `ResumeVersion`（同步）→ toast + nav `resume_versions?versionId={newId}&tab=rewrites`；
  - `blank` → 201 + `ResumeVersion`（同步，structured_profile 空）→ toast + nav `resume_versions?versionId={newId}&tab=edit`（空白起步默认进入编辑标签）；
  - `ai_select` → 202 + `BranchResumeVersionAccepted{resumeVersionId, version, job(jobType=resume_tailor, status=queued)}`→ toast `已创建定制版本 · AI 正在生成 bullet` → nav `resume_versions?versionId={newId}&tab=rewrites`，并在 Rewrites Tab 内启动 `useResumeTailorRunPolling`。
- 实现 Rewrites Tab AI 改写决策面 + 终态状态机：
  - 数据来源：`getResumeVersion(versionId)` 返回的 `version.suggestions[]`（pending / accepted / rejected），辅以 `getResumeTailorRun(tailorRunId)` 在 ai_select / 用户触发 re-run 时的轮询；
  - 单条 bullet 操作：
    - Accept → `acceptResumeTailorSuggestion(versionId, suggestionId)` + IK → 写入 suggestion.status='accepted' + decided_at；UI 行 status 变 accepted；不自动 patch `version.structured_profile`（与 [backend-resume D-12](../../../backend-resume/spec.md#31-已锁定决策) 对齐）；
    - Reject → `rejectResumeTailorSuggestion(versionId, suggestionId)` + IK → 写入 status='rejected'；UI 行 status 变 rejected（line-through）；
    - Manual edit → 用户在 inline textarea 修改 → "保存人工改写" → 当前 generated client 的 `acceptResumeTailorSuggestion` 为 bodyless operation，因此先调用 `updateResumeVersion` patch 当前 version 的 `structuredProfile.manualEdits[]`（含 `suggestionId` + edited text），成功后再调用 bodyless `acceptResumeTailorSuggestion(versionId, suggestionId)` 将该 suggestion 标为 terminal accepted；若 update 成功但 accept 失败，UI 保留 saved-manual-pending retry 状态，不私造 `manualEditText` 协议；
  - "重新运行改写"：`requestResumeTailor({ resumeAssetId, targetJobId, mode })` + IK → 启动 polling → ready 后 suggestions[] 刷新；mode 选项按 UI 真理源默认 `bullet_suggestions`，可由 UI 切到 `gap_review`（与 [backend-resume D-5](../../../backend-resume/spec.md#31-已锁定决策) 对齐）；
- 实现 Edit Tab 结构化编辑 + 保存：
  - 显示 plan 001 占位的 `<ComingSoonTab>` 替换为 `ResumeEditTab` 实体；
  - 字段：Headline / Summary / Experience 列表 / Skills 列表；与 UI 真理源 1:1 复刻；
  - 保存 → `updateResumeVersion(versionId, { displayName?, structuredProfile?, focusAngle?, matchScore? })` + IK；422 inline；409 toast 提示 IK 冲突；不可改字段如 `versionType / resumeAssetId / parentVersionId / targetJobId / seedStrategy` 在 UI 隐藏，并在 mapper 中强制拒绝；
  - MASTER 与 TARGETED 都可编辑，但 Master scope banner 显式提示 "正在编辑主版本"；
- Export PDF 按钮在 TARGETED detail 与 Rewrites / Edit Tab 切换时保持 [plan 001 Phase 3.7](../001-listing-routing-and-detail-readonly/plan.md#phase-3-resumedetailview-preview-tab--原件弹层) P0 行为不退化：`exportResumeVersion(versionId, { idempotencyKey })` + `Idempotency-Key` header + 501 `RESUME_EXPORT_NOT_AVAILABLE` → toast `PDF 导出能力即将开放`；copyText 真实可用（`buildResumePlainText` 投影 → clipboard）；本 plan 在 Rewrites / Edit Tab 入口处复用 plan 001 的 `useResumeExport` 与 `useResumePlainTextCopy` hook，不重复实现；
- i18n（`resumeWorkshop.branch.* / .rewrites.* / .edit.*` namespace）+ a11y（focus / aria / keyboard / scope banner aria-live）+ 隐私红线（originalBullet / suggestedBullet / match_summary / structuredProfile / manualEdit 文本不出现在 console / URL / pendingAction / localStorage / mock transport log / telemetry / toast 内容）+ UI parity gate（Vitest + Playwright pixel parity 对 `ResumeBranchFlow` / `ResumeRewritesTab` / `ResumeEditTab` 三屏 desktop + mobile 断言）；
- 不修改 backend handler / 异步 job / outbox event / AI 调用（这些已由 [backend-resume/002 Phase 4..8](../../../backend-resume/plans/002-versions-tailor-runs-and-save-v1/plan.md) 落地，本 plan 只消费 generated client）；不实现 exportResumeVersion 真实 PDF（仍 P0 stub）；不实现 archiveResumeAsset / 完整 privacy delete（P1）。

## 2 背景

本 plan 是 frontend-resume-workshop 第三批 plan，承担 P0 用户路径 "从已有主版本分叉岗位定制版本 → AI 生成 bullet 建议 → 决策 accept / reject / 人工编辑 → 在 Edit Tab 调整结构化字段 → 保存改动" 的前端端到端。它解锁 [001](../001-listing-routing-and-detail-readonly/plan.md) Detail 中三 tab 体系里 Rewrites / Edit 的实际能力，并把 [002](../002-create-flow-and-onboarding/plan.md) 创建出来的 `structured_master` 链路扩展到 `targeted` 版本族：

- BranchFlow 是用户在已有简历树上派生岗位定制版本的唯一入口；本 plan 完成前 `nav('resume_versions', { flow: 'branch', branchOriginalId })` 仍命中 `<NotImplementedPlaceholder>`。
- Rewrites Tab 是 TARGETED 版本的默认 tab（`resumeDefaultTab(version)` 返回 `'rewrites'`）；plan 001 阶段 TARGETED 默认 tab 不变，但内容为 `<ComingSoonTab>`；本 plan 切实把 AI 改写决策面接通。
- Edit Tab 是 MASTER 与 TARGETED 通用的结构化编辑面板；plan 001 阶段同为 `<ComingSoonTab>`；本 plan 接通真实 `updateResumeVersion` 写入路径。
- 与 [backend-resume/002 Phase 4..8](../../../backend-resume/plans/002-versions-tailor-runs-and-save-v1/plan.md) 协作：当前分支已具备真实 handler / `cmd/api` route / generated client / fixture parity。Frontend 003 仍可用 fixture-backed dev/test 先行，但 Phase 0 必须把这些 real-backend facts 作为回归 gate。
- 已验证的 fixture 形态：`branchResumeVersion.json` 六 scenario；`requestResumeTailor.json default / idempotency-replay`；`getResumeTailorRun.json default(ready) / queued / generating / failed`；`acceptResumeTailorSuggestion.json default / idempotency-replay / already-decided-409`；`rejectResumeTailorSuggestion.json default / idempotency-replay / already-decided-409`；`updateResumeVersion.json default / idempotency-replay / validation-error-422`。`already-decided-409` 已收敛为 `error.code='VALIDATION_FAILED'` + `details.reason='SUGGESTION_ALREADY_DECIDED'`。

每个 phase 是可独立验证的纵向切片：Phase 1 起来就有 ResumeBranchFlow 容器 + 路由 + auth gate；Phase 2 起来就有 branchResumeVersion 三 seedStrategy 提交链路；Phase 3 起来就有 Rewrites Tab UI 与 getResumeVersion 投影；Phase 4 起来就有 accept / reject / manual edit 终态状态机；Phase 5 起来就有 requestResumeTailor + getResumeTailorRun 轮询；Phase 6 起来就有 Edit Tab + updateResumeVersion；Phase 7 起来就有 i18n + a11y + 隐私 + UI parity + BDD + 旧入口负向。

执行本 plan 前必须确认：

- [frontend-resume-workshop/001-listing-routing-and-detail-readonly](../001-listing-routing-and-detail-readonly/plan.md) completed；ResumeDetailView 三 tab 容器已就位（Preview tab 真实、Rewrites / Edit `<ComingSoonTab>` 占位、tab 切换 / URL `tab` param 行为已稳定）。
- [frontend-resume-workshop/002-create-flow-and-onboarding](../002-create-flow-and-onboarding/plan.md) 当前分支实现已落地；`flow=create` 渲染 `ResumeCreateFlow`，Home / Workspace CTA handoff 已可走通。003 Phase 0 只需反查当前代码事实与 plan 002 lifecycle 状态，避免把旧 placeholder 当作仍存在的 create 主路径。
- [backend-resume/002](../../../backend-resume/plans/002-versions-tailor-runs-and-save-v1/plan.md) Phase 4..8 已完成（updateResumeVersion / branchResumeVersion 三路 / requestResumeTailor / getResumeTailorRun / acceptResumeTailorSuggestion / rejectResumeTailorSuggestion handler + cmd/api wiring + fixture 全 ready；本 plan 直接消费）。
- [openapi-v1-contract/004-resume-additive-coverage](../../../openapi-v1-contract/plans/004-resume-additive-coverage/plan.md) Phase 1-5 已完成（generated client 含 8 个 Resume / ResumeTailor op）。
- UI 真理源 [`ui-design/src/screen-resume-workshop.jsx`](../../../../../ui-design/src/screen-resume-workshop.jsx) 的 `ResumeBranchFlow / ResumeRewritesTab / ResumeEditTab` 三组件 + [`docs/ui-design/resume-module.md`](../../../../ui-design/resume-module.md) v1.7 + [`docs/ui-design/jd-resume-management.md`](../../../../ui-design/jd-resume-management.md) v1.5 active。

## 3 质量门禁分类

- **Plan 类型**: `code-internal + feature-behavior`。本 plan 实现前端容器 + 表单 + 多 op IK 调用 + tailor run 轮询 + 终态状态机；用户可见 UI 行为。
- **TDD 策略**: 适用。Red-Green-Refactor 入口：
  1. Vitest 组件单测：`ResumeBranchFlow` 表单校验 / 3 seedStrategy 切换 / 提交 disabled 边界 / IK 行为；`ResumeRewritesTab` 列表选中 / 计数 / accept / reject / edit inline 切换 / scope banner i18n；`ResumeEditTab` 表单 + save / scope banner master vs targeted；
  2. adapter unit test：
     - `mapBranchFormToBranchResumeVersionRequest`：3 seedStrategy × 表单字段 → `BranchResumeVersionRequest` payload；
     - `mapSuggestionToUiRow` / `mapAcceptSuggestionResponseToVersion`：accept / reject 响应字段投影；
     - `mapEditTabFieldsToUpdateResumeVersionRequest`：Headline / Summary / Experience / Skills → `UpdateResumeVersionRequest`；过滤不可编辑字段（versionType / resumeAssetId / parentVersionId / targetJobId / seedStrategy）；
  3. fixture parity test：
     - `Resumes/branchResumeVersion.json` 六 scenario（`default / copy-master-sync / blank-sync / ai-select-202-with-job / idempotent-replay / validation-error-422`）；
     - `ResumeTailor/requestResumeTailor.json default / idempotency-replay` + `getResumeTailorRun.json default / queued / generating / failed`；
     - `Resumes/acceptResumeTailorSuggestion.json default / idempotency-replay / already-decided-409` + `rejectResumeTailorSuggestion.json default / idempotency-replay / already-decided-409`；
     - `Resumes/updateResumeVersion.json default / idempotency-replay / validation-error-422`；
     - `Resumes/exportResumeVersion.json p0-501-not-available`（plan 001 已消费，本 plan 不退化）；
  4. Idempotency-Key contract test：`branchResumeVersion / requestResumeTailor / acceptResumeTailorSuggestion / rejectResumeTailorSuggestion / updateResumeVersion / exportResumeVersion` 六个 op 通过 `generateIdempotencyKey()` 生成 IK，request spy 断言 `Idempotency-Key` header 出现；replay 行为通过 fixture 验证；
  5. tailor run polling test：`useResumeTailorRunPolling` 在 ai_select branch + 用户 re-run tailor 两条路径下 deterministic stepping `queued → generating → ready` / `→ failed`；mock harness 显式标注；
  6. 隐私 grep test：originalBullet / suggestedBullet / matchSummary / structuredProfile / manualEdit 文本不出现在 URL / pendingAction params / localStorage / mock transport log / console / toast；
  7. auth boundary test：未登录访问 `resume_versions?flow=branch&branchOriginalId={id}` 不触发 branchResumeVersion；未登录访问 detail / Rewrites / Edit tab 不触发 protected ops；pendingAction 仅携带 route params，不携带 form draft；
  8. Playwright pixel parity：`ResumeBranchFlow` + `ResumeRewritesTab` + `ResumeEditTab` 三屏 desktop 1440px + mobile 390x844 DOM anchor + computed style + bounding box + screenshot smoke（baseline 可复现时启用 screenshot diff）；
  9. negative grep test：`frontend/src/app/screens/resume-workshop/branch/` 与 `tabs/` 不出现 retired 模块名（welcome / mistake / growth / drill / followup / 旧 STAR / 旧 experiences / voice / 旧 onboarding）；不出现 retired tailor mode `inline | rewrite | mirror`（与 [event-and-outbox-contract D-14](../../../event-and-outbox-contract/spec.md#31-已锁定决策含-jobtype-映射表) 同步）；不 import `ui-design/src/screen-resume-workshop.jsx` / `ui-design/src/data.jsx` 作为运行时依赖。

  执行入口：`/implement frontend-resume-workshop/003-branch-rewrites-and-edit` → `/tdd`。

- **BDD 策略**: 适用（Feature plan requires BDD）。`E2E.P0.084` branch-flow-three-seed-strategies-happy + `E2E.P0.085` rewrites-tab-tailor-run-polling-and-suggestions + `E2E.P0.086` suggestion-accept-reject-edit-and-update-version + `E2E.P0.087` resume-detail-export-copy-consistency-and-parity，详见 [bdd-plan.md](./bdd-plan.md) / [bdd-checklist.md](./bdd-checklist.md)。

- **替代验证 gate**:
  - `pnpm --filter @easyinterview/frontend test` (Vitest)
  - `pnpm --filter @easyinterview/frontend build` + `pnpm --filter @easyinterview/frontend test:pixel-parity` (Playwright)
  - `pnpm --filter @easyinterview/frontend lint` (ESLint + UI parity rules)
  - `pnpm --filter @easyinterview/frontend build`
  - `git grep -nE "welcome|mistake|growth|drill|followup|STAR|experiences|voice|OnboardingScreen|onboarding=true" -- frontend/src/app/screens/resume-workshop/branch/ frontend/src/app/screens/resume-workshop/tabs/`（旧入口 negative；当前 plan 文档 prose 不纳入 raw zero-hit）
  - `git grep -nE "(^|[^A-Za-z0-9_])(inline|rewrite|mirror)([^A-Za-z0-9_]|$)" -- frontend/src/app/screens/resume-workshop/branch/ frontend/src/app/screens/resume-workshop/tabs/`（retired tailor mode negative；与 B3 D-14 同步）
  - `git grep -nE "ui-design/src/(data|screen-resume-workshop)" -- frontend/src/app/screens/resume-workshop/branch/ frontend/src/app/screens/resume-workshop/tabs/`（原型 runtime import negative）
  - `sync-doc-index --check`

### 3.1 Frontend / Backend Operation Matrix

本 plan 走 fixture-backed frontend + real-backend preflight path：分支创建 / 改写运行 / 改写决策 / 结构化编辑 / 导出与复制共五类操作。这里的 fixture-backed dev/test 只表示 Phase 0 gate 通过后可用当前 OpenAPI fixture 做前端确定性验证，不表示可以绕过 backend-resume/002 对 operation / fixture / generated artifact 的收敛要求。

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `branchResumeVersion` | `openapi/fixtures/Resumes/branchResumeVersion.json` `default` / `copy-master-sync` / `blank-sync` / `ai-select-202-with-job` / `idempotent-replay` / `validation-error-422` | `useResumeBranchSubmit` hook（Phase 2）：传入表单数据 + IK；按 seedStrategy 三态映射 nav target | `backend/internal/resume/handler/branch_version.go` + `cmd/api` `POST /api/v1/resume-versions` real route ready | `resume_versions` + 可选 `resume_tailor_runs(queued)` + `async_jobs(resume_tailor)`（ai_select 路径） | `resume.tailor` async downstream (ai_select；frontend 不调 AI) | E2E.P0.084 |
| `requestResumeTailor` | `openapi/fixtures/ResumeTailor/requestResumeTailor.json` `default` / `idempotency-replay` | `useRequestResumeTailor` hook（Phase 5）：从 Rewrites Tab "重新运行改写" 触发 + IK | `backend/internal/resume/handler/request_tailor.go` + `cmd/api` `POST /api/v1/resume/tailor` real route ready | `resume_tailor_runs(queued)` + `async_jobs(resume_tailor)` | downstream `resume.tailor` | E2E.P0.085 |
| `getResumeTailorRun` | `openapi/fixtures/ResumeTailor/getResumeTailorRun.json` `default` (ready) / `queued` / `generating` / `failed` | `useResumeTailorRunPolling` hook（Phase 5）：指数退避轮询 + 终态退出；ready 后触发 getResumeVersion 刷新 suggestions[] | `backend/internal/resume/handler/get_tailor_run.go` + `cmd/api` `GET /api/v1/resume/tailor-runs/{tailorRunId}` real route ready | `resume_tailor_runs` read | none in read path | E2E.P0.085 |
| `acceptResumeTailorSuggestion` | `openapi/fixtures/Resumes/acceptResumeTailorSuggestion.json` `default` / `idempotency-replay` / `already-decided-409` | `useAcceptResumeTailorSuggestion` hook（Phase 4）：bodyless accept CTA + IK；replay / 409 / cross-user error mapping | `backend/internal/resume/handler/accept_suggestion.go` + `cmd/api` accept route ready | `resume_version_suggestions.status='accepted' + decided_at`（不自动 patch structured_profile） | none | E2E.P0.086 |
| `rejectResumeTailorSuggestion` | `openapi/fixtures/Resumes/rejectResumeTailorSuggestion.json` `default` / `idempotency-replay` / `already-decided-409` | `useRejectResumeTailorSuggestion` hook（Phase 4）：bodyless reject CTA + IK | `backend/internal/resume/handler/reject_suggestion.go` + `cmd/api` reject route ready | `resume_version_suggestions.status='rejected' + decided_at` | none | E2E.P0.086 |
| `updateResumeVersion` | `openapi/fixtures/Resumes/updateResumeVersion.json` `default` / `idempotency-replay` / `validation-error-422` | `useUpdateResumeVersion` hook（Phase 6）：Edit Tab Save + Rewrites Tab manual edit explicit patch；filter 不可编辑字段 | `backend/internal/resume/handler/update_version.go` + `cmd/api` `PATCH /api/v1/resume-versions/{resumeVersionId}` real route ready | `resume_versions` UPDATE | none | E2E.P0.086 |
| `exportResumeVersion` | `openapi/fixtures/Resumes/exportResumeVersion.json` `p0-501-not-available`（plan 001 已消费） | 复用 plan 001 `useResumeExport` hook；本 plan 在 Rewrites / Edit Tab 切换上下文不退化行为 | P0 explicit 501 stub；P1 由 backend-resume/003 切真 | none in P0 | none | E2E.P0.087 |
| `getResumeVersion` | `openapi/fixtures/Resumes/getResumeVersion.json` `default` / `master-default` / `targeted-with-suggestions` / `not-found-404`（plan 001 已消费） | `useResumeVersion` hook（plan 001 已实现）：本 plan 在 branch 完成后强制 refetch；在 Rewrites Tab + Edit Tab 切换时复用 | `backend/internal/resume/handler/version_read.go` + `cmd/api` `GET /api/v1/resume-versions/{resumeVersionId}` real route ready | `resume_versions` read | none | E2E.P0.084 + P0.085 + P0.086 |

### 3.2 上游依赖 gate（必须在本 plan 落地前确认）

- 反查 [backend-resume/002 Phase 4..8](../../../backend-resume/plans/002-versions-tailor-runs-and-save-v1/plan.md) 与 cross-owner fixture 当前事实：
  - `branchResumeVersion / requestResumeTailor / getResumeTailorRun / acceptResumeTailorSuggestion / rejectResumeTailorSuggestion / updateResumeVersion` generated client + real handler + `cmd/api` route 均存在；
  - `acceptResumeTailorSuggestion.json` / `rejectResumeTailorSuggestion.json` 已包含 `default / idempotency-replay / already-decided-409`，409 body 为 `error.code='VALIDATION_FAILED'` + `error.details.reason='SUGGESTION_ALREADY_DECIDED'`；
  - `requestResumeTailor.json default / idempotency-replay` 均带 `Idempotency-Key`，`getResumeTailorRun.json` 已包含 `queued / generating / default(ready) / failed`；
  - 若上述事实在实施前缺失，按回归 blocker 停止相关 Phase，不私造客户端协议；
- 验证 plan 001 + plan 002 当前实现状态：`flow=create` 已渲染 `ResumeCreateFlow`，`flow=branch` 路径仍渲染 `<NotImplementedPlaceholder>`，Rewrites / Edit tab 仍为 `<ComingSoonTab>`（plan 003 替换边界清晰）；
- 验证 plan 001 阶段三 tab 容器的 `<ComingSoonTab>` 占位渲染契约：本 plan Phase 3 / 6 替换 Rewrites / Edit tab content 时不破坏 tab 容器 / testid / `tab` URL param 状态机。

## 4 实施步骤

### Phase 0: 上游依赖 gate + retired drift baseline

#### 0.1 上游 fixture / handler 状态确认
- 确认 backend-resume/002 Phase 4..8 当前事实仍成立：`branchResumeVersion / requestResumeTailor / getResumeTailorRun / acceptResumeTailorSuggestion / rejectResumeTailorSuggestion / updateResumeVersion` generated server interface、real handler 与 `cmd/api` route 均存在；
- 确认 fixture 名称与 error envelope 形态：accept / reject fixture 必须包含 `default / idempotency-replay / already-decided-409`，且 409 body 为 `error.code='VALIDATION_FAILED'` + `error.details.reason='SUGGESTION_ALREADY_DECIDED'`；requestTailor 必须含 `default / idempotency-replay` 且 request header 带 `Idempotency-Key`；getTailorRun 必须含 `queued / generating / default(ready) / failed`。若任一事实缺失，Phase 4 / 5 / E2E.P0.085-P0.086 暂停并转回 backend owner 修复，不允许 frontend plan 以旧 envelope 或 synthetic schema 收口。

#### 0.2 retired drift baseline
- `git grep -nE "(^|[^A-Za-z0-9_])(inline|rewrite|mirror)([^A-Za-z0-9_]|$)" -- frontend/src/app/screens/resume-workshop/`：0 命中前置；
- `git grep -nE "welcome|mistake|growth|drill|followup|STAR|experiences|voice|OnboardingScreen|onboarding=true" -- frontend/src/app/screens/resume-workshop/`：0 命中前置（plan 001 / 002 阶段已 enforce，Phase 7.10 收口时再次验证）。

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
- `git grep -nE "welcome|mistake|growth|drill|followup|STAR|experiences|voice|OnboardingScreen|onboarding=true" -- frontend/src/app/screens/resume-workshop/branch/ frontend/src/app/screens/resume-workshop/tabs/`：0 命中；
- `git grep -nE "(^|[^A-Za-z0-9_])(inline|rewrite|mirror)([^A-Za-z0-9_]|$)" -- frontend/src/app/screens/resume-workshop/branch/ frontend/src/app/screens/resume-workshop/tabs/`：0 命中（B3 D-14 同步）；
- `git grep -nE "ui-design/src/(data|screen-resume-workshop)" -- frontend/src/app/screens/resume-workshop/branch/ frontend/src/app/screens/resume-workshop/tabs/`：0 命中。

#### 7.8 spec / history / INDEX 同步
- 确认 frontend-resume-workshop spec.md / history.md / `docs/spec/INDEX.md` 已由本 L1 设计结晶同步到 1.1，并且 §6 C-11 / §7 plan 003 行指向当前 active plan；实施阶段不得为了 checklist 收口重复 bump spec 版本，除非发现新的设计事实需要原地修订。
- 确认 §3.2 accept/reject 口径为 UI 真理源 inline action + terminal-state feedback，不引入未在 `ui-design/` 出现的独立 ConfirmDialog。
- 确认 `docs/spec/frontend-resume-workshop/plans/INDEX.md` 已包含 003 active 行，且 Header / INDEX 投影一致。
- `sync-doc-index --check` PASS。

## 5 验收标准

- 本计划列出的 §4 所有 Phase task 全部完成；
- §3 替代验证 gate 全部通过；
- spec §6 C-11 PASS（BranchFlow + Rewrites Tab + Edit Tab + exportPDF/copyText 一致性）；C-1..C-10 不退化；
- BDD E2E.P0.084 + P0.085 + P0.086 + P0.087 PASS；
- UI parity gate 已接入 `frontend/tests/pixel-parity/resume-workshop-branch-rewrites-edit.spec.ts`，clean checkout PASS 不依赖本地未跟踪 screenshot baseline；
- engineering-roadmap §5.2 `frontend-resume-workshop` 状态保持 active；
- spec.md 1.1 / history.md / plans/INDEX.md / docs/spec/INDEX.md 同步至最新；除非本 plan 实施中引入新的设计事实，否则不重复 bump spec 版本；
- 上游 gate 已满足：backend-resume/002 Phase 4..8 落地的 branchResumeVersion / requestResumeTailor / getResumeTailorRun / acceptResumeTailorSuggestion / rejectResumeTailorSuggestion / updateResumeVersion 真实可用，frontend hook 通过 generated client 切真不需重构。

## 6 风险与应对

| 风险 | 应对 |
|------|------|
| R1: backend-resume/002 已完成但未来 fixture / generated client / cmd/api route 回归，导致 frontend hook 与真实后端不一致 | Phase 0 把当前 real-backend facts 固化为回归 gate；缺失则停止相关 Phase 并回到 backend owner 修复，不私造客户端协议 |
| R2: `acceptResumeTailorSuggestion` / `rejectResumeTailorSuggestion` 的 `already-decided-409` fixture 或 error details path 漂移导致 hook error mapping 与 backend 不一致 | Phase 0.1 / Phase 4 fixture parity 直接读取 `error.details.reason='SUGGESTION_ALREADY_DECIDED'`；若回到旧 `conflict-409` / `TARGET_INVALID_STATE_TRANSITION`，按 regression blocker 处理 |
| R3: IK header fixture 与 generated client `opts.idempotencyKey` 行为漂移 | Phase 0.1 / Phase 5.1 / Phase 6.2 用 fixture `default / idempotency-replay` + request spy 同时断言；fixture 缺 header 时按 regression blocker 处理 |
| R4: `getResumeTailorRun` status sequence 只在单 fixture scenario 中表达，polling 测试若继续使用 synthetic mock 可能绕过真实 schema | Phase 5.2 使用 `queued / generating / default(ready) / failed` fixture variants 组成 deterministic sequence；只允许 mock 调度顺序，不 mock response schema |
| R5: Rewrites Tab manual edit 若误用不存在的 `manualEditText` accept body，会与当前 generated client 签名冲突 | Phase 4.3 锁定 `updateResumeVersion` explicit patch + bodyless accept 两步；新增 type-level / spy test 断言 accept/reject request body 为 `undefined` |
| R6: `ai_select` branch 同事务返回 ResumeVersion + Job 时前端拿到 tailorRunId 的字段路径假设可能错位 | Phase 2.1 / 2.2 显式以 `BranchResumeVersionAccepted{resumeVersionId, version, job}` 形态消费；fixture parity test PASS 后才允许 Phase 5 启动 polling；未来 fixture 字段差异由 fixture parity test 第一时间捕获 |
| R7: Edit Tab P0 仅打通 headline + summary，Experience / Skills 列表的 add / edit / remove 未实现可能让用户期待落空 | Phase 6.1 显式声明 Experience / Skills section P0 仅 placeholder 渲染（不可编辑 individual items）；UI 真理源 Add button 仅 toast `敬请期待`；retrospective 列入 follow-up plan |
| R8: branchFlow + Rewrites Tab + Edit Tab 三屏 pixel parity baseline 数量翻倍，可能引起 CI 时间膨胀 | Phase 7.4 复用 frontend-shell/003 pipeline；screenshot 仅 smoke + DOM/style/bounding box 优先；新机器先跑 `test:pixel-parity:install` 缓存浏览器；retrospective 监控 CI 时长 |
| R9: 用户在 ai_select branch 后离开 Rewrites Tab（切到 Edit Tab 或返回 list），polling hook 资源泄漏 | Phase 5.2 hook 内置 cleanup（component unmount cancel polling）；切换 tab 时 polling 状态保留在父 detail container 或在 unmount 时取消 |
| R10: i18n key namespace 与 plan 002 冲突 | Phase 7.1 严格 namespace；`resumeWorkshop.branch.*` / `.rewrites.*` / `.edit.*` / `.tailor.*` 与 plan 002 `.create.*` / `.parsing.*` / `.preview.*` 不交叉；ESLint i18n rule 捕获重复 |
