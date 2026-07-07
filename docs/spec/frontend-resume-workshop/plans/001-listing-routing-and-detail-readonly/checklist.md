# Frontend Resume Workshop Listing Routing and Detail Readonly Checklist

> **版本**: 1.4
> **状态**: completed
> **更新日期**: 2026-07-07

**关联计划**: [plan](./plan.md)

## Phase 1: Route Shell / Auth Boundary

- [x] 1.1 `ResumeWorkshopScreen` 分派当前 `flow=create|list`、`resumeId`、`tab`、`targetJobId`、`tailorRunId`、`createMode` route params；验证: `ResumeWorkshopScreen.test.tsx`。
- [x] 1.2 未登录态不触发 `listResumes` / `getResume` / `exportResume`，pending action 只携带安全 route params；验证: `ResumeWorkshopAuthGate.test.tsx` 与 P0.036。

## Phase 2: Flat List View

- [x] 2.1 `ResumeListView` 从 `listResumes` 渲染单层 flat table、创建入口和详情入口；验证: `ResumeListView.test.tsx`、`fixture-parity.test.ts`、P0.036。
- [x] 2.2 loading / empty / retryable error / pagination 状态从 API response 派生；验证: `ResumeListView.test.tsx` 与 `fixture-parity.test.ts`。
- [x] 2.3 打开按钮导航到 `resume_versions?resumeId=<id>&tab=preview`；验证: P0.036 scenario Vitest。

## Phase 3: Detail Preview / Original / Export

- [x] 3.1 `ResumeDetailView` 使用 `getResume(resumeId)` 渲染 crumb、header、preview / rewrites / edit tablist，并保留显式 `tab`；验证: `ResumeDetailView.test.tsx` 与 P0.037。
- [x] 3.2 Preview tab 渲染 structured projection、copy text 和 original modal；验证: `ResumePreviewTab.test.tsx`、`OriginalResumePreviewModal.test.tsx`、P0.037。
- [x] 3.3 `exportResume` 使用 `Idempotency-Key` 并把 P0 unavailable response 映射为 toast，不写 blob / localStorage；验证: `ResumeDetailExport.test.tsx` 与 P0.037。
- [x] 3.4 不存在的 `resumeId` 渲染 generic NotFoundEmptyState，不回显 fixture `error.code`；验证: `ResumeDetailFixtureParity.test.tsx` 与 P0.037。

## Phase 4: Privacy / I18n / A11y / Parity

- [x] 4.1 raw resume text、parsed snapshot、structured profile、rewrite text 不进入 URL / pending action / localStorage / console / generic logs；验证: `ResumeWorkshopPrivacy.test.ts`。
- [x] 4.2 中英文案、Accept-Language、tablist、modal focus 和 aria 语义可测试；验证: `ResumeWorkshopI18nA11y.test.tsx`。
- [x] 4.3 UI parity 锚点、computed style、bounding box、viewport 和 screenshot smoke 由 owner gates 承接；验证: `ResumeWorkshopCssParity.test.ts` 与 pixel parity owner。

## Phase 5: BDD / Negative Gate / Closeout

- [x] 5.1 BDD-Gate: E2E.P0.036 flat list + auth boundary PASS；验证: `test/scenarios/e2e/p0-036-resume-flat-list-auth-boundary/`。
- [x] 5.2 BDD-Gate: E2E.P0.037 detail preview + original modal + export 501 + 404 fallback PASS；验证: `test/scenarios/e2e/p0-037-resume-detail-preview-readonly/`。
- [x] 5.3 001 owner docs、P0.036 slug、scenario INDEX 和 context discovery 不再保留旧树形/版本集合正向语义；验证: targeted grep、context validation、`sync-doc-index --check`、`make docs-check`。
