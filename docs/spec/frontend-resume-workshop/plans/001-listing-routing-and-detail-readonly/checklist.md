# Frontend Resume Workshop Listing Routing and Detail Readonly Checklist

> **版本**: 2.4
> **状态**: completed
> **更新日期**: 2026-07-08

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
- [x] 3.6 `mapResumeToUiSource` 对上传文件名和与来源 `title` 相同的文件名 `displayName` 做负向过滤，文件名只可作为 `sourceName`，不可成为列表/详情可见简历名称；验证: `adapters/resume.test.ts`。 <!-- verified: 2026-07-07 method=vitest tests=adapters/resume.test.ts -->
- [x] 3.7 `ResumeDetailView` 对 `queued/processing` 且正文快照为空的上传简历轻量轮询 `getResume`，直到 `parsedTextSnapshot` / `originalText` 可展示或进入失败态；不渲染 parser animation / preview-confirm；验证: `ResumeDetailView.test.tsx`、P0.037。 <!-- verified: 2026-07-07 method=vitest+scenario tests=ResumeDetailView.test.tsx,E2E.P0.037 -->
- [x] 3.8 `ResumeDetailView` 对 `failed` 或已有 `parsedTextSnapshot` / `originalText` 的上传详情停止 `getResume` 轮询，避免同一详情 URL 重复请求；验证: `ResumeDetailView.test.tsx` focused regression。<!-- verified: 2026-07-07 method=vitest tests=ResumeDetailView.test.tsx -->

## Phase 4: Privacy / I18n / A11y / Parity

- [x] 4.1 raw resume text、parsed snapshot、structured profile、rewrite text 不进入 URL / pending action / localStorage / console / generic logs；验证: `ResumeWorkshopPrivacy.test.ts`。
- [x] 4.2 中英文案、Accept-Language、只读详情和 aria 语义可测试；验证: `ResumeWorkshopI18nA11y.test.tsx`。
- [x] 4.3 UI parity 锚点、computed style、bounding box、viewport 和 screenshot smoke 由 owner gates 承接；验证: `ResumeWorkshopCssParity.test.ts` 与 pixel parity owner。

## Phase 5: BDD / Negative Gate / Closeout

- [x] 5.1 BDD-Gate: E2E.P0.036 flat list + auth boundary PASS；验证: `test/scenarios/e2e/p0-036-resume-flat-list-auth-boundary/`。
- [x] 5.2 BDD-Gate: E2E.P0.037 read-only source-format detail + removed actions + generic-name negative + 404 fallback PASS；验证: `test/scenarios/e2e/p0-037-resume-detail-preview-readonly/`。<!-- verified: 2026-07-08 method=scenario scenario=E2E.P0.037 -->
- [x] 5.3 001 owner docs、P0.036 slug、scenario INDEX 和 context discovery 不再保留旧树形/版本集合正向语义；验证: targeted grep、context validation、`sync-doc-index --check`、`make docs-check`。

## Phase 6: Resume module UX optimization

- [x] 6.1 `ResumeListView` 删除底部“上传或粘贴另一份简历”CTA，只保留 Header “新建简历”；验证: `ResumeListView.test.tsx` focused negative。<!-- verified: 2026-07-07 method=vitest command="corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/components/ResumeListView.test.tsx src/app/screens/resume-workshop/create/ResumeCreateFlow.test.tsx src/app/screens/resume-workshop/create/UploadTab.test.tsx src/app/screens/resume-workshop/components/ResumePreviewTab.test.tsx src/app/screens/resume-workshop/components/ResumeDetailView.test.tsx src/app/screens/resume-workshop/ResumeWorkshopCssParity.test.ts" -->
- [x] 6.2 `ResumeListView` 每行支持删除简历，调用 `archiveResume` 成功后隐藏该 row，失败时保留 row 并显示错误；验证: `ResumeListView.test.tsx` focused success/failure。<!-- verified: 2026-07-07 method=vitest -->
- [x] 6.3 `ResumeDetailView` 对 `queued/processing` 且无正文的简历渲染等待动画页并轮询，`ready` 后切换 Markdown 详情，`failed` 且无正文时显示失败态；验证: `ResumeDetailView.test.tsx` focused pending/ready/failed。<!-- verified: 2026-07-07 method=vitest -->
- [x] 6.4 `ResumePreviewTab` 将 `parsedTextSnapshot` 按 Markdown 标题、段落、列表渲染，兼容 `originalText` 纯文本 fallback；验证: `ResumePreviewTab.test.tsx` focused Markdown。<!-- verified: 2026-07-07 method=vitest -->
- [x] 6.5 BDD-Gate: P0.036 / P0.037 场景或对应 focused Vitest 覆盖重复 CTA absent、delete action、waiting/detail Markdown 渲染与失败态；验证: scenario trigger/verify 或 owner-approved substitute gate。<!-- verified: 2026-07-07 method=focused-substitute command="corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/components/ResumeListView.test.tsx src/app/screens/resume-workshop/components/ResumeDetailView.test.tsx src/app/screens/resume-workshop/components/ResumePreviewTab.test.tsx" -->

