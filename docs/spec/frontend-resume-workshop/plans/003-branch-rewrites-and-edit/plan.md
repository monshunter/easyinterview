# Frontend Resume Workshop Rewrites and Edit

> **版本**: 1.4
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

把 [frontend-resume-workshop spec](../../spec.md) 的 flat Resume Workshop 合同落到当前 `frontend/` owner 文档中：

- `resume_versions` 只承载 flat list / create / detail 入口；未知 flow 回落 flat list。
- `ResumeDetailView` 通过 `getResume(resumeId)` 读取 flat resume，并保留 route `targetJobId` 给 Rewrites rerun body。
- `ResumeRewritesTab` 展示 ephemeral suggestions，只提供本地「采纳」；`RewriteSaveConfirmModal` 通过 `updateResume` 覆盖当前简历或 `duplicateResume` 另存新简历。
- accepted rewrites merge 覆盖 `structuredProfile.sections[]`、`experience[]`、`experiences[]`、`projects[]` 的 `bullets` 容器；省略 `structuredProfile` 时 preview fallback 不崩溃。
- `ResumeEditTab` 通过 `updateResume` 保存 flat `displayName` / `headline` / `summary`；Export PDF / copyText 继续复用 flat `exportResume` / `buildResumePlainText` 路径。
- i18n / a11y / privacy / UI parity gate 覆盖 flat detail / Rewrites / Edit desktop + mobile；negative grep 保证 version-tree operation、server-side suggestion decision 和 prototype runtime import 不回流。

本 plan 不修改 backend handler、异步 job、outbox event 或 AI 调用；tailor 真实运行由 backend-resume flat handlers 与 P0.077-P0.080 证明，本 plan 只消费 generated client。

## 2 质量门禁分类

- **Plan 类型**: `code-internal + feature-behavior`。本 plan 覆盖 flat resume detail / rewrites / edit 用户可见 UI 行为，并消费 generated client flat Resume / ResumeTailor operations。
- **TDD 策略**: 适用。Red tests 覆盖 route dispatch、flat detail fallback、accepted rewrites merge、rerun body context、save modal、edit save、IK header、privacy red lines 和 generated-client negative grep。
- **BDD 策略**: 适用。`E2E.P0.084`、`E2E.P0.085`、`E2E.P0.086`、`E2E.P0.087` 维护在 [bdd-plan](./bdd-plan.md) / [bdd-checklist](./bdd-checklist.md)。
- **替代验证 gate**:
  - `pnpm --filter @easyinterview/frontend test` focused Resume Workshop suite
  - `pnpm --filter @easyinterview/frontend build`
  - `pnpm --filter @easyinterview/frontend exec playwright test tests/pixel-parity/resume-workshop-branch-rewrites-edit.spec.ts`
  - P0.084-P0.087 `setup -> trigger -> verify -> cleanup`
  - `make codegen-check`
  - `sync-doc-index --check`
  - `make docs-check`
  - `git diff --check`

## 3 Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `listResumes` | `openapi/fixtures/Resumes/listResumes.json` `default` | `ResumeListView` / `useResumeAssets` flat list | `backend/internal/resume/handler/list.go` + `GET /api/v1/resumes` | `resumes` read | none | E2E.P0.084 |
| `getResume` | `openapi/fixtures/Resumes/getResume.json` `default` / `not-found-404` | `ResumeDetailView` / detail fallback | `backend/internal/resume/handler/get.go` + `GET /api/v1/resumes/{resumeId}` | `resumes` read | none | E2E.P0.084 / P0.086 / P0.087 |
| `requestResumeTailor` | `openapi/fixtures/ResumeTailor/requestResumeTailor.json` `default` / `idempotency-replay` | `useRequestResumeTailor` via Rewrites rerun | `backend/internal/resume/handler/request_tailor.go` + `POST /api/v1/resume/tailor` | `resume_tailor_runs` + `async_jobs` | downstream `resume.tailor` | E2E.P0.085 |
| `getResumeTailorRun` | `openapi/fixtures/ResumeTailor/getResumeTailorRun.json` `default` / `queued` / `generating` / `failed` | `useResumeTailorRunPolling` | `backend/internal/resume/handler/get_tailor_run.go` + `GET /api/v1/resume/tailor-runs/{tailorRunId}` | `resume_tailor_runs` read | none | E2E.P0.085 |
| `updateResume` | `openapi/fixtures/Resumes/updateResume.json` `default` / `idempotency-replay` / `validation-error-422` | overwrite accepted rewrites and Edit Tab save | `backend/internal/resume/handler/update.go` + `PATCH /api/v1/resumes/{resumeId}` | `resumes` update | none | E2E.P0.086 |
| `duplicateResume` | `openapi/fixtures/Resumes/duplicateResume.json` `default` / `idempotency-replay` / `validation-error-422` | save accepted rewrites as new | `backend/internal/resume/handler/duplicate.go` + `POST /api/v1/resumes/{resumeId}/duplicate` | `resumes` insert | none | E2E.P0.086 |
| `exportResume` | `openapi/fixtures/Resumes/exportResume.json` `p0-501-not-available` | detail header Export PDF | P0 explicit 501 stub | none in P0 | none | E2E.P0.087 |

