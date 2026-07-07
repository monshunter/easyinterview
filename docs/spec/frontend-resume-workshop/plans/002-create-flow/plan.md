# Frontend Resume Workshop Create Flow

> **版本**: 1.4
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

本计划承接当前 Resume Workshop 创建路径：

- `resume_versions?flow=create` 渲染 `ResumeCreateFlow`。
- 创建入口只提供 upload / paste 两种输入。
- Upload 路径通过 `createUploadPresign`、browser PUT 和 `registerResume` 完成注册。
- Paste 路径通过 `registerResume` 完成注册，并使用原始文本派生临时可识别标题，不提交通用“粘贴的简历”。
- 注册成功后直接导航到 `resume_versions?resumeId=<id>`，打开只读详情页展示原始内容。
- 创建流不渲染解析动画页、预览确认页或确认保存页；不在 create-flow 中轮询 `getResume` 或调用 `updateResume`。
- Home “1 分钟创建” 与 Workspace missing-resume CTA 都进入当前 create flow。

本计划不实现 backend upload/resume handler、OpenAPI 契约、Resume Rewrites/Edit、PDF 导出或 Practice handoff。

## 2 背景

当前产品已采用 flat Resume IA：简历是平铺资产，创建流程只需要 upload / paste 输入和注册成功后的直接详情跳转。正式前端必须源级复刻 `ui-design/src/screen-resume-workshop.jsx` 的当前 CreateFlow 构图，同时通过 generated client 与 backend-resume / backend-upload 合同集成。

当前实现事实：

- `frontend/src/app/screens/resume-workshop/ResumeWorkshopScreen.tsx` 将 `flow=create` dispatch 到 `ResumeCreateFlow`。
- `frontend/src/app/screens/resume-workshop/create/` 包含 create-flow 容器、tabs、upload/paste hooks 和旧 parsing/preview-confirm negative tests。
- P0.081 / P0.083 场景资产覆盖直接详情跳转、CTA handoff 和隐私边界；原 P0.082 parse-failure 场景退役为 non-current negative，因为创建流不再暴露解析进度页。

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior` + `frontend` + `contract`
- **TDD 策略**: 当前实现已完成。修改 create-flow 逻辑时，先更新 Vitest component / hook / adapter tests，再改组件或 generated-client adapter。
- **BDD 策略**: 适用。当前 BDD gate 为 E2E.P0.081、E2E.P0.083；E2E.P0.082 只保留为 retired/non-current negative，不再作为正向 gate。
- **替代验证 gate**:
  - `corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/create`
  - `corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/ResumeWorkshopScreen.test.tsx src/app/screens/resume-workshop/fixture-parity.test.ts`
  - `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/frontend-resume-workshop/plans/002-create-flow/context.yaml --target frontend`
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
  - `make docs-check`
  - `git diff --check`

## 4 当前交付内容

### Phase 1: Create Flow Shell

`ResumeCreateFlow` owns the input state only. The input state shows upload and paste tabs, preserves route params through auth pending action without raw content, and keeps TopBar / route highlighting stable.

### Phase 2: Upload / Paste Registration

- Upload tab validates file extension and size, requests a presigned upload, uploads the file through browser `fetch`, then calls `registerResume`.
- Paste tab validates non-empty text, derives a meaningful temporary title from the pasted content, and calls `registerResume`.
- Side-effect calls use `Idempotency-Key`; `getResume` polling does not.
- Request headers carry language where required.

### Phase 3: Direct-to-detail Navigation

On successful `registerResume`, upload and paste paths call `navigate({ name: "resume_versions", params: { resumeId } })`. The create flow does not render parser progress, does not poll `getResume`, and does not show a confirm/save page.

### Phase 4: Retired parsing / preview confirm surfaces

`ResumeParseFlow`, `ParsingStage`, `PreviewStage`, `ResumePreviewConfirm`, `useResumeParsingPolling`, `useResumeSave` create-flow usage and `mapParsedSummaryToStructuredProfileDraft` are non-current surfaces. Tests and source negative scans must fail if they are imported by `ResumeCreateFlow` or rendered by the create route.

### Phase 5: CTA / i18n / a11y / Privacy

Home and Workspace CTA paths enter `resume_versions?flow=create`. CreateFlow keeps current i18n keys, tablist semantics, focus behavior, and privacy redlines. The negative tests prevent non-current CreateFlow inputs or prototype runtime imports from returning.

## 5 验收标准

| ID | 场景 | Given | When | Then | 证据 |
|----|------|-------|------|------|------|
| C-1 | Create route | Authenticated user opens `resume_versions?flow=create` | App renders route | `ResumeCreateFlow` appears with upload tab active by default | focused Vitest |
| C-2 | Upload path | Valid file selected | Submit | Presign + PUT + register flow runs with IK and language headers | hook / component tests |
| C-3 | Paste path | Non-empty text | Submit | `registerResume` receives paste payload with content-derived title and direct detail navigation follows | hook / component tests |
| C-4 | Register recovery | Upload/register fails | User retries or returns | Input stays local and content does not leak | upload/paste tests |
| C-5 | Old surfaces absent | Register succeeds | Route updates | Parser animation, preview confirm and create-flow `updateResume` save path do not render or run | negative tests |
| C-6 | CTA handoff | Home or Workspace create CTA | Click | Route lands on current CreateFlow without raw data in pending action | integration tests |
| C-7 | BDD gates | P0.081 / P0.083 assets plus P0.082 retired negative | Scenario verify | Direct-to-detail main path, old parser/confirm absence and CTA handoff are covered | BDD docs + scenario scripts |

## 6 风险与应对

| 风险 | 应对 |
|------|------|
| Generated client drift | Keep fixture parity and real-mode frontend owner tests before changing hook payloads |
| CreateFlow privacy regression | Preserve URL / pendingAction / storage / console tests for raw resume content |
| UI parity drift | Keep CreateFlow DOM anchors, tab roles and screenshot smoke coverage aligned with `ui-design/` |
