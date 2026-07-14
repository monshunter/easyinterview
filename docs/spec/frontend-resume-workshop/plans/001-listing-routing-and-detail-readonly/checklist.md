# Frontend Resume Workshop Listing Routing and Detail Readonly Checklist

> **版本**: 3.7
> **状态**: completed
> **更新日期**: 2026-07-14

**关联计划**: [plan](./plan.md)

## Phase 1: Route Shell / Auth Boundary

- [x] 1.1 `ResumeWorkshopScreen` 分派当前 `flow=create|list`、`resumeId`、`targetJobId`、`createMode` route params，忽略 out-of-scope `tab` / `tailorRunId`；验证: `ResumeWorkshopScreen.test.tsx`。
- [x] 1.2 未登录态不触发 `listResumes` / `getResume` 或详情二次操作，pending action 只携带安全 route params；验证: `ResumeWorkshopAuthGate.test.tsx` 与 P0.036。

## Phase 2: Flat List View

- [x] 2.1 `ResumeListView` 从 `listResumes` 渲染单层 flat table、创建入口和详情入口；验证: `ResumeListView.test.tsx`、`fixture-parity.test.ts`、P0.036。
- [x] 2.2 loading / empty / retryable error / pagination 状态从 API response 派生；验证: `ResumeListView.test.tsx` 与 `fixture-parity.test.ts`。
- [x] 2.3 打开按钮导航到 `resume_versions?resumeId=<id>`；验证: P0.036 scenario Vitest。

## Phase 3: Read-only Detail

- [x] 3.1 `ResumeDetailView` 使用 `getResume(resumeId)` 渲染 crumb、header 和只读简历正文；out-of-scope `tab=rewrites|edit` 不能 materialize tab；验证: `ResumeDetailView.test.tsx` 与 P0.037。
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

## Phase 11: P0.036 test lifecycle isolation

- [x] 11.1 P0.036 out-of-scope 同步负向用例在断言后显式 unmount，清除无关 runtime/interview provider updates（验证：P0.036 focused 无 act warning、Resume Workshop owner/full frontend tests、build、owner context/docs gates）
  <!-- verified: 2026-07-10 method=resume-workshop-test-lifecycle-isolation evidence="Focused red preserved two AppRuntimeProvider and one InterviewContextProvider act warnings while all 4 assertions passed. Added explicit unmount to the synchronous out-of-scope test without changing assertions or production code. P0.036 4 tests and Resume Workshop owner 21 files/115 tests pass warning-free; frontend build and owner/product contexts pass. Full frontend 137 files/829 tests pass with zero React update warnings; completed-state docs/diff/pruning gates rerun during closeout." -->

## Phase 12: PDF.js on-demand loading

- [x] 12.1 `PdfPageStackPreview` 在现有 loading shell 后动态导入 PDF.js/runtime worker，保持 credential、cancel/error/page-stack 合同（验证：focused red/green、Resume Workshop/full frontend tests、build+sourcemap 主 chunk/PDF chunk byte delta、owner context/docs gates）
  <!-- verified: 2026-07-10 method=pdfjs-on-demand-loading evidence="Focused red proved getDocument ran during synchronous render. Moved PDF.js and worker URL runtime imports into the existing effect while retaining type-only contracts and cancellation paths. Focused component 2, Resume Workshop owner 115 and full frontend 830 tests pass warning-free; typecheck/build and owner/product contexts pass. Sourcemap build moved pdfjs-dist source count in main 1->0, main JS 1,161,279->667,768 bytes and emitted a 495,319-byte PDF runtime chunk; completed-state docs/diff/pruning gates rerun during closeout." -->

## Phase 13: P0.037 async test lifecycle

- [x] 13.1 Capture P0.037 stderr in scenario evidence and add a verify gate that fails on an unwrapped React update warning; confirm the current failed-PDF wait is RED.
  <!-- verified: 2026-07-10 method=p0-037-react-update-warning-red evidence="A direct focused run emitted PdfPageStackPreview 'not wrapped in act' while all 6 tests passed. trigger.sh now captures stderr and verify.sh rejects that marker; a subsequent process was race-clean, confirming the warning is timing-sensitive rather than assertion-deterministic." -->
- [x] 13.2 Wrap the failed-PDF 350ms no-poll observation windows in P0.037 and its `ResumeDetailView` owner mirror with Testing Library `act`, without changing production code or business assertions.
  <!-- verified: 2026-07-10 method=resume-pdf-observation-wait-green evidence="Both duplicate raw 350ms waits now run inside Testing Library act. Focused P0.037 plus ResumeDetailView pass 2 files/14 tests warning-free; production PdfPageStackPreview and all request-count/content assertions are unchanged." -->
- [x] 13.3 Run focused P0.037, its four-stage wrapper, Resume Workshop regressions, owner/product contexts, docs, diff and pruning gates; then restore the owner to `completed`.
  <!-- verified: 2026-07-10 method=p0-037-async-test-lifecycle evidence="Focused duplicate tests pass 2 files/14 tests; Resume Workshop plus P0.037 passes 20/111; full frontend passes 138/839 and complete stderr has zero React update warning. Typecheck, final four-stage P0.037 wrapper, owner/product contexts and docs/diff/pruning gates pass with real_residuals=0." -->