## 4 当前实施 Gate

### 4.1 Flat Route Gate

- `parseResumeWorkshopParams` / `ResumeWorkshopScreen` 只 materialize `flow=list | create` + `resumeId` + `tab` + optional `tailorRunId` / `targetJobId`。
- `flow=branch` 或未知 flow 回落到 flat list；runtime 不渲染 `resume-branch-flow`。
- Runtime Resume Workshop source 中 `ResumeBranchFlow` / `branchResumeVersion` / `seedStrategy` / `resumeVersionId` / `resumeAssetId` / `listResumeVersions` / `getResumeVersion` 保持 0 命中。

验证：`ResumeWorkshopScreen.test.tsx` + P0.084 verify grep。

### 4.2 Rewrites Accept-Only Save Gate

- `ResumeRewritesTab` 只提供本地「采纳」；不出现 reject / inline edit / server-side suggestion decision。
- `RewriteSaveConfirmModal` 覆盖 `updateResume` overwrite 与 `duplicateResume` save-as-new。
- 保存 request 带 `Idempotency-Key`，不发送 `acceptResumeTailorSuggestion` / `rejectResumeTailorSuggestion` / `updateResumeVersion`。

验证：`ResumeRewritesTab.test.tsx` + `ResumeDetailView.test.tsx` + P0.086 verify。

### 4.3 Flat StructuredProfile Merge Gate

- accepted rewrite merge 覆盖 `structuredProfile.sections[]`、`experience[]`、`experiences[]`、`projects[]` 的 `bullets`；未匹配 bullet 保持不变。
- 保存 payload 不写 UI-only `acceptedRewrites` / modal state。
- flat `getResume` response 省略 `structuredProfile` 且没有 source text 时，preview fallback 不崩溃。

验证：`ResumeDetailView.test.tsx` sections overwrite、flat bullet arrays、omitted profile fallback regressions。

### 4.4 Tailor Rerun Context Gate

- route 解析出的 optional `targetJobId` 必须从 `ResumeWorkshopScreen` 透传到 detail container 与 Rewrites Tab rerun hook。
- 有 `targetJobId` 时 `requestResumeTailor` body 为 `{ resumeId, targetJobId, mode }`；无 `targetJobId` 时才允许 `{ resumeId, mode }` generic rerun。
- rerun body 不恢复 `resumeAssetId` / `resumeVersionId`。

验证：`ResumeDetailView.test.tsx` route-context regression + `useRequestResumeTailor.test.tsx` IK/body contract。

### 4.5 Edit Tab + Export / Copy Gate

- `ResumeEditTab` 保存通过 `updateResume` 覆盖 flat resume `displayName` / `structuredProfile.headline` / `structuredProfile.summary`；422 / 409 / 404 error mapping 仍可见。
- Export PDF 使用 `exportResume` P0 501 friendly toast；copyText 使用 `buildResumePlainText` clipboard path；Rewrites / Edit tab 切换不隐藏 header export。