## Phase 7: Markdown engine remediation

- [x] 7.1 `ResumePreviewTab` 使用 `react-markdown` + `remark-gfm` 渲染 `parsedTextSnapshot`，覆盖 heading/list/paragraph 以及 inline `strong` / link，禁止 `**...**`、`[label](url)` 以源码形式露出；验证: `corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/components/ResumePreviewTab.test.tsx` 与 `corepack pnpm --filter @easyinterview/frontend typecheck`。<!-- verified: 2026-07-07 method=vitest+typecheck -->

## Phase 8: Source-format renderer

- [x] 8.1 `ResumePreviewTab` 对 upload-backed PDF 简历使用 source endpoint renderer，source URL 指向 generated client baseUrl 下的 `/resumes/{resumeId}/source`；不得渲染 Markdown fallback、copy/export/original-modal 或新 tab；原 2026-07-07 native object 实现已由 Phase 9 page-stack refinement 取代。<!-- verified: 2026-07-08 method=vitest+typecheck tests=adapters/resume.test.ts,ResumePreviewTab.test.tsx,PdfPageStackPreview.test.tsx,frontend typecheck -->
- [x] 8.2 paste、Markdown upload 和 TXT upload 继续使用 Markdown engine，保持 heading/list/paragraph/inline strong/link DOM 断言；验证: focused frontend tests。<!-- verified: 2026-07-07 method=vitest tests=adapters/resume.test.ts,ResumePreviewTab.test.tsx -->

## Phase 9: PDF page-stack refinement

- [x] 9.1 `ResumePreviewTab` 对 upload-backed PDF 简历渲染从上到下平铺的 PDF 页面栈，source URL 仍指向 generated client baseUrl 下的 `/resumes/{resumeId}/source`；不得渲染 `<object>` / `<iframe>` / `<embed>`、browser PDF viewer toolbar、Markdown fallback、copy/export/original-modal 或新 tab；验证: `ResumePreviewTab.test.tsx` focused red/green。<!-- verified: 2026-07-08 method=red-green-vitest tests=ResumePreviewTab.test.tsx,PdfPageStackPreview.test.tsx -->
- [x] 9.2 PDF 页面栈样式与 UI truth source 同步，desktop/mobile pixel smoke 断言 `resume-detail-pdf-preview-stack` 与 page anchors 可见且没有 native viewer shell；验证: `ResumeWorkshopCssParity.test.ts` + `frontend/tests/pixel-parity/resume-workshop.spec.ts` focused smoke。<!-- verified: 2026-07-08 method=vitest+playwright tests=ResumeWorkshopCssParity.test.ts,pixel-parity/resume-workshop.spec.ts pdf page-stack desktop/mobile -->

## Phase 10: Source-format reading surface alignment

- [x] 10.1 `ResumePreviewTab` Markdown renderer 只渲染 `buildResumeBodyMarkdown(resume)` body，不在 body card 内额外注入 `displayName` / `uiResume.name` / summary / source metadata；验证: `ResumePreviewTab.test.tsx` focused red/green。<!-- verified: 2026-07-08 method=red-green-vitest tests=ResumePreviewTab.test.tsx -->
- [x] 10.2 PDF 与 Markdown renderer 使用统一阅读背景板；Markdown 也渲染白色 page surface，CSS parity 与 pixel smoke 覆盖 Markdown page anchor / PDF page-stack anchor / shared background；验证: `ResumePreviewTab.test.tsx`、`ResumeWorkshopCssParity.test.ts`、`frontend/tests/pixel-parity/resume-workshop.spec.ts` focused smoke。<!-- verified: 2026-07-08 method=vitest+playwright tests=ResumePreviewTab.test.tsx,ResumeWorkshopCssParity.test.ts,pixel-parity/resume-workshop.spec.ts desktop/mobile -->