## Phase 14: orphan Resume toast bridge removal

- [x] 14.1 Add a scoped source-surface RED gate proving the formal Resume Workshop helper and P0.036 still contain the unconsumed prototype toast bridge.
  <!-- verified: 2026-07-10 method=orphan-resume-toast-source-red evidence="Focused ResumeWorkshopPrivacy failed exactly on components/toast.ts and p0-036-resume-flat-list-auth-boundary.test.tsx, with no third offender." -->
- [x] 14.2 Delete `components/toast.ts` and P0.036 toast capture/assertion scaffolding; keep the `ui-design/` prototype unchanged.
  <!-- verified: 2026-07-10 method=orphan-resume-toast-removal evidence="Deleted the entire unimported helper and removed P0.036's self-only global capture plus constant-false assertion. Scoped literal search is empty; privacy/P0.036 focused tests pass 2 files/9 tests. ui-design source is unchanged." -->
- [x] 14.3 Run focused Resume Workshop/P0.036 tests, the P0.036 wrapper, owner/product contexts, docs, diff and pruning gates; then restore the owner to `completed`.
  <!-- verified: 2026-07-10 method=orphan-resume-toast-removal evidence="Scoped zero-reference gate plus P0.036 pass 9/9; P0.036 setup/trigger/verify/cleanup passes 4/4; Resume Workshop plus P0.036 passes 20 files/110 tests and typecheck. Owner/product contexts and docs/index/link/diff/pruning gates pass with real_residuals=0." -->

## Phase 15: P0.037 ready DOM synchronization

- [x] 15.1 Capture the full-suite RED where the second `getResume` call is observed before the ready heading commits, and identify the same request-count-only wait in both P0.037 and its owner mirror.
  <!-- verified: 2026-07-10 method=p0-037-ready-dom-race-red evidence="A full frontend run under concurrent build load failed with 840/841 because the scenario observed getResume call 2 but synchronously found only Loading resume. Focused P0.037 then passed 6/6, proving timing sensitivity; source inspection found the same waitFor(call count 2) followed by synchronous getAllByRole in both scenario and ResumeDetailView tests." -->
- [x] 15.2 Synchronize both tests on the ready `displayName` heading, then assert the exact two-call polling contract and unchanged PDF page-stack/content negatives.
  <!-- verified: 2026-07-10 method=p0-037-ready-dom-race-green evidence="Both duplicate tests now wait for the ready displayName heading and only then assert exactly two getResume calls. Focused scenario plus owner mirror pass 2 files/14 tests; the request-count-first source pattern is zero and all page-stack URL/content/native-viewer negatives remain unchanged." -->
- [x] 15.3 Run repeated focused tests, the P0.037 four-stage wrapper, Resume Workshop/full frontend regressions, owner/product contexts and docs/diff/pruning gates; then restore the owner to `completed`.
  <!-- verified: 2026-07-10 method=p0-037-ready-dom-race-regression evidence="Four concurrent focused processes each pass 2 files/14 tests. P0.037 setup/trigger/verify/cleanup passes 6/6; Resume Workshop plus scenario passes 20 files/113 tests, typecheck passes, and full frontend passes 137 files/841 tests. Both owner contexts validate, BUG-0153 records the diagnosis, and final docs/diff/pruning gates run during closeout. No environment restart or data cleanup occurred." -->

## Phase 16: zero-consumer detail CSS pruning

- [x] 16.1 Add a source-level RED gate that names every zero-consumer breadcrumb, structured-preview and original-modal selector still present in `screens.css`.
  <!-- verified: 2026-07-10 method=resume-css-source-red evidence="Focused ResumeWorkshopCssParity ran 4 tests: all 3 existing parity contracts passed and only the new zero-consumer selector contract failed, first reporting .ei-resume-detail-breadcrumb still present." -->
- [x] 16.2 Delete those selectors and the dead shared-button/mobile branches without adding aliases, placeholders or removal markers; preserve all current list/detail selectors.
  <!-- verified: 2026-07-10 method=resume-css-source-green evidence="ResumeWorkshopCssParity passes 4/4. Target selectors are absent from formal CSS/DOM/prototype sources and remain only as negative test literals; current back/header/Markdown/PDF/parse-state selectors retain production consumers." -->
- [x] 16.3 Run focused Resume Workshop CSS/component tests, full Resume Workshop owner tests, typecheck/build, owner/product contexts and docs/index/diff/pruning gates; then restore the owner to `completed`.
  <!-- verified: 2026-07-10 method=resume-detail-zero-consumer-css-pruning evidence="CSS parity passes 4/4; Resume Workshop owner passes 19 files/108 tests; full frontend passes 136 files/837 tests; typecheck and production build pass. Target CSS selectors are absent outside the negative gate, live detail selectors retain production consumers, and both owner/product contexts validate. Final docs/index/diff/pruning gates run during closeout. No environment restart or data cleanup occurred." -->