验证：`ResumeEditTab.test.tsx` + `ResumeDetailExport.test.tsx` + P0.087 verify。

### 4.6 UI Parity / Privacy / BDD Wrapper Gate

- P0.084-P0.087 trigger/verify 使用当前 flat test files，verify 显式拒绝 no-test / fail marker，并检查 real-backend generated-client preflight marker。
- P0.087 Playwright parity 跑 `frontend/tests/pixel-parity/resume-workshop-branch-rewrites-edit.spec.ts`，证明 flat detail/Rewrites/Edit desktop + mobile DOM / style / bounding / screenshot smoke / axe。
- 隐私 gate 覆盖 originalBullet / suggestedBullet / matchSummary / structuredProfile / manual edit 文本不进入 URL / pendingAction / localStorage / mock transport log / toast。

验证：P0.084-P0.087 `setup -> trigger -> verify -> cleanup` + focused/full frontend gates。

## 5 验收标准

- Flat Resume Workshop 当前路径只依赖 §3 的 7 个 operationId。
- §4 所有 gate 已完成并有 checklist 证据。
- BDD E2E.P0.084 / P0.085 / P0.086 / P0.087 按当前 flat scenario contract PASS。
- BUG-0123 类 gate 已固化：omitted `structuredProfile` fallback、accepted rewrites 写入 flat bullet containers、route `targetJobId` 进入 rerun body。
- UI parity gate 已接入 `frontend/tests/pixel-parity/resume-workshop-branch-rewrites-edit.spec.ts`，clean checkout PASS 不依赖本地未跟踪 screenshot baseline。
- `make codegen-check` 证明 version-tree operation generated surface 不回流。
- `sync-doc-index --check`、`make docs-check`、`git diff --check` 通过。

## 6 风险与应对

| 风险 | 应对 |
|------|------|
| R1: OpenAPI / generated client shape 与 flat frontend hooks 漂移 | §4.1 / §4.6 通过 codegen-check、real-mode gate 和 runtime negative grep 阻断；缺失则停止并回到对应 owner 修复 |
| R2: accepted rewrite 保存漏掉 D-20 flat `experience[]` / `projects[]` | §4.3 固化 flat profile merge regression，必须覆盖 `sections`、`experience`、`experiences`、`projects` |
| R3: `getResume` response 省略 `structuredProfile` 时 detail fallback 崩溃 | §4.3 固化 omitted profile fallback regression；fixture 或 component test 任一失败都不能收口 |
| R4: `targetJobId` 只停留在 root route data attribute，rerun body 丢失 JD context | §4.4 要求 body-level assertion；有 `targetJobId` 时 `requestResumeTailor` 必须携带，且不恢复 asset/version ids |
| R5: IK header fixture 与 generated client `opts.idempotencyKey` 行为漂移 | §4.2 / §4.5 用 hook/spy tests 断言 `requestResumeTailor`、`updateResume`、`duplicateResume`、`exportResume` 的 IK 行为 |
| R6: polling hook 资源泄漏 | `useResumeTailorRunPolling` cleanup + P0.085 unmount fake-timer gate 证明无后续 `getResumeTailorRun` |
| R7: Experience / Skills 编辑范围扩展导致用户期待与实现不一致 | 当前 P0 保存字段锁定为 `displayName` / `headline` / `summary`；若 UI 真理源扩展 Experience / Skills 编辑，先更新 `ui-design/` 与 spec，再新增 plan |

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-07 | 1.4 | Reconcile owner plan to the current flat Resume Workshop Rewrites/Edit contract and remove non-current phase prose. |
| 2026-06-14 | 1.3 | Align D-20 flat Resume contract with accept-only rewrites save, `updateResume`, `duplicateResume`, and route `targetJobId` gates. |
| 2026-05-23 | 1.2 | Add real-backend generated-client gate for P0.084-P0.087. |
| 2026-05-18 | 1.1 | Complete Rewrites/Edit UI, polling, save, export/copy, i18n, privacy, a11y, parity, and BDD gates. |
| 2026-05-17 | 1.0 | Initial Resume Workshop Rewrites/Edit implementation plan. |
