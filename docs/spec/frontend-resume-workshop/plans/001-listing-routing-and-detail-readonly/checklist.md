# Frontend Resume Workshop Listing Routing and Detail Readonly Checklist

> **版本**: 1.7
> **状态**: completed
> **更新日期**: 2026-07-07

**关联计划**: [plan](./plan.md)

## Phase 1: Route Shell / Auth Boundary

- [x] 1.1 `ResumeWorkshopScreen` 分派当前 `flow=create|list`、`resumeId`、`targetJobId`、`createMode` route params，忽略旧 `tab` / `tailorRunId`；验证: `ResumeWorkshopScreen.test.tsx`。
- [x] 1.2 未登录态不触发 `listResumes` / `getResume` 或详情二次操作，pending action 只携带安全 route params；验证: `ResumeWorkshopAuthGate.test.tsx` 与 P0.036。

## Phase 2: Flat List View

- [x] 2.1 `ResumeListView` 从 `listResumes` 渲染单层 flat table、创建入口和详情入口；验证: `ResumeListView.test.tsx`、`fixture-parity.test.ts`、P0.036。
- [x] 2.2 loading / empty / retryable error / pagination 状态从 API response 派生；验证: `ResumeListView.test.tsx` 与 `fixture-parity.test.ts`。
- [x] 2.3 打开按钮导航到 `resume_versions?resumeId=<id>`；验证: P0.036 scenario Vitest。

## Phase 3: Read-only Detail

- [x] 3.1 `ResumeDetailView` 使用 `getResume(resumeId)` 渲染 crumb、header 和只读简历正文；旧 `tab=rewrites|edit` 不能 materialize tab；验证: `ResumeDetailView.test.tsx` 与 P0.037。
- [x] 3.2 `ResumePreviewTab` 优先渲染 `parsedTextSnapshot` / `originalText` 原始正文，不因 structured projection 丢失原文信息；不渲染 copy text、original modal、export、rewrite、preview-confirm 或 edit 控件；验证: `ResumePreviewTab.test.tsx`、`ResumeDetailView.test.tsx`、`ResumeDetailExport.test.tsx`、P0.037。
- [x] 3.3 详情页不调用 `exportResume` / `requestResumeTailor` / detail `updateResume`；验证: `ResumeDetailView.test.tsx`、`ResumeDetailExport.test.tsx` 与 P0.037。
- [x] 3.4 不存在的 `resumeId` 渲染 generic NotFoundEmptyState，不回显 fixture `error.code`；验证: `ResumeDetailFixtureParity.test.tsx` 与 P0.037。
- [x] 3.5 `mapResumeToUiSource` 对通用“粘贴的简历 / 上传的简历 / Pasted resume / Uploaded resume”做负向过滤，并只从 LLM-derived `displayName` / structured headline 派生可识别名称；不得把 raw resume 第一行作为名称；验证: `adapters/resume.test.ts`、`ResumeListView.test.tsx`、`ResumeDetailView.test.tsx`。 <!-- verified: 2026-07-07 method=vitest tests=adapters/resume.test.ts,ResumeDetailView.test.tsx,resume-workshop-suite -->
- [x] 3.6 `ResumeDetailView` 对 `queued/processing` 且正文快照为空的上传简历轻量轮询 `getResume`，直到 `parsedTextSnapshot` / `originalText` 可展示或进入失败态；不渲染 parser animation / preview-confirm；验证: `ResumeDetailView.test.tsx`、P0.037。 <!-- verified: 2026-07-07 method=vitest+scenario tests=ResumeDetailView.test.tsx,E2E.P0.037 -->

## Phase 4: Privacy / I18n / A11y / Parity

- [x] 4.1 raw resume text、parsed snapshot、structured profile、rewrite text 不进入 URL / pending action / localStorage / console / generic logs；验证: `ResumeWorkshopPrivacy.test.ts`。
- [x] 4.2 中英文案、Accept-Language、只读详情和 aria 语义可测试；验证: `ResumeWorkshopI18nA11y.test.tsx`。
- [x] 4.3 UI parity 锚点、computed style、bounding box、viewport 和 screenshot smoke 由 owner gates 承接；验证: `ResumeWorkshopCssParity.test.ts` 与 pixel parity owner。

## Phase 5: BDD / Negative Gate / Closeout

- [x] 5.1 BDD-Gate: E2E.P0.036 flat list + auth boundary PASS；验证: `test/scenarios/e2e/p0-036-resume-flat-list-auth-boundary/`。
- [x] 5.2 BDD-Gate: E2E.P0.037 read-only original-content detail + removed actions + generic-name negative + 404 fallback PASS；验证: `test/scenarios/e2e/p0-037-resume-detail-preview-readonly/`。
- [x] 5.3 001 owner docs、P0.036 slug、scenario INDEX 和 context discovery 不再保留旧树形/版本集合正向语义；验证: targeted grep、context validation、`sync-doc-index --check`、`make docs-check`。