## Phase 17: detail CSS cascade consolidation

- [x] 17.1 Add a source RED gate requiring one detail-back rule with the complete effective declarations and no flex-preview grid declaration.
  <!-- verified: 2026-07-10 method=resume-detail-css-cascade-red evidence="Focused ResumeWorkshopCssParity ran 5 tests: all 4 existing contracts passed and only the new cascade gate failed because two detail-back rules remained." -->
- [x] 17.2 Merge the effective declarations into one rule and delete the ineffective media block without changing selectors, DOM or computed values.
  <!-- verified: 2026-07-10 method=resume-detail-css-cascade-green evidence="CSS parity passes 5/5. Exactly one detail-back rule retains the effective inline-flex/padding/color/background/border/radius/font declarations, and the flex preview no longer carries grid-template-columns; all remaining grid declarations belong to grid layouts." -->
- [x] 17.3 Run focused/full Resume Workshop, full frontend, typecheck/build, owner/product contexts and docs/index/diff/pruning gates; then restore the owner to `completed`.
  <!-- verified: 2026-07-10 method=resume-detail-css-cascade-consolidation evidence="CSS parity passes 5; Resume Workshop passes 19 files/110 tests; full frontend passes 136 files/843 tests; typecheck/build and both contexts pass. Detail-back has one complete effective rule, the flex-preview grid declaration is absent, and final docs/index/diff/pruning gates run during closeout. No Bug/retrospective report, environment restart or data cleanup was needed." -->

## Phase 18: empty pending-decision section removal

- [x] 18.1 Record a scoped RED gate proving the active spec still contains an empty pending-decision section.
  <!-- verified: 2026-07-10 method=frontend-resume-empty-pending-red evidence="The scoped absence gate failed first on the active spec's combined decisions/pending heading; source inventory confirmed the only 3.2 content was a no-pending status sentence and no active anchor referenced it." -->
- [x] 18.2 Delete the empty section, synchronize spec/history/contexts/indexes, and run owner/product context plus docs/index/link/diff/pruning gates; then restore the owner to `completed`.
  <!-- verified: 2026-07-10 method=frontend-resume-empty-pending-section-removal evidence="Spec v2.14 keeps only the current decisions heading and has no empty pending section or no-pending status sentence. History, both owner contexts and the top-level spec index are synchronized; 001/002/product contexts, links and final docs/index/diff/pruning gates pass. No code, UI, BDD, Bug/retrospective report, environment restart or data cleanup was needed." -->

## Phase 19: Resume summary consumption and idempotent initial reads

- [x] 19.1 RED：generated client / fixture parity / list hook / `ResumeListView` / Home selector tests 证明当前列表仍可见完整详情字段；新增 exact-key 断言，只允许 `id,title,displayName,language,sourceType,parseStatus,summaryHeadline,hasReadableContent,updatedAt`，并对 `originalText|parsedTextSnapshot|structuredProfile|fileObjectId|parsedSummary|createdAt|deletedAt` 建立编译期或运行时负向 gate。
- [x] 19.2 GREEN：在 B2 generated `ResumeSummary` 与 `PaginatedResume.items: ResumeSummary[]` 就位后，列表与 Home selector 只消费 summary；不得新增 pagination wrapper；`hasReadableContent` / `summaryHeadline` 取代从正文详情推断；`ResumeDetailView` 继续仅由 `getResume` 消费 full `Resume`。验证：adapter/hook/component/fixture parity focused Vitest + typecheck。
- [x] 19.3 RED/GREEN：StrictMode harness 同时覆盖 `listResumes` 与 ready `getResume`，以底层 `fetch`/transport spy 证明相同 no-signal request identity 当前产生重复 transport，再修复为恰好 1 次；不得删除 StrictMode 或只改测试 method mock。
- [x] 19.4 reject/retry/abort/polling：第一次相同请求 reject 后 registry 清空，用户 retry 发起新的 transport 并成功；resolve 后新用户动作也可发起新 transport；带 `AbortSignal` 请求不共享；queued/processing 详情仅在前次 settle 后轮询，ready/failed/已有正文不轮询。验证：focused hook/client/component tests。
- [x] 19.5 `BDD-Gate: E2E.P0.036` PASS：closed ResumeSummary list、forbidden detail fields absent、StrictMode 单次实际 transport、失败后 retry 新 transport。
- [x] 19.6 `BDD-Gate: E2E.P0.037` PASS：full Resume 只由详情读取、ready 初始单次实际 transport、pending 轮询串行、失败后可重试。
- [x] 19.7 收口：frontend focused/full tests、typecheck/build、owner contexts、`sync-doc-index --check`、`make docs-check`、`git diff --check` 与 pruning gate PASS；完成后同步 checklist 证据并恢复 completed。
  <!-- verified: 2026-07-14 evidence="P0.036 and P0.037 fresh setup/trigger/verify/cleanup PASS with exact summary keys, forbidden-detail negatives, StrictMode list/detail transport=1, retry=2 and serial polling. Full frontend 125 files / 1004 tests plus typecheck/build PASS." -->
